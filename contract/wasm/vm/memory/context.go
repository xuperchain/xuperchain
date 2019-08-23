package memory

import (
	"context"
	"math/big"

	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

type codeContext struct {
	args      map[string]interface{}
	bridgeCtx *bridge.Context
	syscall   *bridge.SyscallService
}

func (c *codeContext) Args() map[string]interface{} {
	return c.args
}

func (c *codeContext) TxID() []byte {
	panic("not implemented")
}

func (c *codeContext) Caller() string {
	panic("not implemented")
}

func (c *codeContext) PutObject(key []byte, value []byte) error {
	_, err := c.syscall.PutObject(context.TODO(), &pb.PutRequest{
		Header: &pb.SyscallHeader{
			Ctxid: c.bridgeCtx.ID,
		},
		Key:   key,
		Value: value,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *codeContext) GetObject(key []byte) ([]byte, error) {
	resp, err := c.syscall.GetObject(context.TODO(), &pb.GetRequest{
		Header: &pb.SyscallHeader{
			Ctxid: c.bridgeCtx.ID,
		},
		Key: key,
	})
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}

func (c *codeContext) DeleteObject(key []byte) error {
	_, err := c.syscall.DeleteObject(context.TODO(), &pb.DeleteRequest{
		Header: &pb.SyscallHeader{
			Ctxid: c.bridgeCtx.ID,
		},
		Key: key,
	})
	return err
}

func (c *codeContext) NewIterator(start []byte, limit []byte) code.Iterator {
	panic("not implemented")
}

func (c *codeContext) QueryTx(txid []byte) (*pb.Transaction, error) {
	panic("not implemented")
}

func (c *codeContext) QueryBlock(blockid []byte) (*pb.Block, error) {
	panic("not implemented")
}

func (c *codeContext) Transfer(to string, amount *big.Int) error {
	panic("not implemented")
}

func (c *codeContext) Call(module string, method string, args map[string]interface{}) (*code.Response, error) {
	panic("not implemented")
}
