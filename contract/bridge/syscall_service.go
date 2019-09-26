// Copyright (c) 2019, Baidu.com, Inc. All Rights Reserved.

package bridge

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"

	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
	xchainpb "github.com/xuperchain/xuperunion/pb"
)

var (
	ErrOutOfDiskLimit = errors.New("out of disk limit")
)

const (
	DefaultCap = 1000
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

	rawBlockid, err := hex.DecodeString(in.Blockid)
	if err != nil {
		return nil, err
	}

	block, err := nctx.Cache.QueryBlock(rawBlockid)
	if err != nil {
		return nil, err
	}

	txids := []string{}
	for _, t := range block.Transactions {
		txids = append(txids, hex.EncodeToString(t.Txid))
	}

	blocksdk := &pb.Block{
		Blockid:  hex.EncodeToString(block.Blockid),
		PreHash:  hex.EncodeToString(block.PreHash),
		Proposer: block.Proposer,
		Sign:     hex.EncodeToString(block.Sign),
		Pubkey:   block.Pubkey,
		Height:   block.Height,
		Txids:    txids,
		TxCount:  block.TxCount,
		InTrunk:  block.InTrunk,
		NextHash: hex.EncodeToString(block.NextHash),
	}

	return &pb.QueryBlockResponse{
		Block: blocksdk,
	}, nil
}

// QueryTx implements Syscall interface
func (c *SyscallService) QueryTx(ctx context.Context, in *pb.QueryTxRequest) (*pb.QueryTxResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}

	rawTxid, err := hex.DecodeString(in.Txid)
	if err != nil {
		return nil, err
	}

	tx, confirmed, err := nctx.Cache.QueryTx(rawTxid)
	if err != nil {
		return nil, err
	}

	if !confirmed {
		return nil, fmt.Errorf("Unconfirm tx:%s", in.Txid)
	}

	txsdk := ConvertTxToSDKTx(tx)

	return &pb.QueryTxResponse{
		Tx: txsdk,
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

	if nctx.ExceedDiskLimit() {
		return nil, ErrOutOfDiskLimit
	}
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

	limit := in.Cap
	if limit <= 0 {
		limit = DefaultCap
	}
	iter, err := nctx.Cache.Select(nctx.ContractName, in.Start, in.Limit)
	if err != nil {
		return nil, err
	}
	out := new(pb.IteratorResponse)
	for iter.Next() && limit > 0 {
		out.Items = append(out.Items, &pb.IteratorItem{
			Key:   iter.Key(),
			Value: iter.Data().GetPureData().GetValue(),
		})
		limit -= 1
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
	var args []*pb.ArgPair
	for key, value := range nctx.Args {
		args = append(args, &pb.ArgPair{
			Key:   key,
			Value: value,
		})
	}
	sort.Slice(args, func(i, j int) bool {
		return args[i].Key < args[j].Key
	})
	return &pb.CallArgs{
		Method:      nctx.Method,
		Args:        args,
		Initiator:   nctx.Initiator,
		AuthRequire: nctx.AuthRequire,
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

func ConvertTxToSDKTx(tx *xchainpb.Transaction) *pb.Transaction {
	txIns := []*pb.TxInput{}
	for _, in := range tx.TxInputs {
		txIn := &pb.TxInput{
			RefTxid:      hex.EncodeToString(in.RefTxid),
			RefOffset:    in.RefOffset,
			FromAddr:     in.FromAddr,
			Amount:       AmountBytesToString(in.Amount),
			FrozenHeight: in.FrozenHeight,
		}
		txIns = append(txIns, txIn)
	}

	txOuts := []*pb.TxOutput{}
	for _, out := range tx.TxOutputs {
		txOut := &pb.TxOutput{
			Amount:       AmountBytesToString(out.Amount),
			ToAddr:       out.ToAddr,
			FrozenHeight: out.FrozenHeight,
		}
		txOuts = append(txOuts, txOut)
	}

	txsdk := &pb.Transaction{
		Txid:        hex.EncodeToString(tx.Txid),
		Blockid:     hex.EncodeToString(tx.Blockid),
		TxInputs:    txIns,
		TxOutputs:   txOuts,
		Desc:        tx.Desc,
		Initiator:   tx.Initiator,
		AuthRequire: tx.AuthRequire,
	}

	return txsdk
}

func AmountBytesToString(buf []byte) string {
	n := new(big.Int)
	n.SetBytes(buf)
	return n.String()
}
