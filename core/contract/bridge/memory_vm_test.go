package bridge

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/bridge/memrpc"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/exec"
)

type memoryInstanceCreator struct {
	config InstanceCreatorConfig
}

func newMemoryInstanceCreator(config *InstanceCreatorConfig) (InstanceCreator, error) {
	return &memoryInstanceCreator{
		config: *config,
	}, nil
}

func (m *memoryInstanceCreator) CreateInstance(ctx *Context, cp ContractCodeProvider) (Instance, error) {
	codebuf, err := cp.GetContractCode(ctx.ContractName)
	if err != nil {
		return nil, err
	}
	contract, err := memoryDecode(codebuf)
	if err != nil {
		return nil, err
	}
	log.Printf("%T", contract)
	return newMemoryInstance(contract, ctx, m.config.SyscallService), nil
}

func (m *memoryInstanceCreator) RemoveCache(contractName string) {
}

type memoryInstance struct {
	contract      code.Contract
	bridgeContext *Context
	rpcServer     *memrpc.Server
}

func newMemoryInstance(contract code.Contract, ctx *Context, syscall *SyscallService) *memoryInstance {
	return &memoryInstance{
		contract:      contract,
		bridgeContext: ctx,
		rpcServer:     memrpc.NewServer(syscall),
	}
}

func (m *memoryInstance) bridgeCall(method string, request proto.Message, response proto.Message) error {
	requestBuf, _ := proto.Marshal(request)
	responseBuf, err := m.rpcServer.CallMethod(context.TODO(), m.bridgeContext.ID, method, requestBuf)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(responseBuf, response)
	return err
}

func (m *memoryInstance) Exec() error {
	exec.RunContract(m.bridgeContext.ID, m.contract, m.bridgeCall)
	return nil
}

func (m *memoryInstance) ResourceUsed() contract.Limits {
	return contract.Limits{}
}

func (m *memoryInstance) Release() {
}

func (m *memoryInstance) Abort(msg string) {
}

// memoryEncode encodes a contract handler to bytes which can be later Decoded to contract
func memoryEncode(contract code.Contract) []byte {
	gob.Register(contract)
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(&contract)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// memoryDecode decodes bytes to contract
// The underlying type must be known to Decode function
func memoryDecode(buf []byte) (code.Contract, error) {
	var contract code.Contract
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	err := dec.Decode(&contract)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

func init() {
	Register(TypeNative, "memory", newMemoryInstanceCreator)
}
