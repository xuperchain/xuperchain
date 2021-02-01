package evm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/execution/exec"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	xabi "github.com/xuperchain/xuperchain/core/contract/evm/abi"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
)

const (
	initializeMethod = "initialize"
)

type evmCreator struct {
	vm *evm.EVM
}

func newEvmCreator(config *bridge.InstanceCreatorConfig) (bridge.InstanceCreator, error) {
	opt := evm.Options{}
	opt.DebugOpcodes = true
	vm := evm.New(opt)
	return &evmCreator{
		vm: vm,
	}, nil
}

// CreateInstance instances an evm virtual machine instance which can run a single contract call
func (e *evmCreator) CreateInstance(ctx *bridge.Context, cp bridge.ContractCodeProvider) (bridge.Instance, error) {
	state := newStateManager(ctx)
	blockState := newBlockStateManager(ctx)
	return &evmInstance{
		vm:         e.vm,
		ctx:        ctx,
		state:      state,
		blockState: blockState,
		cp:         cp,
	}, nil
}

func (e *evmCreator) RemoveCache(name string) {
}

type evmInstance struct {
	vm         *evm.EVM
	ctx        *bridge.Context
	state      *stateManager
	blockState *blockStateManager
	cp         bridge.ContractCodeProvider
	code       []byte
	abi        []byte
	gasUsed    uint64
}

func (e *evmInstance) Exec() error {
	var err error

	// 获取合约的 code。
	e.code, err = e.cp.GetContractCode(e.ctx.ContractName)
	if err != nil {
		return err
	}

	// 部署合约或者调用合约时参数未使用 abi 编码时需要获取到合约的 abi。执行结果也需要使用 abi 解析。
	e.abi, err = e.cp.GetContractAbi(e.ctx.ContractName)
	if err != nil {
		return err
	}

	if e.ctx.Method == initializeMethod {
		return e.deployContract()
	}

	var caller crypto.Address
	if DetermineContractAccount(e.state.ctx.Initiator) {
		caller, err = ContractAccountToEVMAddress(e.state.ctx.Initiator)
	} else {
		caller, err = XchainToEVMAddress(e.state.ctx.Initiator)
	}
	if err != nil {
		return err
	}

	callee, err := ContractNameToEVMAddress(e.ctx.ContractName)
	if err != nil {
		return err
	}

	gas := uint64(contract.MaxLimits.Cpu)

	input := []byte{}
	if e.ctx.AbiNotEncoded {
		// 客户端未将参数进行编码，此处来编码。
		input, err = e.encodeInvokeInput()
		if err != nil {
			return err
		}
	} else {
		// 在客户端已经将参数 abi 编码。
		input = e.ctx.Args["input"]
	}

	value := big.NewInt(0)
	ok := false
	if e.ctx.TransferAmount != "" {
		value, ok = new(big.Int).SetString(e.ctx.TransferAmount, 0)
		if !ok {
			return fmt.Errorf("get evm value error")
		}
	}
	params := engine.CallParams{
		CallType: exec.CallTypeCode,
		Caller:   caller,
		Callee:   callee,
		Input:    input,
		Value:    value,
		Gas:      &gas,
	}
	out, err := e.vm.Execute(e.state, e.blockState, e, params, e.code)
	if err != nil {
		return err
	}

	// 执行结果根据 abi 解码，返回 json 格式的数组。
	out, err = decodeRespWithAbiForEVM(string(e.abi), e.ctx.Method, out)
	if err != nil {
		return err
	}

	e.gasUsed = uint64(contract.MaxLimits.Cpu) - *params.Gas

	e.ctx.Output = &pb.Response{
		Status: 200,
		Body:   out,
	}
	return nil
}

func (e *evmInstance) ResourceUsed() contract.Limits {
	return contract.Limits{
		Cpu: int64(e.gasUsed),
	}
}

func (e *evmInstance) Release() {
}

func (e *evmInstance) Abort(msg string) {
}

func (e *evmInstance) Call(call *exec.CallEvent, exception *errors.Exception) error {
	return nil
}

func (e *evmInstance) Log(log *exec.LogEvent) error {
	return nil
}

func (e *evmInstance) deployContract() error {
	var caller crypto.Address
	var err error
	if DetermineContractAccount(e.state.ctx.Initiator) {
		caller, err = ContractAccountToEVMAddress(e.state.ctx.Initiator)
	} else {
		caller, err = XchainToEVMAddress(e.state.ctx.Initiator)
	}
	if err != nil {
		return err
	}

	callee, err := ContractNameToEVMAddress(e.ctx.ContractName)
	if err != nil {
		return err
	}

	gas := uint64(contract.MaxLimits.Cpu)

	input := []byte{}
	if e.ctx.AbiNotEncoded {
		// 客户端未将参数编码。
		if input, err = e.encodeDeployInput(); err != nil {
			return err
		}
	} else {
		input = e.code
	}

	params := engine.CallParams{
		CallType: exec.CallTypeCode,
		Origin:   caller,
		Caller:   caller,
		Callee:   callee,
		Input:    input,
		Value:    big.NewInt(0),
		Gas:      &gas,
	}
	contractCode, err := e.vm.Execute(e.state, e.blockState, e, params, input)
	if err != nil {
		return err
	}

	key := evmCodeKey(e.ctx.ContractName)
	err = e.ctx.Cache.Put("contract", key, contractCode)
	if err != nil {
		return err
	}

	e.gasUsed = uint64(contract.MaxLimits.Cpu) - *params.Gas

	e.ctx.Output = &pb.Response{
		Status: 200,
	}
	return nil
}

func evmCodeKey(contractName string) []byte {
	return []byte(contractName + "." + "code")
}

func evmAbiKey(contractName string) []byte {
	return []byte(contractName + "." + "abi")
}

func init() {
	bridge.Register(bridge.TypeEvm, "evm", newEvmCreator)
}

// func encodeArgsWithAbiForEVM(abiData []byte, funcName string, args []byte) ([]byte, error) {
// 	packedBytes, _, err := abi.EncodeFunctionCall(string(abiData), funcName, nil, args)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return packedBytes, nil
// }

func decodeRespWithAbiForEVM(abiData, funcName string, resp []byte) ([]byte, error) {
	Variables, err := abi.DecodeFunctionReturn(abiData, funcName, resp)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]string, 0, len(Variables))
	for _, v := range Variables {
		result = append(result, map[string]string{
			v.Name: v.Value,
		})
	}

	out, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (e *evmInstance) encodeDeployInput() ([]byte, error) {
	enc, err := xabi.New(e.abi)
	if err != nil {
		return nil, err
	}

	args := make(map[string]interface{}, len(e.ctx.Args))
	for k, v := range e.ctx.Args {
		args[k] = string(v)
	}

	input, err := enc.Encode("", args)
	if err != nil {
		return nil, err
	}

	evmCode := string(e.code) + hex.EncodeToString(input)
	codeBuf, err := hex.DecodeString(evmCode)
	if err != nil {
		return nil, err
	}

	return codeBuf, nil
}

func (e *evmInstance) encodeInvokeInput() ([]byte, error) {
	enc, err := xabi.New(e.abi)
	if err != nil {
		return nil, err
	}

	args := make(map[string]interface{}, len(e.ctx.Args))
	for k, v := range e.ctx.Args {
		args[k] = string(v)
	}

	input, err := enc.Encode(e.ctx.Method, args)
	if err != nil {
		return nil, err
	}

	return input, nil
}
