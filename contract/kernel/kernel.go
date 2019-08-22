package kernel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"sync"

	"github.com/mitchellh/mapstructure"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

// ChainRegister register blockchains
type ChainRegister interface {
	RegisterBlockChain(name string) error
	UnloadBlockChain(name string) error
}

// Kernel is the kernel contract
type Kernel struct {
	datapath          string
	log               log.Logger
	register          ChainRegister
	context           *contract.TxContext
	minNewChainAmount *big.Int        //创建平行链的最小花费
	newChainWhiteList map[string]bool //能创建链的address白名单
	mutex             *sync.Mutex
	bcName            string
}

var (
	// ErrBlockChainExist is returned when create an existed block chain
	ErrBlockChainExist = errors.New("BlockChain Exist")
	// ErrCreateBlockChain is returned when create block chain error
	ErrCreateBlockChain = errors.New("Create BlockChain error")
	// ErrMethodNotImplemented is returned when calling a nonexisted kernel method
	ErrMethodNotImplemented = errors.New("Method not implemented")
	// ErrNoEnoughUTXO is returned when has no enough money to create new chain
	ErrNoEnoughUTXO = errors.New("No enough money to create new chain")
	// ErrAddrNotInWhiteList is returned when address not in whitelist
	ErrAddrNotInWhiteList = errors.New("Address not in white list")
	// ErrPermissionDenied is returned when has no permission to call contract
	ErrPermissionDenied = errors.New("Permission denied to call this contract")
	// ErrInvalidChainName is returned when chain name is invalid
	ErrInvalidChainName = errors.New("Invalid Chain name")
)

// Init initialize kernel contract
func (k *Kernel) Init(path string, log log.Logger, register ChainRegister, bcName string) {
	k.datapath = path
	k.log = log
	k.register = register
	k.minNewChainAmount = big.NewInt(0)
	k.mutex = &sync.Mutex{}
	k.bcName = bcName
}

// SetMinNewChainAmount set the minimum amount of token to create a block chain
func (k *Kernel) SetMinNewChainAmount(amount string) {
	n := big.NewInt(0)
	n.SetString(amount, 10)
	k.minNewChainAmount = n
}

// SetNewChainWhiteList set the whitelit of address who can create new block chain
func (k *Kernel) SetNewChainWhiteList(whiteList map[string]bool) {
	k.newChainWhiteList = whiteList
}

// GetKVEngineType get kv engine type from xuper.json
func (k *Kernel) GetKVEngineType(data []byte) (string, error) {
	rootJSON := map[string]interface{}{}
	err := json.Unmarshal(data, &rootJSON)
	if err != nil {
		return "", err
	}
	kvEngineType := rootJSON["kvengine"]
	if kvEngineType == nil {
		return "default", nil
	}
	return kvEngineType.(string), nil
}

// GetCryptoType get crypto type from xuper.json
func (k *Kernel) GetCryptoType(data []byte) (string, error) {
	rootJSON := map[string]interface{}{}
	err := json.Unmarshal(data, &rootJSON)
	if err != nil {
		return "", err
	}
	cryptoType := rootJSON["crypto"]
	if cryptoType == nil {
		return client.CryptoTypeDefault, nil
	}
	return cryptoType.(string), nil
}

// init permission model of kernel contract
func (k *Kernel) initPermissionModel(data []byte) error {
	rootJSON := map[string]interface{}{}
	err := json.Unmarshal(data, &rootJSON)
	if err != nil {
		k.log.Warn("permission model data parse error", "error", err)
		return err
	}
	permModel, ok := rootJSON["permission"]
	if !ok || permModel == nil {
		return nil
	}
	switch permModel.(type) {
	case map[string]interface{}:
		modelset := permModel.(map[string]interface{})
		for method, model := range modelset {
			modelItem, ok := model.(map[string]interface{})
			if !ok {
				k.log.Warn("permission model parse error", "method", method, "model", model)
				continue
			}
			ruleItem, ok := modelItem["rule"]
			if !ok {
				k.log.Warn("permission ruleItem parse error", "method", method, "model", model)
				continue
			}
			ruleKeyword, ok := ruleItem.(string)
			if !ok {
				k.log.Warn("permission ruleKeyword parse error", "method", method, "model", model)
				continue
			}
			ruleInt, ok := pb.PermissionRule_value[ruleKeyword]
			if !ok {
				k.log.Warn("get permission rule by keyword error", "method", method, "model", model, "error", err)
				continue
			}
			rule := pb.PermissionRule(ruleInt)
			// process PermissionRule_NULL
			if rule == pb.PermissionRule_NULL {
				k.log.Info("kernel contract method initialized with Null permission rule", "method", method)
				continue
			}
			// TODO: unmarshall ACL and set contract method ACL
		}
	default:
		k.log.Warn("Permission field error in config")
	}
	return nil
}

