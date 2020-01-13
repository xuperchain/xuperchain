package xvm

import (
	"errors"

	"github.com/xuperchain/xuperunion/common/log"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contract/wasm/vm"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/xvm/debug"
	"github.com/xuperchain/xuperunion/xvm/exec"
	"github.com/xuperchain/xuperunion/xvm/runtime/emscripten"
	gowasm "github.com/xuperchain/xuperunion/xvm/runtime/go"
)

func createInstance(ctx *bridge.Context, code *contractCode, debugLogger *log.Logger) (vm.Instance, error) {
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
	instance.InitDebugWriter(debugLogger)
	return instance, nil
}

type xvmInstance struct {
	bridgeCtx *bridge.Context
	execCtx   exec.Context
	desc      pb.WasmCodeDesc
}

func (x *xvmInstance) Exec(function string) error {
	mem := x.execCtx.Memory()
	if mem == nil {
		return errors.New("bad contract, no memory")
	}
	var args []int64
	// go's entry function expects argc and argv these two arguments
	if x.desc.GetRuntime() == "go" {
		args = []int64{0, 0}
	}
	_, err := x.execCtx.Exec(function, args)
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

func (x *xvmInstance) InitDebugWriter(logger *log.Logger) {
	if logger == nil {
		return
	}
	instanceLogger := logger.New("contract", x.bridgeCtx.ContractName, "ctxid", x.bridgeCtx.ID)
	instanceLogWriter := newDebugWriter(instanceLogger)
	debug.SetWriter(x.execCtx, instanceLogWriter)
}
