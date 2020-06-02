package native

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	log15 "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"
	pb "github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	pbrpc "github.com/xuperchain/xuperchain/core/contractsdk/go/pbrpc"
	xpb "github.com/xuperchain/xuperchain/core/pb"

	"google.golang.org/grpc"
)

type contractProcess struct {
	cfg *config.NativeConfig

	name       string
	basedir    string
	binpath    string
	chainsock  string
	nativesock string
	desc       *xpb.WasmCodeDesc

	process       Process
	rpcClient     pbrpc.NativeCodeClient
	monitorStopch chan struct{}
	monitorWaiter sync.WaitGroup
	logger        log15.Logger
}

func newContractProcess(cfg *config.NativeConfig, name, basedir, chainsock string, desc *xpb.WasmCodeDesc) (*contractProcess, error) {
	process := &contractProcess{
		cfg:        cfg,
		name:       name,
		basedir:    basedir,
		binpath:    filepath.Join(basedir, nativeCodeFileName(desc)),
		chainsock:  chainsock,
		nativesock: filepath.Join(basedir, nativeCodeSockFileName(desc)),
		desc:       desc,
		logger:     log.DefaultLogger.New("contract", name),
	}

	process.process = process.makeHostProcess()

	relsockpath := NormalizeSockPath(process.nativesock)
	conn, err := grpc.Dial("unix:"+relsockpath, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	process.rpcClient = pbrpc.NewNativeCodeClient(conn)

	return process, nil
}

func (p *contractProcess) makeHostProcess() Process {
	envs := []string{
		"XCHAIN_CODE_SOCK=" + p.nativesock,
		"XCHAIN_CHAIN_SOCK=" + p.chainsock,
	}
	if !p.cfg.Docker.Enable {
		return &HostProcess{
			basedir:  p.basedir,
			startcmd: p.binpath,
			envs:     envs,
			Logger:   p.logger,
		}
	}
	mounts := []string{
		p.basedir, p.chainsock,
	}
	return &DockerProcess{
		basedir:  p.basedir,
		startcmd: p.binpath,
		envs:     envs,
		mounts:   mounts,
		cfg:      &p.cfg.Docker,
		Logger:   p.logger,
	}
}

// wait the subprocess to be ready
func (p *contractProcess) waitReply() error {
	const waitTimeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.TODO(), waitTimeout)
	defer cancel()
	for {
		_, err := p.rpcClient.Ping(ctx, new(pb.PingRequest))
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

func (c *contractProcess) RpcClient() pbrpc.NativeCodeClient {
	return c.rpcClient
}

func (c *contractProcess) restartProcess() error {
	c.process.Stop(time.Second)
	return c.start(false)
}

func (c *contractProcess) start(startMonitor bool) error {
	err := c.process.Start()
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
