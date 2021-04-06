package native

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	log15 "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/pbrpc"
	xpb "github.com/xuperchain/xuperchain/core/pb"

	"google.golang.org/grpc"
)

type contractProcess struct {
	cfg *config.NativeConfig

	name      string
	basedir   string
	binpath   string
	chainAddr string
	desc      *xpb.WasmCodeDesc

	process       Process
	monitorStopch chan struct{}
	monitorWaiter sync.WaitGroup
	logger        log15.Logger

	mutex     sync.Mutex
	rpcPort   int
	rpcConn   *grpc.ClientConn
	rpcClient pbrpc.NativeCodeClient
}

func newContractProcess(cfg *config.NativeConfig, name, basedir, chainAddr string, desc *xpb.WasmCodeDesc) (*contractProcess, error) {
	process := &contractProcess{
		cfg:           cfg,
		name:          name,
		basedir:       basedir,
		binpath:       filepath.Join(basedir, nativeCodeFileName(desc)),
		chainAddr:     chainAddr,
		desc:          desc,
		monitorStopch: make(chan struct{}),
		logger:        log.DefaultLogger.New("contract", name),
	}
	return process, nil
}

func (c *contractProcess) makeHostProcess() (Process, error) {
	envs := []string{
		"XCHAIN_CODE_PORT=" + strconv.Itoa(c.rpcPort),
		"XCHAIN_CHAIN_ADDR=" + c.chainAddr,
	}
	startcmd, err := c.makeStartCommand()
	if err != nil {
		return nil, err
	}
	if !c.cfg.Docker.Enable {
		return &HostProcess{
			basedir:  c.basedir,
			startcmd: startcmd,
			envs:     envs,
			Logger:   c.logger,
		}, nil
	}
	mounts := []string{
		c.basedir,
	}
	return &DockerProcess{
		basedir:  c.basedir,
		startcmd: startcmd,
		envs:     envs,
		mounts:   mounts,
		// ports:    []string{strconv.Itoa(c.rpcPort)},
		cfg:    &c.cfg.Docker,
		Logger: c.logger,
	}, nil
}

// wait the subprocess to be ready
func (c *contractProcess) waitReply() error {
	const waitTimeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.TODO(), waitTimeout)
	defer cancel()
	for {
		_, err := c.rpcClient.Ping(ctx, new(pb.PingRequest))
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting native code start timeout. error:%s", err)
		default:
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (c *contractProcess) heartBeat() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	_, err := c.rpcClient.Ping(ctx, new(pb.PingRequest))
	return err
}

func (c *contractProcess) monitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	c.monitorWaiter.Add(1)
	defer c.monitorWaiter.Done()
forloop:
	for {
		select {
		case <-c.monitorStopch:
			return
		case <-ticker.C:
			err := c.heartBeat()
			if err == nil {
				continue forloop
			}
			c.logger.Error("process heartbeat error", "error", err)
			err = c.restartProcess()
			if err != nil {
				c.logger.Error("restart process error", "error", err)
			}
		}
	}
}

func (c *contractProcess) resetRpcClient() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.rpcConn != nil {
		c.rpcConn.Close()
	}
	port, err := makeFreePort()
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.rpcPort = port
	c.rpcConn = conn
	c.rpcClient = pbrpc.NewNativeCodeClient(c.rpcConn)
	return nil
}

func (c *contractProcess) RpcClient() pbrpc.NativeCodeClient {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.rpcClient
}

func (c *contractProcess) restartProcess() error {
	c.process.Stop(time.Second)
	return c.start(false)
}

func (c *contractProcess) start(startMonitor bool) error {
	err := c.resetRpcClient()
	if err != nil {
		return err
	}
	c.process, err = c.makeHostProcess()
	if err != nil {
		return err
	}

	err = c.process.Start()
	if err != nil {
		return err
	}
	err = c.waitReply()
	if err != nil {
		// 避免启动失败后产生僵尸进程
		c.process.Stop(time.Second)
		return err
	}
	if startMonitor {
		go c.monitor()
	}

	return nil
}

func (c *contractProcess) Start() error {
	return c.start(true)
}

func (c *contractProcess) Stop() {
	// close monitor and waiting monitor stoped
	close(c.monitorStopch)
	c.monitorWaiter.Wait()

	err := c.process.Stop(time.Second)
	if err != nil {
		c.logger.Error("process stoped error", "error", err)
	}
}

func (c *contractProcess) GetDesc() *xpb.WasmCodeDesc {
	return c.desc
}

func (c *contractProcess) makeStartCommand() (string, error) {
	switch c.desc.GetRuntime() {
	case "java":
		return "java -jar " + c.binpath, nil
	case "go":
		return c.binpath, nil
	case "py":
		//TODO @fengjin
		// only support python3 as python2 is meeting its EOL in
		return "python3 " + c.binpath, nil
	default:
		return "", fmt.Errorf("unsupported native contract runtime %s", c.desc.GetRuntime())
	}
}

func makeFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	addr := l.Addr().(*net.TCPAddr)
	l.Close()
	return addr.Port, nil
	//return 9999, nil
}
