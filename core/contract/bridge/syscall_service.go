// Copyright (c) 2019, Baidu.com, Inc. All Rights Reserved.

package bridge

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/xuperchain/xuperchain/core/contract"
	pb "github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	xchainpb "github.com/xuperchain/xuperchain/core/pb"
)

var (
	ErrOutOfDiskLimit = errors.New("out of disk limit")
)

const (
	DefaultCap           = 1000
	MaxContractCallDepth = 10
)

type VmManager interface {
	GetVirtualMachine(name string) (contract.VirtualMachine, bool)
}

// SyscallService is the handler of contract syscalls
type SyscallService struct {
	ctxmgr *ContextManager
	vmm    VmManager
}

// NewSyscallService instances a new SyscallService
func NewSyscallService(ctxmgr *ContextManager, vmm VmManager) *SyscallService {
	return &SyscallService{
		ctxmgr: ctxmgr,
		vmm:    vmm,
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
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}
	amount, ok := new(big.Int).SetString(in.GetAmount(), 10)
	if !ok {
		return nil, errors.New("parse amount error")
	}
	if in.GetTo() == "" {
		return nil, errors.New("empty to address")
	}
	err := nctx.Cache.Transfer(nctx.ContractName, in.GetTo(), amount)
	if err != nil {
		return nil, err
	}
	resp := &pb.TransferResponse{}
	return resp, nil
}

// ContractCall implements Syscall interface
func (c *SyscallService) ContractCall(ctx context.Context, in *pb.ContractCallRequest) (*pb.ContractCallResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}
	if nctx.ContractSet[in.GetContract()] {
		return nil, errors.New("recursive contract call not permitted")
	}

	if len(nctx.ContractSet) >= MaxContractCallDepth {
		return nil, errors.New("max contract call depth exceeds")
	}

	ok, err := nctx.Core.VerifyContractPermission(nctx.Initiator, nctx.AuthRequire, in.GetContract(), in.GetMethod())
	if !ok || err != nil {
		return nil, errors.New("verify contract permission failed")
	}

	vm, ok := c.vmm.GetVirtualMachine(in.GetModule())
	if !ok {
		return nil, errors.New("module not found")
	}
	currentUsed := nctx.ResourceUsed()
	limits := new(contract.Limits).Add(nctx.ResourceLimits).Sub(currentUsed)
	// disk usage is shared between all context
	limits.Disk = nctx.ResourceLimits.Disk

	args := make(map[string][]byte)
	for _, arg := range in.GetArgs() {
		args[arg.GetKey()] = arg.GetValue()
	}

	nctx.ContractSet[in.GetContract()] = true
	cfg := &contract.ContextConfig{
		ContractName:   in.GetContract(),
		XMCache:        nctx.Cache,
		CanInitialize:  false,
		AuthRequire:    nctx.AuthRequire,
		Initiator:      nctx.Initiator,
		Core:           nctx.Core,
		ResourceLimits: *limits,
		ContractSet:    nctx.ContractSet,
	}
	vctx, err := vm.NewContext(cfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		vctx.Release()
		delete(nctx.ContractSet, in.GetContract())
	}()

	vresp, err := vctx.Invoke(in.GetMethod(), args)
	if err != nil {
		return nil, err
	}
	nctx.SubResourceUsed.Add(vctx.ResourceUsed())

	return &pb.ContractCallResponse{
		Response: &pb.Response{
			Status:  int32(vresp.Status),
			Message: vresp.Message,
			Body:    vresp.Body,
		}}, nil
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
		nctx.Instance.Abort(ErrOutOfDiskLimit.Error())
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
			Key:   append([]byte(""), iter.Data().GetPureData().GetKey()...), //make a copy
			Value: append([]byte(""), iter.Data().GetPureData().GetValue()...),
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
		Method:         nctx.Method,
		Args:           args,
		Initiator:      nctx.Initiator,
		AuthRequire:    nctx.AuthRequire,
		TransferAmount: nctx.TransferAmount,
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

func (c *SyscallService) GetAccountAddresses(ctx context.Context, in *pb.GetAccountAddressesRequest) (*pb.GetAccountAddressesResponse, error) {
	nctx, ok := c.ctxmgr.Context(in.GetHeader().Ctxid)
	if !ok {
		return nil, fmt.Errorf("bad ctx id:%d", in.Header.Ctxid)
	}
	addresses, err := nctx.Core.GetAccountAddresses(in.GetAccount())
	if err != nil {
		return nil, err
	}
	return &pb.GetAccountAddressesResponse{
		Addresses: addresses,
	}, nil
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