// CreateBlockChain create a new block chain from xuper.json
func (k *Kernel) CreateBlockChain(name string, data []byte) error {
	k.log.Debug("create block chain by contract", "from", k.bcName, "toCreate", name)
	if k.bcName != "xuper" {
		k.log.Warn("only xuper chain can create side-chain", "bcName", k.bcName)
		return ErrPermissionDenied
	}
	fullpath := k.datapath + "/" + name
	if global.PathExists(fullpath) {
		k.log.Warn("fullpath exist", "fullpath", fullpath)
		return ErrBlockChainExist
	}
	err := os.Mkdir(fullpath, os.ModePerm)
	if err != nil {
		k.log.Warn("can't create path[" + fullpath + "] %v")
		return err
	}
	rootfile := fullpath + "/" + global.SBlockChainConfig
	err = ioutil.WriteFile(rootfile, data, 0666)
	if err != nil {
		k.log.Warn("write file error ", "file", rootfile)
		os.RemoveAll(fullpath)
		return err
	}
	kvEngineType, err := k.GetKVEngineType(data)
	if err != nil {
		k.log.Warn("failed to get `kvengine`", "err", err)
		return err
	}
	cryptoType, err := k.GetCryptoType(data)
	if err != nil {
		k.log.Warn("failed to get `crypto`", "err", err)
		return err
	}
	ledger, err := ledger.NewLedger(fullpath, k.log, nil, kvEngineType, cryptoType)
	if err != nil {
		k.log.Warn("NewLedger error", "fullpath", fullpath)
		os.RemoveAll(fullpath)
		return err
	}
	//TODO 因为是创建创世块，所以这里不填写publicKey和address, 后果是如果存在合约的话，肯定是执行失败
	utxovm, err := utxo.NewUtxoVM(name, ledger, fullpath, "", "", nil, k.log, false, kvEngineType, cryptoType)
	if err != nil {
		k.log.Warn("NewUtxoVM error", "fullpath", fullpath)
		os.RemoveAll(fullpath)
		return err
	}
	defer ledger.Close()
	defer utxovm.Close()
	// init kernel contract method permission model
	err = k.initPermissionModel(data)
	if err != nil {
		k.log.Warn("Init permission  model error", "err", err)
		return err
	}
	tx, err := utxovm.GenerateRootTx(data)
	if err != nil {
		k.log.Warn("GenerateRootTx error", "fullpath", fullpath)
		os.RemoveAll(fullpath)
		return err
	}
	utxovm.DebugTx(tx)
	txlist := []*pb.Transaction{tx}
	//b, err := ledger.FormatBlock(txlist, nil, nil, 0, nil)
	b, err := ledger.FormatRootBlock(txlist)
	if err != nil {
		k.log.Warn("format block error")
		os.RemoveAll(fullpath)
		return ErrCreateBlockChain
	}
	k.log.Trace("Start to ConfirmBlock")
	ledger.ConfirmBlock(b, true)
	k.log.Info("ConfirmBlock Success", "Height", 1)
	err = utxovm.Play(b.Blockid)
	if err != nil {
		k.log.Warn("utxo play error ", "error", err, "blockid", b.Blockid)
	}
	return nil
}

// RemoveBlockChainData remove all the data associate to the named blockchain
func (k *Kernel) RemoveBlockChainData(name string) error {
	if k.bcName != "xuper" {
		k.log.Warn("only xuper chain can remove side-chain", "bcName", k.bcName)
		return ErrPermissionDenied
	}
	fullpath := k.datapath + "/" + name
	trashPath := k.datapath + "/../trash"
	if !global.PathExists(trashPath) {
		err := os.Mkdir(trashPath, os.ModePerm)
		if err != nil {
			k.log.Warn("can't create path[" + trashPath + "] ")
			return err
		}
	}
	randomName := name + "_" + global.Glogid()
	return os.Rename(fullpath, trashPath+"/"+randomName)
}

