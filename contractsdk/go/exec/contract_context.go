package exec

import (
	"math/big"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

const (
	methodPut          = "PutObject"
	methodGet          = "GetObject"
	methodDelete       = "DeleteObject"
	methodOutput       = "SetOutput"
	methodGetCallArgs  = "GetCallArgs"
	methodTransfer     = "Transfer"
	methodContractCall = "ContractCall"
	methodQueryTx      = "QueryTx"
	methodQueryBlock   = "QueryBlock"
	methodNewIterator  = "NewIterator"
)

var (
	_ code.Context = (*contractContext)(nil)
)

type contractContext struct {
	callArgs       pb.CallArgs
	contractArgs   map[string][]byte
	bridgeCallFunc BridgeCallFunc
	header         pb.SyscallHeader
}

func newContractContext(ctxid int64, bridgeCallFunc BridgeCallFunc) *contractContext {
	return &contractContext{
		contractArgs:   make(map[string][]byte),
		bridgeCallFunc: bridgeCallFunc,
		header: pb.SyscallHeader{
			Ctxid: ctxid,
		},
	}
}

func (c *contractContext) Init() error {
	var request pb.GetCallArgsRequest
	request.Header = &c.header
	err := c.bridgeCallFunc(methodGetCallArgs, &request, &c.callArgs)
	if err != nil {
		return err
	}
	for _, pair := range c.callArgs.GetArgs() {
		c.contractArgs[pair.GetKey()] = pair.GetValue()
	}
	return nil
}

func (c *contractContext) Method() string {
	return c.callArgs.GetMethod()
}

func (c *contractContext) Args() map[string][]byte {
	return c.contractArgs
}

func (c *contractContext) Caller() string {
	return ""
}

func (c *contractContext) Initiator() string {
	return c.callArgs.Initiator
}

func (c *contractContext) AuthRequire() []string {
	return c.callArgs.AuthRequire
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
	return newKvIterator(c, start, limit)
}

func (c *contractContext) QueryTx(txid string) (*pb.Transaction, error) {
	req := &pb.QueryTxRequest{
		Header: &c.header,
		Txid:   string(txid),
	}
	resp := new(pb.QueryTxResponse)
	if err := c.bridgeCallFunc(methodQueryTx, req, resp); err != nil {
		return nil, err
	}
	return resp.Tx, nil
}

func (c *contractContext) QueryBlock(blockid string) (*pb.Block, error) {
	req := &pb.QueryBlockRequest{
		Header:  &c.header,
		Blockid: string(blockid),
	}
	resp := new(pb.QueryBlockResponse)
	if err := c.bridgeCallFunc(methodQueryBlock, req, resp); err != nil {
		return nil, err
	}
	return resp.Block, nil
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
	var argPairs []*pb.ArgPair
	// 在合约里面单次合约调用的map迭代随机因子是确定的，因此这里不需要排序
	for key, value := range args {
		argPairs = append(argPairs, &pb.ArgPair{
			Key:   key,
			Value: value,
		})
	}
	req := &pb.ContractCallRequest{
		Header:   &c.header,
		Module:   module,
		Contract: contract,
		Method:   method,
		Args:     argPairs,
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
