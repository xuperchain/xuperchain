package utxo

import (
	"encoding/hex"
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
func (uv *UtxoVM) SystemCall(reader xmodel.XMReader, contractName, methodName, watchedKey string, args map[string][]byte) ([]byte, int64, int64, error) {
	var lastConfirmedTime int64
	var lastConfirmedHeight int64
	modelCache, err := xmodel.NewXModelCache(reader, uv)
	if err != nil {
		return nil, lastConfirmedTime, lastConfirmedHeight, err
	}
	contextConfig := &contract.ContextConfig{
		XMCache:        modelCache,
		ResourceLimits: contract.MaxLimits,
		ContractName:   contractName,
	}
	vm, err := uv.vmMgr3.GetVM("wasm")
	if err != nil {
		return nil, lastConfirmedTime, lastConfirmedHeight, err
	}
	ctx, err := vm.NewContext(contextConfig)
	if err != nil {
		return nil, lastConfirmedTime, lastConfirmedHeight, err
	}
	invokeRes, invokeErr := ctx.Invoke(methodName, args)
	if invokeErr != nil {
		ctx.Release()
		return nil, lastConfirmedTime, lastConfirmedHeight, invokeErr
	}
	ctx.Release()
	vData, err := reader.Get(contractName, []byte(watchedKey))
	if err != nil {
		uv.xlog.Error("SystemCall get nothing from XMReader", "RefTxid", hex.EncodeToString(vData.GetRefTxid()))
		return nil, lastConfirmedTime, lastConfirmedHeight, ErrorNotConfirm
	}
	block, err := uv.ledger.QueryBlockByTxid(vData.GetRefTxid())
	if block.GetTimestamp() > lastConfirmedTime {
		lastConfirmedTime = block.GetTimestamp()
		lastConfirmedHeight = block.GetHeight()
	}
	return invokeRes.Body, lastConfirmedTime, lastConfirmedHeight, err
}