func (k *Kernel) validateCreateBC(desc *contract.TxDesc) (string, string, error) {
	bcName := ""
	bcData := ""
	if desc.Args["name"] == nil {
		return bcName, bcData, errors.New("block chain name is empty")
	}
	if desc.Args["data"] == nil {
		return bcName, bcData, errors.New("first block data is empty")
	}
	switch desc.Args["name"].(type) {
	case string:
		bcName = desc.Args["name"].(string)
	default:
		return bcName, bcData, errors.New("the type of name should be string")
	}
	switch desc.Args["data"].(type) {
	case string:
		bcData = desc.Args["data"].(string)
	default:
		return bcName, bcData, errors.New("the type of data should be string")
	}
	return bcName, bcData, nil
}

func (k *Kernel) validateUpdateMaxBlockSize(desc *contract.TxDesc) error {
	for _, argName := range []string{"new_block_size", "old_block_size"} {
		if desc.Args[argName] == nil {
			return fmt.Errorf("miss argument in contact: %s", argName)
		}
		switch tp := desc.Args[argName].(type) {
		case float64:
			return nil
		default:
			return fmt.Errorf("invalid arg type: %s, %v", argName, tp)
		}
	}
	return nil
}

func (k *Kernel) validateUpdateForbiddenContract(desc *contract.TxDesc, name string) (*pb.InvokeRequest, error) {
	result := ledger.InvokeRequest{}

	// 检测参数
	if desc.Args[name] == nil {
		return nil, fmt.Errorf("miss argument in contract: %s", name)
	}
	// 获取参数内容
	args, ok := desc.Args[name].(interface{})
	if !ok {
		return nil, fmt.Errorf("validateUpdateForbiddenContract argName:%s invalid", name)
	}
	// 解析参数至结构体中
	err := mapstructure.Decode(args, &result)
	if err != nil {
		return nil, err
	}
	// 将ledger.InvokeRequest转化为pb.InvokeRequest
	forbiddenContractParam, transErr := ledger.InvokeRequestFromJSON2Pb([]ledger.InvokeRequest{result})
	if transErr != nil {
		return nil, transErr
	}

	k.log.Info("Kernel validateUpdateForbiddenContract succes", "param", forbiddenContractParam)
	if len(forbiddenContractParam) >= 1 {
		return forbiddenContractParam[0], nil
	}

	return nil, errors.New("validateForbiddenContract failed")
}

func (k *Kernel) validateUpdateReservedContract(desc *contract.TxDesc, name string) (
	[]*pb.InvokeRequest, error) {
	result := []ledger.InvokeRequest{}
	for _, argName := range []string{"old_reserved_contracts", "reserved_contracts"} {
		if desc.Args[argName] == nil {
			return nil, fmt.Errorf("miss argument in contact: %s", argName)
		}
		args, ok := desc.Args[argName].([]interface{})
		if !ok {
			return nil, fmt.Errorf("validateUpdateReservedContract argName:%s invalid", argName)
		}

		params := []ledger.InvokeRequest{}
		for _, arg := range args {
			param := ledger.InvokeRequest{}
			err := mapstructure.Decode(arg, &param)
			if err != nil {
				return nil, fmt.Errorf("validateUpdateReservedContract transfer invokeRequest failed")
			}
			params = append(params, param)
		}

		if argName == name {
			result = params
		}
	}

	reservedContractParams, _ := ledger.InvokeRequestFromJSON2Pb(result)

	k.log.Info("Kernel validateUpdateReservedContract success", "params", reservedContractParams)
	return reservedContractParams, nil
}

// Run implements ContractInterface
func (k *Kernel) Run(desc *contract.TxDesc) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	switch desc.Method {
	case "CreateBlockChain":
		bcName, bcData, err := k.validateCreateBC(desc) //需要校验，否则容易panic
		if err != nil {
			return err
		}
		k.log.Debug("contract: create block chain", "from", k.bcName, "toCrate", bcName)
		if k.bcName != "xuper" {
			k.log.Warn("only xuper chain can create side-chain", "bcName", k.bcName)
			return ErrPermissionDenied
		}

		if !desc.Tx.FromAddrInList(k.newChainWhiteList) {
			k.log.Warn("tx from addr not in whitelist to create blockchain")
			return ErrAddrNotInWhiteList
		}
		investment := desc.Tx.GetAmountByAddress(bcName)
		k.log.Info("create blockchain", "chain", bcName, "investment", investment, "need", k.minNewChainAmount)
		if investment.Cmp(k.minNewChainAmount) < 0 {
			return ErrNoEnoughUTXO
		}
		err = k.CreateBlockChain(bcName, []byte(bcData))
		if err == ErrBlockChainExist { //暂时忽略
			return nil
		}
		if err != nil {
			return err
		}
		if k.register != nil {
			k.log.Info("register block chain", "name", bcName)
			return k.register.RegisterBlockChain(bcName)
		}
		return nil
	case "UpdateMaxBlockSize":
		return k.runUpdateMaxBlockSize(desc)
	case "UpdateReservedContract":
		return k.runUpdateReservedContract(desc)
	case "UpdateForbiddenContract":
		return k.runUpdateForbiddenContract(desc)
	default:
		k.log.Warn("method not implemented", "method", desc.Method)
		return ErrMethodNotImplemented
	}
}

