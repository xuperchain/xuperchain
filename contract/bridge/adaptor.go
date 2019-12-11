package bridge

import (
	"errors"
	"fmt"

	"github.com/xuperchain/xuperunion/contract"
)

const (
	initMethod = "initialize"
)

// ContractError indicates the error of the contract running result
type ContractError struct {
	Status  int
	Message string
}

// Error implements error interface
func (c *ContractError) Error() string {
	return fmt.Sprintf("contract error status:%d message:%s", c.Status, c.Message)
}

// vmContextImpl 为vm.Context的实现，
// 它组合了合约内核态数据(ctx)以及用户态的虚拟机数据(instance)
type vmContextImpl struct {
	ctx      *Context
	instance Instance
	release  func()
}

func (v *vmContextImpl) Invoke(method string, args map[string][]byte) (*contract.Response, error) {
	if !v.ctx.CanInitialize && method == initMethod {
		return nil, errors.New("invalid contract method " + method)
	}

	v.ctx.Method = method
	v.ctx.Args = args
	err := v.instance.Exec()
	if err != nil {
		return nil, err
	}
	if v.ctx.ResourceUsed().Exceed(v.ctx.ResourceLimits) {
		return nil, errors.New("resource exceeds")
	}

	if v.ctx.Output == nil {
		return nil, &ContractError{
			Status:  500,
			Message: "internal error",
		}
	}

	return &contract.Response{
		Status:  int(v.ctx.Output.GetStatus()),
		Message: v.ctx.Output.GetMessage(),
		Body:    v.ctx.Output.GetBody(),
	}, nil
}

func (v *vmContextImpl) ResourceUsed() contract.Limits {
	return v.ctx.ResourceUsed()
}

func (v *vmContextImpl) Release() error {
	// release the context of instance
	v.instance.Release()
	v.release()
	return nil
}

// vmImpl 为vm.VirtualMachine的实现
// 它是vmContextImpl的工厂类，根据不同的虚拟机类型(Executor)生成对应的vmContextImpl
type vmImpl struct {
	ctxmgr *ContextManager
	name   string
	exec   Executor
}

func (v *vmImpl) GetName() string {
	return v.name
}

func (v *vmImpl) NewContext(ctxCfg *contract.ContextConfig) (contract.Context, error) {
	ctx := v.ctxmgr.MakeContext()
	ctx.Cache = ctxCfg.XMCache
	ctx.ContractName = ctxCfg.ContractName
	ctx.Initiator = ctxCfg.Initiator
	ctx.AuthRequire = ctxCfg.AuthRequire
	ctx.ResourceLimits = ctxCfg.ResourceLimits
	ctx.CanInitialize = ctxCfg.CanInitialize
	ctx.Core = ctxCfg.Core
	ctx.TransferAmount = ctxCfg.TransferAmount
	ctx.ContractSet = ctxCfg.ContractSet
	if ctx.ContractSet == nil {
		ctx.ContractSet = make(map[string]bool)
		ctx.ContractSet[ctx.ContractName] = true
	}
	release := func() {
		v.ctxmgr.DestroyContext(ctx)
	}
	instance, err := v.exec.NewInstance(ctx)
	if err != nil {
		v.ctxmgr.DestroyContext(ctx)
		return nil, err
	}
	ctx.Instance = instance
	return &vmContextImpl{
		ctx:      ctx,
		instance: instance,
		release:  release,
	}, nil
}
