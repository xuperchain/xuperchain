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
	xchainpb "github.com/xuperchain/xuperchain/core/pb"
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
		e.code, err = e.cp.GetContractCode(e.ctx.ContractName)
		if err != nil {
			fmt.Println("get evm code error")
			return err
		}
		// there is no need to get evm abi
		//e.abi, err = e.cp.GetContractAbi(e.ctx.ContractName)
		//if err != nil {
		//	fmt.Println("get evm abi error")
		//	return err
		//}
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
	contractName, _, err := DetermineEVMAddress(log.Address)
	if err != nil {
		return err
	}

	contractAbiByte, err := e.cp.GetContractAbi(contractName)
	if err != nil {
		return err
	}
	event,err := unpackEventFromAbi(contractAbiByte,contractName,log)
	if err != nil {
		return err
	}
	e.ctx.Cache.AddEvent(event)
	return nil
}


func unpackEventFromAbi(abiByte []byte,contractName string,log *exec.LogEvent) (*xchainpb.ContractEvent,error){
	var eventID abi.EventID
	copy(eventID[:], log.GetTopic(0).Bytes())
	spec, err := abi.ReadSpec(abiByte)
	if err != nil {
		return nil,err
	}
	eventSpec, ok := spec.EventsByID[eventID]
	if !ok {
		return nil,fmt.Errorf("The Event By ID Not Found ")
	}

	vals := make([]interface{}, len(eventSpec.Inputs))
	for i := range vals {
		vals[i] = new(string)
	}
	if err := abi.UnpackEvent(eventSpec, log.Topics, log.Data, vals...); err != nil {
		return nil,err
	}

	fields := []interface{}{}
	for i := range vals {
		val := vals[i].(*string)
		m := make(map[string]string)
		m[eventSpec.Inputs[i].Name] = *val
		fields = append(fields, m)
	}
	event := &xchainpb.ContractEvent{
		Contract: contractName,
	}
	event.Name = eventSpec.Name
	data, err := json.Marshal(fields)
	if err != nil {
		return nil,err
	}
	event.Body = data
	return event,nil
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
	input := e.code
	params := engine.CallParams{
		CallType: exec.CallTypeCode,
		Origin:   caller,
		Caller:   caller,
		Callee:   callee,
		Input:    input,
		Value:    big.NewInt(0),
		Gas:      &gas,
	}
	contractCode, err := e.vm.Execute(e.state, e.blockState, e, params, e.code)
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

//func encodeArgsWithAbiForEVM(abiData []byte, funcName string, args []byte) ([]byte, error) {
//	packedBytes, _, err := abi.EncodeFunctionCall(string(abiData), funcName, nil, args)
//	if err != nil {
//		return nil, err
//	}
//	return packedBytes, nil
//}

//func DecodeRespWithAbiForEVM(abiData, funcName string, resp []byte) error {
//	Variables, err := abi.DecodeFunctionReturn(abiData, funcName, resp)
//	if err != nil {
//		return err
//	}
//
//	fmt.Println("contract response:")
//	for i := range Variables {
//		fmt.Println("key,value:", Variables[i].Name, Variables[i].Value)
//	}
//
//	return nil
//}