// Rollback implements ContractInterface
func (k *Kernel) Rollback(desc *contract.TxDesc) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	switch desc.Method {
	case "CreateBlockChain":
		bcName, _, err := k.validateCreateBC(desc) //需要校验，否则容易panic
		if err != nil {
			return err
		}
		fullpath := k.datapath + "/" + bcName
		if !global.PathExists(fullpath) {
			return nil //no need to rollback
		}
		err = k.RemoveBlockChainData(bcName)
		if err != nil {
			return err
		}
		if k.register != nil {
			return k.register.UnloadBlockChain(bcName)
		}
		return nil
	case "UpdateMaxBlockSize":
		return k.rollbackUpdateMaxBlockSize(desc)
	case "UpdateReservedContract":
		return k.rollbackUpdateReservedContract(desc)
	case "UpdateForbiddenContract":
		return k.rollbackUpdateForbiddenContract(desc)
	default:
		k.log.Warn("method not implemented", "method", desc.Method)
		return ErrMethodNotImplemented
	}
}

func (k *Kernel) runUpdateMaxBlockSize(desc *contract.TxDesc) error {
	if k.context == nil || k.context.LedgerObj == nil {
		return fmt.Errorf("failed to update block size, because no ledger object in context")
	}
	vErr := k.validateUpdateMaxBlockSize(desc)
	if vErr != nil {
		return vErr
	}
	newBlockSize := int64(desc.Args["new_block_size"].(float64))
	oldBlockSize := int64(desc.Args["old_block_size"].(float64))
	k.log.Info("update max block size", "old", oldBlockSize, "new", newBlockSize)
	curMaxBlockSize := k.context.LedgerObj.GetMaxBlockSize()
	if oldBlockSize != curMaxBlockSize {
		return fmt.Errorf("unexpected old block size, got %v, expected: %v", oldBlockSize, curMaxBlockSize)
	}
	err := k.context.LedgerObj.UpdateMaxBlockSize(newBlockSize, k.context.UtxoBatch)
	return err
}

func (k *Kernel) rollbackUpdateMaxBlockSize(desc *contract.TxDesc) error {
	if k.context == nil || k.context.LedgerObj == nil {
		return fmt.Errorf("failed to update block size, because no ledger object in context")
	}
	vErr := k.validateUpdateMaxBlockSize(desc)
	if vErr != nil {
		return vErr
	}
	oldBlockSize := int64(desc.Args["old_block_size"].(float64))
	err := k.context.LedgerObj.UpdateMaxBlockSize(oldBlockSize, k.context.UtxoBatch)
	return err
}

func (k *Kernel) runUpdateForbiddenContract(desc *contract.TxDesc) error {
	if k.context == nil || k.context.LedgerObj == nil {
		return fmt.Errorf("failed to update forbidden contract, because no ledger object in context")
	}

	oldParams, err := k.validateUpdateForbiddenContract(desc, "old_forbidden_contract")
	if err != nil {
		return err
	}
	k.log.Info("run update forbidden contract, params", "oldParams", oldParams)

	originalForbiddenContract := k.context.LedgerObj.GetMeta().ForbiddenContract

	originalModuleName := originalForbiddenContract.GetModuleName()
	originalContractName := originalForbiddenContract.GetContractName()
	originalMethodName := originalForbiddenContract.GetMethodName()
	originalArgs := originalForbiddenContract.GetArgs()
	oldParamsModuleName := oldParams.GetModuleName()
	oldParamsContractName := oldParams.GetContractName()
	oldParamsMethodName := oldParams.GetMethodName()
	oldParamsArgs := oldParams.GetArgs()

	// compare originalForbiddenContract with oldParams
	if originalModuleName != oldParamsModuleName || originalContractName != oldParamsContractName || originalMethodName != oldParamsMethodName || len(originalArgs) != len(oldParamsArgs) {
		return fmt.Errorf("old_forbidden_contract conf doesn't match current node forbidden_contract conf")
	}

	for oldKey, oldValue := range oldParamsArgs {
		if originalValue, ok := originalArgs[oldKey]; !ok || !reflect.DeepEqual(oldValue, originalValue) {
			return fmt.Errorf("old_forbidden_contract args doesn't match current node forbidden_contract args")
		}
	}

	params, err := k.validateUpdateForbiddenContract(desc, "forbidden_contract")
	if err != nil {
		return err
	}
	k.log.Info("update reservered contract", "params", params)
	err = k.context.LedgerObj.UpdateForbiddenContract(params, k.context.UtxoBatch)

	return err
}

