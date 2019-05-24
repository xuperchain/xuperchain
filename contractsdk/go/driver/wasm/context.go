// +build wasm

package wasm

import (
	"math/big"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/litepb"
)

const (
	methodPut         = "PutObject"
	methodGet         = "GetObject"
	methodDelete      = "DeleteObject"
	methodOutput      = "SetOutput"
	methodGetCallArgs = "GetCallArgs"
)

var (
	_ code.Context = (*handlerContext)(nil)
)

type handlerContext struct {
	method string
	args   map[string]interface{}
}

func newHandlerContext() (*handlerContext, error) {
	var request pb.GetCallArgsRequest
	var response pb.CallArgs
	err := syscall(methodGetCallArgs, &request, &response)
	if err != nil {
		return nil, err
	}
	args := make(map[string]interface{})
	for k, v := range response.Args {
		args[k] = string(v)
	}
	return &handlerContext{
		method: response.Method,
		args:   args,
	}, nil
}

func (c *handlerContext) Method() string {
	return c.method
}

func (c *handlerContext) Args() map[string]interface{} {
	return c.args
}

func (c *handlerContext) TxID() []byte {
	return nil
}

func (c *handlerContext) Caller() string {
	return ""
}

func (c *handlerContext) PutObject(key, value []byte) error {
	req := &pb.PutRequest{
		Key:   key,
		Value: value,
	}
	rep := new(pb.PutResponse)
	return syscall(methodPut, req, rep)
}

func (c *handlerContext) GetObject(key []byte) ([]byte, error) {
	req := &pb.GetRequest{
		Key: key,
	}
	rep := new(pb.GetResponse)
	err := syscall(methodGet, req, rep)
	if err != nil {
		return nil, err
	}
	return rep.Value, nil
}

func (c *handlerContext) DeleteObject(key []byte) error {
	req := &pb.DeleteRequest{
		Key: key,
	}
	rep := new(pb.DeleteResponse)
	return syscall(methodDelete, req, rep)
}

func (c *handlerContext) NewIterator(start, limit []byte) code.Iterator {
	return nil
}

func (c *handlerContext) QueryTx(txid []byte) (*code.TxStatus, error) {
	return nil, nil
}

func (c *handlerContext) QueryBlock(blockid []byte) (*code.Block, error) {
	return nil, nil
}
func (c *handlerContext) Transfer(to string, amount *big.Int) error {
	return nil
}

func (c *handlerContext) Call(module, method string, args map[string]interface{}) (*code.Response, error) {
	return nil, nil
}
