package native

import (
	"context"
	"os"
	"path/filepath"

	"github.com/docker/go-connections/sockets"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	pb "github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	pbrpc "github.com/xuperchain/xuperchain/core/contractsdk/go/pbrpc"
	"google.golang.org/grpc"
)

type nativeCreator struct {
	config    *bridge.InstanceCreatorConfig
	chainsock string
	pm        *processManager
}

func newNativeCreator(cfg *bridge.InstanceCreatorConfig) (bridge.InstanceCreator, error) {
	creator := &nativeCreator{
		config:    cfg,
		chainsock: filepath.Join(cfg.Basedir, "xuper.sock"),
	}
	err := os.MkdirAll(cfg.Basedir, 0755)
	if err != nil {
		return nil, err
	}

	pm, err := newProcessManager(cfg.VMConfig.(*config.NativeConfig), cfg.Basedir, creator.chainsock)
	if err != nil {
		return nil, err
	}
	creator.pm = pm

	err = creator.startRpcServer(cfg.SyscallService)
	if err != nil {
		return nil, err
	}
	return creator, nil
}

func (n *nativeCreator) startRpcServer(service *bridge.SyscallService) error {
	uid, gid := os.Getuid(), os.Getgid()
	relpath := NormalizeSockPath(n.chainsock)

	listener, err := sockets.NewUnixSocketWithOpts(relpath, sockets.WithChown(uid, gid), sockets.WithChmod(0660))
	if err != nil {
		return err
	}
	rpcServer := grpc.NewServer()
	pbrpc.RegisterSyscallServer(rpcServer, service)
	go rpcServer.Serve(listener)
	return nil
}

func (n *nativeCreator) CreateInstance(ctx *bridge.Context, cp bridge.ContractCodeProvider) (bridge.Instance, error) {
	process, err := n.pm.GetProcess(ctx.ContractName, cp)
	if err != nil {
		return nil, err
	}
	return newNativeVmInstance(ctx, process), nil
}

func (n *nativeCreator) RemoveCache(name string) {

}

type nativeVmInstance struct {
	ctx     *bridge.Context
	process *contractProcess
}

func newNativeVmInstance(ctx *bridge.Context, process *contractProcess) *nativeVmInstance {
	return &nativeVmInstance{
		ctx:     ctx,
		process: process,
	}
}

func (i *nativeVmInstance) Exec() error {
	request := &pb.NativeCallRequest{
		Ctxid: i.ctx.ID,
	}
	_, err := i.process.RpcClient().Call(context.TODO(), request)
	return err
}

func (i *nativeVmInstance) ResourceUsed() contract.Limits {
	return contract.Limits{
		XFee: 1,
	}
}

func (i *nativeVmInstance) Release() {

}

func (i *nativeVmInstance) Abort(msg string) {
}

func init() {
	bridge.Register(bridge.TypeNative, "native", newNativeCreator)
}
