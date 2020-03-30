package utxo

import (
	"encoding/json"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

// ResolveChain implement contract service
func (uv *UtxoVM) ResolveChain(chainName string) (*pb.CrossQueryMeta, error) {
	modelCache, err := xmodel.NewXModelCache(uv.GetXModel(), uv)
	if err != nil {
		return nil, err
	}
	contextConfig := &contract.ContextConfig{
		XMCache:        modelCache,
		ResourceLimits: contract.MaxLimits,
		ContractName:   "crossQueryNaming",
	}
	vm, err := uv.vmMgr3.GetVM("wasm")
	if err != nil {
		return nil, err
	}
	ctx, err := vm.NewContext(contextConfig)
	if err != nil {
		return nil, err
	}
	args := map[string][]byte{}
	args["name"] = []byte(chainName)
	invokeRes, invokeErr := ctx.Invoke("Resolve", args)
	if invokeErr != nil {
		ctx.Release()
		return nil, invokeErr
	}
	ctx.Release()
	res := &pb.CrossQueryMeta{}
	err = json.Unmarshal(invokeRes.Body, res)
	return res, err
}
