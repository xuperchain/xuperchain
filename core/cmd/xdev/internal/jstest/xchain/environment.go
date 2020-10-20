package xchain

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	_ "github.com/xuperchain/xuperchain/core/contract/evm"
	_ "github.com/xuperchain/xuperchain/core/contract/native"
	_ "github.com/xuperchain/xuperchain/core/contract/wasm/xvm"
	"github.com/xuperchain/xuperchain/core/pb"
)

type environment struct {
	xbridge *bridge.XBridge
	model   *mockStore
	basedir string
}

func newEnvironment() (*environment, error) {
	basedir, err := ioutil.TempDir("", "xdev-env")
	if err != nil {
		return nil, err
	}
	store := newMockStore()
	wasmconfig := &config.WasmConfig{
		Driver: "ixvm",
	}
	nativeconfig := &config.NativeConfig{
		Enable: true,
	}
	evmconfig := &config.EVMConfig{
		Enable: true,
		Driver: "evm",
	}

	xbridge, err := bridge.New(&bridge.XBridgeConfig{
		Basedir: basedir,
		VMConfigs: map[bridge.ContractType]bridge.VMConfig{
			bridge.TypeWasm:   wasmconfig,
			bridge.TypeNative: nativeconfig,
			bridge.TypeEvm:    evmconfig,
		},
		XModel:    store,
		LogWriter: os.Stderr,
	})
	if err != nil {
		os.RemoveAll(basedir)
		return nil, err
	}

	return &environment{
		xbridge: xbridge,
		model:   store,
		basedir: basedir,
	}, nil
}

type deployArgs struct {
	Name     string                 `json:"name"`
	Code     string                 `json:"code"`
	Lang     string                 `json:"lang"`
	InitArgs map[string]interface{} `json:"init_args"`
	Type     string                 `json:"type"`
	ABIFile  string                 `json:"abi"`

	trueArgs map[string][]byte
	codeBuf  []byte
}

func convertArgs(ori map[string]interface{}) map[string][]byte {
	ret := make(map[string][]byte)
	for k, v := range ori {
		ret[k] = []byte(v.(string))
	}
	return ret
}

func (e *environment) Deploy(args deployArgs) (*ContractResponse, error) {
	dargs := make(map[string][]byte)
	dargs["contract_name"] = []byte(args.Name)
	dargs["contract_code"] = args.codeBuf
	initArgs, err := json.Marshal(args.trueArgs)
	if err != nil {
		return nil, err
	}
	dargs["init_args"] = initArgs

	descpb := new(pb.WasmCodeDesc)
	descpb.Runtime = args.Lang
	descpb.ContractType = args.Type
	desc, err := proto.Marshal(descpb)
	if err != nil {
		return nil, err
	}
	dargs["contract_desc"] = desc

	xcache := e.model.NewCache()
	resp, _, err := e.xbridge.DeployContract(&contract.ContextConfig{
		XMCache:        xcache,
		ResourceLimits: contract.MaxLimits,
		Core:           new(chainCore),
	}, dargs)
	if err != nil {
		return nil, err
	}

	err = e.model.Commit(xcache)
	if err != nil {
		return nil, err
	}
	return newContractResponse(resp), nil
}

type invokeOptions struct {
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

type invokeArgs struct {
	Method   string                 `json:"method"`
	Args     map[string]interface{} `json:"args"`
	trueArgs map[string][]byte
	Options  invokeOptions
}

func (e *environment) ContractExists(name string) bool {
	vm, ok := e.xbridge.GetVirtualMachine("wasm")
	if !ok {
		return false
	}

	xcache := e.model.NewCache()

	ctx, err := vm.NewContext(&contract.ContextConfig{
		ContractName:   name,
		XMCache:        xcache,
		ResourceLimits: contract.MaxLimits,
	})
	if err != nil {
		return false
	}
	ctx.Release()
	return true
}

func (e *environment) Invoke(name string, args invokeArgs) (*ContractResponse, error) {
	vm, ok := e.xbridge.GetVirtualMachine("wasm")
	if !ok {
		return nil, errors.New("vm not found")
	}

	xcache := e.model.NewCache()

	ctx, err := vm.NewContext(&contract.ContextConfig{
		Initiator:      args.Options.Account,
		TransferAmount: args.Options.Amount,
		ContractName:   name,
		XMCache:        xcache,
		Core:           new(chainCore),
		ResourceLimits: contract.MaxLimits,
	})
	if err != nil {
		return nil, err
	}
	defer ctx.Release()

	resp, err := ctx.Invoke(args.Method, args.trueArgs)
	if err != nil {
		return nil, err
	}

	if resp.Status >= contract.StatusErrorThreshold {
		return newContractResponse(resp), nil
	}

	err = e.model.Commit(xcache)
	if err != nil {
		return nil, err
	}

	return newContractResponse(resp), nil
}

func (e *environment) Close() {
	os.RemoveAll(e.basedir)
}

type ContractResponse struct {
	Status  int
	Message string
	Body    string
}

func newContractResponse(resp *contract.Response) *ContractResponse {
	return &ContractResponse{
		Status:  resp.Status,
		Message: resp.Message,
		Body:    string(resp.Body),
	}
}
