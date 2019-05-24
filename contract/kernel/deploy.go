package kernel

import (
	"errors"

	"github.com/xuperchain/xuperunion/contract/wasm"
	"github.com/xuperchain/xuperunion/permission/acl/utils"
)

// DeployMethod define Deploy type
type DeployMethod struct {
	vmm *wasm.VMManager
}

// Invoke Deploy contract method implementation
func (dm *DeployMethod) Invoke(ctx *KContext, args map[string][]byte) ([]byte, error) {
	// check if account exist
	accountName := args["account_name"]
	contractName := args["contract_name"]
	if accountName == nil || contractName == nil {
		return nil, errors.New("invoke DeployMethod error, account name or contract name is nil")
	}
	_, err := ctx.ModelCache.Get(utils.GetAccountBucket(), accountName)
	if err != nil {
		return nil, err
	}

	out, gasUsed, err := dm.vmm.DeployContract(ctx.ModelCache, args, ctx.GasLimit)
	if err != nil {
		return nil, err
	}
	ctx.AddGasUsed(gasUsed)

	// key: contract, value: account
	err = ctx.ModelCache.Put(utils.GetContract2AccountBucket(), contractName, accountName)
	if err != nil {
		return nil, err
	}
	key := utils.MakeAccountContractKey(string(accountName), string(contractName))
	err = ctx.ModelCache.Put(utils.GetAccount2ContractBucket(), []byte(key), []byte(utils.GetAccountContractValue()))
	if err != nil {
		return nil, err
	}
	return out, nil
}
