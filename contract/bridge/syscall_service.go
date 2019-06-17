// Copyright (c) 2019, Baidu.com, Inc. All Rights Reserved.

package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

// SyscallService is the handler of contract syscalls
type SyscallService struct {
	ctxmgr *ContextManager
}

// NewSyscallService instances a new SyscallService
func NewSyscallService(ctxmgr *ContextManager) *SyscallService {
	return &SyscallService{
		ctxmgr: ctxmgr,
	}
}

// Ping implements Syscall interface
func (c *SyscallService) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	return new(pb.PingResponse), nil
}

// QueryBlock implements Syscall interface
func (c *SyscallService) QueryBlock(ctx context.Context, in *pb.QueryBlockRequest) (*pb.QueryBlockResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}

	block, err := nctx.Cache.QueryBlock(in.Blockid)
	if err != nil {
		return nil, err
	}

	blockbuf, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	return &pb.QueryBlockResponse{
		Block: blockbuf,
	}, nil
}

// QueryTx implements Syscall interface
func (c *SyscallService) QueryTx(ctx context.Context, in *pb.QueryTxRequest) (*pb.QueryTxResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}
	tx, err := nctx.Cache.QueryTx(in.Txid)
	if err != nil {
		return nil, err
	}

	txbuf, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return &pb.QueryTxResponse{
		Tx: txbuf,
	}, nil
}

// Transfer implements Syscall interface
func (c *SyscallService) Transfer(ctx context.Context, in *pb.TransferRequest) (*pb.TransferResponse, error) {
	resp := &pb.TransferResponse{}
	return resp, nil
}

// ContractCall implements Syscall interface
func (c *SyscallService) ContractCall(ctx context.Context, in *pb.ContractCallRequest) (*pb.ContractCallResponse, error) {
	resp := new(pb.ContractCallResponse)
	return resp, nil
}

// PutObject implements Syscall interface
func (c *SyscallService) PutObject(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}
	if in.Value == nil {
		return nil, errors.New("put nil value")
	}

	err := nctx.Cache.Put(nctx.ContractName, in.Key, in.Value)
	if err != nil {
		return nil, err
	}
	log.Printf("put value:%s=%s", in.Key, in.Value)
	return &pb.PutResponse{}, nil
}

// GetObject implements Syscall interface
func (c *SyscallService) GetObject(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}

	value, err := nctx.Cache.Get(nctx.ContractName, in.Key)
	if err != nil {
		return nil, err
	}
	log.Printf("get value:%s=%s", in.Key, value.GetPureData().GetValue())
	return &pb.GetResponse{
		Value: value.GetPureData().GetValue(),
	}, nil
}

// DeleteObject implements Syscall interface
func (c *SyscallService) DeleteObject(ctx context.Context, in *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}

	err := nctx.Cache.Del(nctx.ContractName, in.Key)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteResponse{}, nil
}

// NewIterator implements Syscall interface
func (c *SyscallService) NewIterator(ctx context.Context, in *pb.IteratorRequest) (*pb.IteratorResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}

	iter, err := nctx.Cache.Select(nctx.ContractName, in.Start, in.Limit)
	if err != nil {
		return nil, err
	}
	out := new(pb.IteratorResponse)
	for iter.Next() {
		out.Items = append(out.Items, &pb.IteratorItem{
			Key:   iter.Key(),
			Value: iter.Data().GetPureData().GetValue(),
		})
	}
	if iter.Error() != nil {
		return nil, err
	}
	iter.Release()
	return out, nil
}

// GetCallArgs implements Syscall interface
func (c *SyscallService) GetCallArgs(ctx context.Context, in *pb.GetCallArgsRequest) (*pb.CallArgs, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}
	return &pb.CallArgs{
		Method: nctx.Method,
		Args:   nctx.Args,
		// TODO zq
	}, nil
}

// SetOutput implements Syscall interface
func (c *SyscallService) SetOutput(ctx context.Context, in *pb.SetOutputRequest) (*pb.SetOutputResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.Header.Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}
	nctx.Output = in.GetResponse()
	return new(pb.SetOutputResponse), nil
}
