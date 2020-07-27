package utxo

import (
	"errors"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

// This file is used to call contract from systerm
// 1. Get xpoa validates
// TODO:
// 2. Resorve cross query chain
// 3. Group chain

var (
	// ErrorNotConfirm return the responce not confirmed
	ErrorNotConfirm = errors.New("The result not confirmed")
)

// SystemCall used to call contract from systerm
func (uv *UtxoVM) SystemCall(reader xmodel.XMReader, contractName, methodName string, args map[string][]byte) ([]byte, error) {
	modelCache, err := xmodel.NewXModelCache(reader, uv)
	if err != nil {
		return nil, err
	}
	contextConfig := &contract.ContextConfig{
		XMCache:        modelCache,
		ResourceLimits: contract.MaxLimits,
		ContractName:   contractName,
	}
	vm, err := uv.vmMgr3.GetVM("wasm")
	if err != nil {
		return nil, err
	}
	ctx, err := vm.NewContext(contextConfig)
	if err != nil {
		return nil, err
	}
	invokeRes, invokeErr := ctx.Invoke(methodName, args)
	ctx.Release()
	if invokeErr != nil {
		return nil, invokeErr
	}
	return invokeRes.Body, err
}
