// +build wasm

package exec

import (
	"math/big"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/litepb"
)

const (
	methodPut          = "PutObject"
	methodGet          = "GetObject"
	methodDelete       = "DeleteObject"
	methodOutput       = "SetOutput"
	methodGetCallArgs  = "GetCallArgs"
	methodTransfer     = "Transfer"
	methodContractCall = "ContractCall"
)

var (
	_ code.Context = (*contractContext)(nil)
)

type contractContext struct {
	callArgs       pb.CallArgs
	bridgeCallFunc BridgeCallFunc
	header         pb.SyscallHeader
}

func newContractContext(ctxid int64, bridgeCallFunc BridgeCallFunc) *contractContext {
	return &contractContext{
		bridgeCallFunc: bridgeCallFunc,
		header: pb.SyscallHeader{
			Ctxid: ctxid,
		},
	}
}

func (c *contractContext) Init() error {
	var request pb.GetCallArgsRequest
	request.Header = &c.header
	return c.bridgeCallFunc(methodGetCallArgs, &request, &c.callArgs)
}

func (c *contractContext) Method() string {
	return c.callArgs.GetMethod()
}

func (c *contractContext) Args() map[string][]byte {
	return c.callArgs.GetArgs()
}

func (c *contractContext) Caller() string {
	return ""
}

func (c *contractContext) PutObject(key, value []byte) error {
	req := &pb.PutRequest{
		Header: &c.header,
		Key:    key,
		Value:  value,
	}
	rep := new(pb.PutResponse)
	return c.bridgeCallFunc(methodPut, req, rep)
}

func (c *contractContext) GetObject(key []byte) ([]byte, error) {
	req := &pb.GetRequest{
		Header: &c.header,
		Key:    key,
	}
	rep := new(pb.GetResponse)
	err := c.bridgeCallFunc(methodGet, req, rep)
	if err != nil {
		return nil, err
	}
	return rep.Value, nil
}

func (c *contractContext) DeleteObject(key []byte) error {
	req := &pb.DeleteRequest{
		Header: &c.header,
		Key:    key,
	}
	rep := new(pb.DeleteResponse)
	return c.bridgeCallFunc(methodDelete, req, rep)
}

func (c *contractContext) NewIterator(start, limit []byte) code.Iterator {
	return nil
}

func (c *contractContext) QueryTx(txid []byte) (*code.TxStatus, error) {
	return nil, nil
}

func (c *contractContext) QueryBlock(blockid []byte) (*code.Block, error) {
	return nil, nil
}

func (c *contractContext) Transfer(to string, amount *big.Int) error {
	req := &pb.TransferRequest{
		Header: &c.header,
		To:     to,
		Amount: amount.Text(10),
	}
	rep := new(pb.TransferResponse)
	return c.bridgeCallFunc(methodTransfer, req, rep)
}

func (c *contractContext) Call(module, contract, method string, args map[string][]byte) (*code.Response, error) {
	req := &pb.ContractCallRequest{
		Header:   &c.header,
		Module:   module,
		Contract: contract,
		Method:   method,
		Args:     args,
	}
	rep := new(pb.ContractCallResponse)
	err := c.bridgeCallFunc(methodContractCall, req, rep)
	if err != nil {
		return nil, err
	}
	return &code.Response{
		Status:  int(rep.Response.Status),
		Message: rep.Response.Message,
		Body:    rep.Response.Body,
	}, nil
}

func (c *contractContext) SetOutput(response *code.Response) error {
	req := &pb.SetOutputRequest{
		Header: &c.header,
		Response: &pb.Response{
			Status:  int32(response.Status),
			Message: response.Message,
			Body:    response.Body,
		},
	}
	rep := new(pb.SetOutputResponse)
	return c.bridgeCallFunc(methodOutput, req, rep)
}
