package memory

import (
	"errors"
	"reflect"
	"strings"

	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contract/wasm/vm"
	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

type memoryInstanceCreator struct {
	config    vm.InstanceCreatorConfig
	codeCache map[string]reflect.Value
}

func newMemoryInstanceCreator(config *vm.InstanceCreatorConfig) (vm.InstanceCreator, error) {
	return &memoryInstanceCreator{
		config:    *config,
		codeCache: make(map[string]reflect.Value),
	}, nil
}

func (m *memoryInstanceCreator) CreateInstance(ctx *bridge.Context, cp vm.ContractCodeProvider) (vm.Instance, error) {
	var contractv reflect.Value
	contractv, ok := m.codeCache[ctx.ContractName]
	if !ok {
		codebuf, err := cp.GetContractCode(ctx.ContractName)
		if err != nil {
			return nil, err
		}
		contract, err := Decode(codebuf)
		if err != nil {
			return nil, err
		}
		contractv = reflect.ValueOf(contract)
		m.codeCache[ctx.ContractName] = contractv
	}
	return &memoryInstance{
		contract: contractv,
		codeContext: &codeContext{
			bridgeCtx: ctx,
			syscall:   m.config.SyscallService,
		},
	}, nil
}

func (m *memoryInstanceCreator) RemoveCache(contractName string) {
	delete(m.codeCache, contractName)
}

type memoryInstance struct {
	contract    reflect.Value
	codeContext *codeContext
}

func (m *memoryInstance) Exec(function string) error {
	args := make(map[string]interface{})
	for k, v := range m.codeContext.bridgeCtx.Args {
		args[k] = string(v)
	}
	m.codeContext.args = args
	var resp code.Response
	methodName := m.codeContext.bridgeCtx.Method
	methodv := m.contract.MethodByName(strings.Title(methodName))
	if !methodv.IsValid() {
		return errors.New("bad method " + methodName)
	}
	method, ok := methodv.Interface().(func(code.Context) code.Response)
	if !ok {
		return errors.New("bad method type " + methodName)
	}
	resp = method(m.codeContext)
	m.codeContext.bridgeCtx.Output = &pb.Response{
		Status:  int32(resp.Status),
		Message: resp.Message,
		Body:    resp.Body,
	}
	return nil
}

func (m *memoryInstance) GasUsed() int64 {
	return 0
}

func (m *memoryInstance) Release() {
}

func init() {
	vm.Register("memory", newMemoryInstanceCreator)
}
