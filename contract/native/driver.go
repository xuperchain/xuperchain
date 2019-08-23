package native

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/client"
	log "github.com/xuperchain/log15"
	"google.golang.org/grpc"

	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
	pbrpc "github.com/xuperchain/xuperunion/contractsdk/go/pbrpc"
	"github.com/xuperchain/xuperunion/crypto/hash"
	xpb "github.com/xuperchain/xuperunion/pb"
)

type nativeCodeStatus int

const (
	statusRegistered nativeCodeStatus = iota
	statusReady
	statusInvalid
)

type standardNativeContract struct {
	name    string
	version string
	status  nativeCodeStatus
	// base directory of driver.
	// eg ${XCHAIN_ROOT}/data/blockchain/xuper/native/driver/math
	basedir string
	bindir  string
	binpath string
	//	dbdir         string
	sockpath      string
	chainSockPath string

	lostBeatheart bool
	mutex         *sync.Mutex
	rpcClient     pbrpc.NativeCodeClient
	desc          *xpb.NativeCodeDesc

	process       Process
	monitorStopch chan struct{} //退出监控器
	monitorWaiter sync.WaitGroup

	mgr  *GeneralSCFramework
	vsnc *versionedStandardNativeContract
	log.Logger
}

func (snc *standardNativeContract) Init() error {
	// 刚启动的时候SetContext没有调用，QueryContract会得到空指针
	snc.bindir = filepath.Join(snc.basedir, "bin")
	snc.binpath = filepath.Join(snc.bindir, "nativecode")
	snc.sockpath = filepath.Join(snc.basedir, path.Base(snc.name)+".sock")
	relsockpath, err := RelPathOfCWD(snc.sockpath)
	if err != nil {
		return err
	}
	conn, err := grpc.Dial("unix:"+relsockpath, grpc.WithInsecure())
	if err != nil {
		return err
	}
	snc.rpcClient = pbrpc.NewNativeCodeClient(conn)
	return nil
}

func (snc *standardNativeContract) heartBeat() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	_, err := snc.rpcClient.Ping(ctx, new(pb.PingRequest))
	return err
}

// wait the subprocess to be ready
func (snc *standardNativeContract) waitReply() error {
	const waitTimeout = 2 * time.Second
	ctx, cancel := context.WithTimeout(context.TODO(), waitTimeout)
	defer cancel()
	for {
		_, err := snc.rpcClient.Ping(ctx, new(pb.PingRequest))
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

// monitor subprocess by consuming reply msg
func (snc *standardNativeContract) monitor() {
	snc.monitorWaiter.Add(1)
	defer snc.monitorWaiter.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-snc.monitorStopch:
			return
		case <-ticker.C:
			err := snc.heartBeat()
			if err != nil {
				//subprocess died
				snc.lostBeatheart = true
				snc.Warn("monitor: lostBeatheart", "error", err)
				return
			}
		}
	}
}

func (snc *standardNativeContract) GetNativeCodeDigest() ([]byte, error) {
	content, err := ioutil.ReadFile(snc.binpath)
	if err != nil {
		return nil, err
	}
	return hash.DoubleSha256(content), nil
}

func (snc *standardNativeContract) newProcess() Process {
	if snc.mgr.cfg.Docker.Enable {
		client, _ := client.NewEnvClient()
		return &DockerProcess{
			basedir:       snc.basedir,
			binpath:       snc.binpath,
			sockpath:      snc.sockpath,
			chainSockPath: snc.chainSockPath,
			cfg:           &snc.mgr.cfg.Docker,
			client:        client,
			Logger:        snc.Logger.New("module", "DockerProcess"),
		}
	}
	return &HostProcess{
		basedir:       snc.basedir,
		binpath:       snc.binpath,
		sockpath:      snc.sockpath,
		chainSockPath: snc.chainSockPath,
		Logger:        snc.Logger.New("module", "HostProcess"),
	}
}

func (snc *standardNativeContract) Start() error {
	snc.mutex.Lock()
	defer snc.mutex.Unlock()
	process := snc.newProcess()
	err := process.Start()
	if err != nil {
		return err
	}
	err = snc.waitReply()
	if err != nil {
		// 避免启动失败后产生僵尸进程
		process.Stop(time.Duration(snc.mgr.cfg.StopTimeout) * time.Second)
		return err
	}
	snc.process = process
	snc.Info("start process success")
	snc.monitorStopch = make(chan struct{})
	go snc.monitor()
	snc.lostBeatheart = false
	return nil
}

func (snc *standardNativeContract) Restart() error {
	snc.Stop()
	err := snc.Start()
	if err != nil {
		return err
	}
	return nil
}

func (snc *standardNativeContract) Stop() {
	//quit contract process gracefully
	snc.mutex.Lock()
	defer snc.mutex.Unlock()
	//先保证不会被重启
	snc.lostBeatheart = true
	if snc.process == nil {
		snc.Info("deactivate snc stop, process is nil")
		return
	}
	close(snc.monitorStopch)

	err := snc.process.Stop(time.Duration(snc.mgr.cfg.StopTimeout) * time.Second)
	if err != nil {
		snc.Info("process stoped with error", "error", err)
	}
	// snc.Info("process done", "pid", snc.cmd.Process.Pid)
	snc.monitorWaiter.Wait()
	snc.process = nil
}
