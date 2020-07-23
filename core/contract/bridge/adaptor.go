package bridge

import (
	"errors"
	"fmt"

	"github.com/xuperchain/xuperchain/core/contract"
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
	ctxmgr       *ContextManager
	xbridge      *XBridge
	name         string
	codeProvider ContractCodeProvider
}

func (v *vmImpl) GetName() string {
	return v.name
}

func (v *vmImpl) NewContext(ctxCfg *contract.ContextConfig) (contract.Context, error) {
	// test if contract exists
	desc, err := newCodeProvider(ctxCfg.XMCache).GetContractCodeDesc(ctxCfg.ContractName)
	if err != nil {
		return nil, err
	}
	tp, err := getContractType(desc)
	if err != nil {
		return nil, err
	}
	vm := v.xbridge.getCreator(tp)
	if vm == nil {
		return nil, fmt.Errorf("vm for contract type %s not supported", tp)
	}
	var cp ContractCodeProvider
	// 如果当前在部署合约，合约代码从cache获取
	// 合约调用的情况则从model中拿取合约代码，避免交易中包含合约代码的引用。
	if ctxCfg.ContractCodeFromCache {
		cp = newCodeProvider(ctxCfg.XMCache)
	} else {
		cp = newDescProvider(v.codeProvider, desc)
	}

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
	ctx.Logger = v.xbridge.debugLogger.New("contract", ctx.ContractName, "ctxid", ctx.ID)
	release := func() {
		v.ctxmgr.DestroyContext(ctx)
	}

	instance, err := vm.CreateInstance(ctx, cp)
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
