package evm

import (
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
}

func (e *evmInstance) Exec() error {
	// fmt.Printf("%#v\n", e.ctx)
	var err error
	if e.ctx.Method == initializeMethod {
		code, err := e.cp.GetContractCode(e.ctx.ContractName)
		if err != nil {
			return err
		}
		//abi, err := e.cp.GetContractAbi(e.ctx.ContractName)
		//if err != nil {
		//	return err
		//}
		e.code = code
		e.abi = e.ctx.Args["contract_abi"]
	} else {
		codeDataWithVersion, err := e.ctx.Cache.Get("contract", evmCodeKey(e.ctx.ContractName))
		if err != nil {
			fmt.Println("get evm code error")
			return err
		}
		abiDataWithVersion, err := e.ctx.Cache.Get("contract", evmAbiKey(e.ctx.ContractName))
		if err != nil {
			fmt.Println("get evm abi error")
			return err
		}
		e.code = codeDataWithVersion.GetPureData().GetValue()
		e.abi = abiDataWithVersion.GetPureData().GetValue()
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

	var gas uint64 = 100000
	input := e.ctx.Args["input"]

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

	var resp string
	if e.ctx.Method != "" {
		resp, err = decodeRespWithAbiForEVM(string(e.abi), e.ctx.Method, out)
	} else {
		resp, err = decodeRespWithAbiForEVM(string(e.abi), "", out)
	}
	if err != nil {
		return err
	}

	//respBytes, err := hex.DecodeString(resp)
	//if err != nil {
	//	return err
	//}

	e.ctx.Output = &pb.Response{
		Status: 200,
		//Body:   []byte(hex.EncodeToString(out)),
		Body: []byte(resp),
	}
	return nil
}

func (e *evmInstance) ResourceUsed() contract.Limits {
	return contract.Limits{}
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
	var gas uint64 = 100000
	input := e.code
	params := engine.CallParams{
		CallType: exec.CallTypeCode,
		Origin:   caller,
		Caller:   caller,
		Callee:   crypto.ZeroAddress,
		Input:    input,
		Value:    big.NewInt(0),
		Gas:      &gas,
	}
	fmt.Printf("input:%x\n", params.Input)
	//e.ctx.Cache.Transfer()
	contractCode, err := e.vm.Execute(e.state, e.blockState, e, params, e.code)
	if err != nil {
		return err
	}
	key := evmCodeKey(e.ctx.ContractName)
	err = e.ctx.Cache.Put("contract", key, contractCode)
	if err != nil {
		return err
	}
	//key = evmAbiKey(e.ctx.ContractName)
	//err = e.ctx.Cache.Put("contract", key, e.ctx.Args["abi"])
	//if err != nil {
	//	return err
	//}
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

//func encodeArgsWithAbiForEVM(abiData []byte, funcName string, args []byte) ([]byte, error) {
//	packedBytes, _, err := abi.EncodeFunctionCall(string(abiData), funcName, nil, args)
//	if err != nil {
//		return nil, err
//	}
//	return packedBytes, nil
//}

func decodeRespWithAbiForEVM(abiData, funcName string, resp []byte) (string, error) {
	respArray := make([]string, 10, 10)
	Variables, err := abi.DecodeFunctionReturn(abiData, funcName, resp)
	if err != nil {
		return "", err
	}
	for i := range Variables {
		fmt.Println("key,value:", Variables[i].Name, Variables[i].Value)
		respArray = append(respArray, Variables[i].Value)
	}

	// transfer resp from array to json
	respStr, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	return string(respStr), nil
}
