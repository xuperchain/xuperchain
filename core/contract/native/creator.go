package native

import (
	"context"
	"net"
	"os"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	pb "github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	pbrpc "github.com/xuperchain/xuperchain/core/contractsdk/go/pbrpc"
	"google.golang.org/grpc"
)

type nativeCreator struct {
	config   *bridge.InstanceCreatorConfig
	listener net.Listener
	pm       *processManager
}

func newNativeCreator(cfg *bridge.InstanceCreatorConfig) (bridge.InstanceCreator, error) {
	creator := &nativeCreator{
		config: cfg,
	}
	err := os.MkdirAll(cfg.Basedir, 0755)
	if err != nil {
		return nil, err
	}

	listenAddr, err := creator.startRpcServer(cfg.SyscallService)
	if err != nil {
		return nil, err
	}

	pm, err := newProcessManager(cfg.VMConfig.(*config.NativeConfig), cfg.Basedir, listenAddr)
	if err != nil {
		return nil, err
	}
	creator.pm = pm

	return creator, nil
}

func (n *nativeCreator) startRpcServer(service *bridge.SyscallService) (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	n.listener = listener
	rpcServer := grpc.NewServer()
	pbrpc.RegisterSyscallServer(rpcServer, service)
	go rpcServer.Serve(listener)

	addr := "tcp://" + listener.Addr().String()
	return addr, nil
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
