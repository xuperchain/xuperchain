package native

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

type chainClient struct {
	header    *pb.SyscallHeader
	rpcClient pb.SyscallClient
}

func newChainClient(ctxid int64, rpcClient pb.SyscallClient) *chainClient {
	return &chainClient{
		rpcClient: rpcClient,
		header: &pb.SyscallHeader{
			Ctxid: ctxid,
		},
	}
}

func (c *chainClient) QueryTx(txid []byte) (*code.TxStatus, error) {
	request := &pb.QueryTxRequest{
		Header: c.header,
		Txid:   txid,
	}
	ctx := context.Background()
	resp, err := c.rpcClient.QueryTx(ctx, request)
	if err != nil {
		return nil, err
	}
	txStatus := new(code.TxStatus)
	err = json.Unmarshal(resp.GetTx(), txStatus)
	if err != nil {
		return nil, err
	}
	return txStatus, nil
}

func (c *chainClient) QueryBlock(blockid []byte) (*code.Block, error) {
	request := &pb.QueryBlockRequest{
		Header:  c.header,
		Blockid: blockid,
	}
	ctx := context.Background()
	resp, err := c.rpcClient.QueryBlock(ctx, request)
	if err != nil {
		return nil, err
	}
	block := new(code.Block)
	err = json.Unmarshal(resp.GetBlock(), block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (c *chainClient) Transfer(to string, amount *big.Int) error {
	request := &pb.TransferRequest{
		Header: c.header,
		To:     to,
		Amount: amount.String(),
	}
	ctx := context.Background()
	_, err := c.rpcClient.Transfer(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (c *chainClient) Call(module, method string, args map[string]interface{}) (*code.Response, error) {
	argbuf, _ := json.Marshal(args)
	request := &pb.ContractCallRequest{
		Header: c.header,
		Module: module,
		Method: method,
		Args:   string(argbuf),
	}
	ctx := context.Background()
	resp, err := c.rpcClient.ContractCall(ctx, request)
	if err != nil {
		return nil, err
	}
	response := new(code.Response)
	err = json.Unmarshal(resp.Response, response)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response error:%s", err)
	}
	return response, nil
}

func (c *chainClient) PutObject(key, value []byte) error {
	ctx := context.TODO()
	request := &pb.PutRequest{
		Header: c.header,
		Key:    key,
		Value:  value,
	}
	_, err := c.rpcClient.PutObject(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func (c *chainClient) GetObject(key []byte) ([]byte, error) {
	ctx := context.TODO()
	request := &pb.GetRequest{
		Header: c.header,
		Key:    key,
	}
	resp, err := c.rpcClient.GetObject(ctx, request)
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
}

func (c *chainClient) DeleteObject(key []byte) error {
	ctx := context.TODO()
	request := &pb.DeleteRequest{
		Header: c.header,
		Key:    key,
	}
	_, err := c.rpcClient.DeleteObject(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

// TODO: 分段请求iterator
func (c *chainClient) NewIterator(start, limit []byte) code.Iterator {
	ctx := context.TODO()
	request := &pb.IteratorRequest{
		Header: c.header,
		Start:  start,
		Limit:  limit,
	}
	resp, err := c.rpcClient.NewIterator(ctx, request)
	if err != nil {
		return newErrorArrayIterator(err)
	}
	return newArrayIterator(resp.Items)
}
