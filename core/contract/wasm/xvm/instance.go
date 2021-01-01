package xvm

import (
	"context"
	"errors"

	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	sdkpb "github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xvm/debug"
	"github.com/xuperchain/xuperchain/core/xvm/exec"
	"github.com/xuperchain/xuperchain/core/xvm/runtime/emscripten"
	gowasm "github.com/xuperchain/xuperchain/core/xvm/runtime/go"
)

func createInstance(ctx *bridge.Context, code *contractCode, syscall *bridge.SyscallService) (bridge.Instance, error) {
	log.Info("instance resource limit", "limits", ctx.ResourceLimits)
	execCtx, err := code.ExecCode.NewContext(&exec.ContextConfig{
		GasLimit: ctx.ResourceLimits.Cpu,
	})
	if err != nil {
		log.Error("create contract context error", "error", err, "contract", ctx.ContractName)
		return nil, err
	}
	switch code.Desc.GetRuntime() {
	case "go":
		gowasm.RegisterRuntime(execCtx)
	case "c":
		err = emscripten.Init(execCtx)
		if err != nil {
			return nil, err
		}
	}
	execCtx.SetUserData(contextIDKey, ctx.ID)
	instance := &xvmInstance{
		bridgeCtx: ctx,
		execCtx:   execCtx,
		desc:      code.Desc,
	}
	instance.InitDebugWriter(syscall)
	return instance, nil
}

type xvmInstance struct {
	bridgeCtx *bridge.Context
	execCtx   exec.Context
	desc      pb.WasmCodeDesc
	syscall   *bridge.SyscallService
}

func (x *xvmInstance) Exec() error {
	mem := x.execCtx.Memory()
	if mem == nil {
		return errors.New("bad contract, no memory")
	}
	var args []int64
	// go's entry function expects argc and argv these two arguments
	if x.desc.GetRuntime() == "go" {
		args = []int64{0, 0}
	}
	function, err := x.guessEntry()
	if err != nil {
		return err
	}
	_, err = x.execCtx.Exec(function, args)
	if err != nil {
		log.Error("exec contract error", "error", err, "contract", x.bridgeCtx.ContractName)
	}
	return err
}

func (x *xvmInstance) ResourceUsed() contract.Limits {
	limits := contract.Limits{
		Cpu: x.execCtx.GasUsed(),
	}
	mem := x.execCtx.Memory()
	if mem != nil {
		limits.Memory = int64(len(mem))
	}
	return limits
}

func (x *xvmInstance) Release() {
	x.execCtx.Release()
}

func (x *xvmInstance) Abort(msg string) {
	exec.Throw(exec.NewTrap(msg))
}

func (x *xvmInstance) InitDebugWriter(syscall *bridge.SyscallService) {
	if syscall == nil {
		return
	}

	flushfunc := func(str string) {
		request := &sdkpb.PostLogRequest{
			Header: &sdkpb.SyscallHeader{
				Ctxid: x.bridgeCtx.ID,
			},
			Entry: str,
		}
		syscall.PostLog(context.Background(), request)
	}
	instanceLogWriter := newDebugWriter(flushfunc)
	debug.SetWriter(x.execCtx, instanceLogWriter)
}

func (x *xvmInstance) guessEntry() (string, error) {
	runtime := x.desc.GetRuntime()
	switch runtime {
	case "go":
		return "run", nil
	case "c":
		return "_" + x.bridgeCtx.Method, nil
	default:
		return "", errors.New("runtime must be either go or c")
	}
}