func (k *Kernel) rollbackUpdateForbiddenContract(desc *contract.TxDesc) error {
	if k.context == nil || k.context.LedgerObj == nil {
		return fmt.Errorf("failed to update forbidden contract, because no ledger object in context")
	}
	params, err := k.validateUpdateForbiddenContract(desc, "old_forbidden_contract")
	if err != nil {
		return err
	}
	k.log.Info("rollback forbidden contract: params", "params", params)
	err = k.context.LedgerObj.UpdateForbiddenContract(params, k.context.UtxoBatch)

	return err
}

func (k *Kernel) runUpdateReservedContract(desc *contract.TxDesc) error {
	if k.context == nil || k.context.LedgerObj == nil {
		return fmt.Errorf("failed to update reservered contract, because no ledger object in context")
	}

	oldParams, err := k.validateUpdateReservedContract(desc, "old_reserved_contracts")
	if err != nil {
		return err
	}
	k.log.Info("run update reservered contract, params", "oldParams", oldParams)

	originalReservedContracts := k.context.LedgerObj.GetMeta().GetReservedContracts()

	for i, vold := range oldParams {
		for j, vorig := range originalReservedContracts {
			if i != j {
				continue
			}
			if vold.ModuleName != vorig.ModuleName || vold.ContractName != vorig.ContractName ||
				vold.MethodName != vorig.MethodName || len(vold.Args) != len(vorig.Args) {
				return fmt.Errorf("old_reserved_contracts values are not equal to the current node")
			}
			for k, vp := range vold.Args {
				if vo, ok := vorig.Args[k]; !ok || !reflect.DeepEqual(vp, vo) {
					return fmt.Errorf("old_reserved_contracts values are not equal to the current node")
				}
			}
		}
	}

	params, err := k.validateUpdateReservedContract(desc, "reserved_contracts")
	if err != nil {
		return err
	}
	k.log.Info("update reservered contract", "params", params)
	err = k.context.LedgerObj.UpdateReservedContract(params, k.context.UtxoBatch)
	return err
}

func (k *Kernel) rollbackUpdateReservedContract(desc *contract.TxDesc) error {
	if k.context == nil || k.context.LedgerObj == nil {
		return fmt.Errorf("failed to update reservered contract, because no ledger object in context")
	}
	params, err := k.validateUpdateReservedContract(desc, "old_reserved_contracts")
	if err != nil {
		return err
	}
	k.log.Info("rollback reservered contract: params", "params", params)
	if err != nil {
		return err
	}
	k.log.Info("rollback reservered contract")
	err = k.context.LedgerObj.UpdateReservedContract(params, k.context.UtxoBatch)
	return err
}

// Finalize implements ContractInterface
func (k *Kernel) Finalize(blockid []byte) error {
	return nil
}

// Stop implements ContractInterface
func (k *Kernel) Stop() {
}

// SetContext implements ContractInterface
func (k *Kernel) SetContext(context *contract.TxContext) error {
	k.context = context
	return nil
}

// ReadOutput implements ContractInterface
func (k *Kernel) ReadOutput(desc *contract.TxDesc) (contract.ContractOutputInterface, error) {
	return nil, nil
}

// GetVerifiableAutogenTx 实现VAT接口
func (k *Kernel) GetVerifiableAutogenTx(blockHeight int64, maxCount int, timestamp int64) ([]*pb.Transaction, error) {
	return nil, nil
}

// GetVATWhiteList 实现VAT接口
func (k *Kernel) GetVATWhiteList() map[string]bool {
	whiteList := map[string]bool{
		"UpdateMaxBlockSize": true,
	}
	return whiteList
}
