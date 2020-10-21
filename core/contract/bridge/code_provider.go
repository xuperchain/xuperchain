package bridge

import (
	"errors"
	"fmt"

	"github.com/xuperchain/xuperchain/core/pb"
	xmodel_pb "github.com/xuperchain/xuperchain/core/xmodel/pb"

	"github.com/golang/protobuf/proto"
)

type xmodelStore interface {
	Get(bucket string, key []byte) (*xmodel_pb.VersionedData, error)
}

type codeProvider struct {
	xstore xmodelStore
}

func newCodeProvider(xstore xmodelStore) ContractCodeProvider {
	return &codeProvider{
		xstore: xstore,
	}
}

func (c *codeProvider) GetContractCode(name string) ([]byte, error) {
	value, err := c.xstore.Get("contract", contractCodeKey(name))
	if err != nil {
		return nil, fmt.Errorf("get contract code for '%s' error:%s", name, err)
	}
	codebuf := value.GetPureData().GetValue()
	if len(codebuf) == 0 {
		return nil, errors.New("empty wasm code")
	}
	return codebuf, nil
}

func (c *codeProvider) GetContractAbi(name string) ([]byte, error) {
	value, err := c.xstore.Get("contract", contractAbiKey(name))
	if err != nil {
		return nil, fmt.Errorf("get contract abi for '%s' error:%s", name, err)
	}
	abiBuf := value.GetPureData().GetValue()
	if len(abiBuf) == 0 {
		return nil, errors.New("empty abi")
	}
	return abiBuf, nil
}

func (c *codeProvider) GetContractCodeDesc(name string) (*pb.WasmCodeDesc, error) {
	value, err := c.xstore.Get("contract", ContractCodeDescKey(name))
	if err != nil {
		return nil, fmt.Errorf("get contract desc for '%s' error:%s", name, err)
	}
	descbuf := value.GetPureData().GetValue()
	// FIXME: 如果key不存在ModuleCache不应该返回零长度的value
	if len(descbuf) == 0 {
		return nil, errors.New("empty wasm code desc")
	}
	var desc pb.WasmCodeDesc
	err = proto.Unmarshal(descbuf, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

type descProvider struct {
	ContractCodeProvider
	desc *pb.WasmCodeDesc
}

func newDescProvider(cp ContractCodeProvider, desc *pb.WasmCodeDesc) ContractCodeProvider {
	return &descProvider{
		ContractCodeProvider: cp,
		desc:                 desc,
	}
}

func (d *descProvider) GetContractCodeDesc(name string) (*pb.WasmCodeDesc, error) {
	return d.desc, nil
}
