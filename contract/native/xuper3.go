package native

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/go-connections/sockets"
	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/bridge"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
	pbrpc "github.com/xuperchain/xuperunion/contractsdk/go/pbrpc"
	"google.golang.org/grpc"
)

// RegisterSyscallService implements bridge.Executor
func (gscf *GeneralSCFramework) RegisterSyscallService(service *bridge.SyscallService) {
	rpcServer := grpc.NewServer()
	pbrpc.RegisterSyscallServer(rpcServer, service)
	uid, gid := os.Getuid(), os.Getgid()

	relpath, err := RelPathOfCWD(gscf.chainSock3Path)
	if err != nil {
		panic(err)
	}
	listener, err := sockets.NewUnixSocketWithOpts(relpath, sockets.WithChown(uid, gid), sockets.WithChmod(0660))
	if err != nil {
		gscf.Error("NewUnixSocketWithOpts error", "error", err, "chainSockPath", gscf.chainSock3Path)
		panic(err)
	}
	go rpcServer.Serve(listener)
}

// NewInstance implements bridge.Executor
func (gscf *GeneralSCFramework) NewInstance(ctx *bridge.Context) (bridge.Instance, error) {
	vsnc := gscf.getVSNC(ctx.ContractName)
	if vsnc == nil {
		return nil, fmt.Errorf("contract %s not found", ctx.ContractName)
	}
	snc, err := vsnc.GetSNC(ctx.ContractName, vsnc.curVersion)
	if err != nil {
		return nil, err
	}
	return &nativeInstance{
		ctx: ctx,
		snc: snc,
	}, nil
}

type nativeInstance struct {
	ctx *bridge.Context
	snc *standardNativeContract
}

func (n *nativeInstance) ResourceUsed() contract.Limits {
	return contract.Limits{
		XFee: 1,
	}
}

func (n *nativeInstance) Release() {
}

func (n *nativeInstance) Exec() error {
	snc := n.snc
	switch snc.status {
	case statusRegistered:
		return fmt.Errorf("this driver isn't ready, name=%s", n.ctx.ContractName)
	case statusReady:
		if snc.lostBeatheart {
			snc.Info("callNativeCode error, retrying", "snc", fmt.Sprintf("%#v", snc))
			return common.ErrContractConnectionError
		}
	default:
		return fmt.Errorf("unknown status:%d", snc.status)
	}

	request := &pb.NativeCallRequest{
		Ctxid: n.ctx.ID,
	}
	_, err := snc.rpcClient.Call(context.TODO(), request)
	if err != nil {
		return err
	}

	return nil
}
