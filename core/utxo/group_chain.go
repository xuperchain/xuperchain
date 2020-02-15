package utxo

import (
	"fmt"
	"strings"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

func (uv *UtxoVM) QueryChainInList() map[string]bool {
	return uv.getChainInList()
}

func (uv *UtxoVM) QueryIPsInList(bcname string) map[string]bool {
	return uv.getIPsInList(bcname)
}

func (uv *UtxoVM) getIPsInList(bcname string) map[string]bool {
	ipMap := map[string]bool{}
	args := map[string][]byte{
		"bcname": []byte(bcname),
	}

	groupChainContract := uv.GetGroupChainContract()
	if groupChainContract == nil {
		return ipMap
	}
	moduleName := groupChainContract.ModuleName
	contractName := groupChainContract.ContractName
	methodName := groupChainContract.MethodName + "Node"

	fmt.Println("============>", "moduleName:", moduleName, "contractName:", contractName, "methodName:", methodName)

	if moduleName == "" && contractName == "" && methodName == "" {
		return ipMap
	}

	status, target, err := uv.queryGroupChain(moduleName, contractName, methodName, args)
	if status >= 400 || err != nil || string(target) == "" {
		return ipMap
	}
	res := strings.Split(string(target), "\x01")
	for _, v := range res {
		ipMap[v] = true
	}

	return ipMap
}

func (uv *UtxoVM) getChainInList() map[string]bool {
	chainMap := map[string]bool{}
	args := map[string][]byte{}

	groupChainContract := uv.GetGroupChainContract()
	if groupChainContract == nil {
		return chainMap
	}
	moduleName := groupChainContract.ModuleName
	contractName := groupChainContract.ContractName
	methodName := groupChainContract.MethodName + "Chain"

	if moduleName == "" && contractName == "" && methodName == "" {
		return chainMap
	}

	status, target, err := uv.queryGroupChain(moduleName, contractName, methodName, args)
	if status >= 400 || err != nil || string(target) == "" {
		return chainMap
	}
	res := strings.Split(string(target), "\x01")
	for _, v := range res {
		chainMap[v] = true
	}

	return chainMap
}

func (uv *UtxoVM) queryGroupChain(moduleName, contractName, methodName string, args map[string][]byte) (int, []byte, error) {
	modelCache, err := xmodel.NewXModelCache(uv.GetXModel(), uv)
	if err != nil {
		return 400, nil, err
	}
	contextConfig := &contract.ContextConfig{
		XMCache:        modelCache,
		ResourceLimits: contract.MaxLimits,
		ContractName:   contractName,
	}
	vm, err := uv.vmMgr3.GetVM(moduleName)
	if err != nil {
		return 400, nil, err
	}
	ctx, err := vm.NewContext(contextConfig)
	if err != nil {
		return 400, nil, err
	}
	invokeRes, invokeErr := ctx.Invoke(methodName, args)
	if invokeErr != nil {
		ctx.Release()
		return 400, nil, invokeErr
	}
	ctx.Release()

	return invokeRes.Status, invokeRes.Body, nil
}
