package utxo

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	ledger_pkg "github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

var (
	// TxSizePercent max percent of txs' size in one block
	TxSizePercent = 0.8
)

// GetNewAccountResourceAmount get account for creating an account
func (uv *UtxoVM) GetNewAccountResourceAmount() (int64, error) {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.GetNewAccountResourceAmount(), nil
}

// LoadNewAccountResourceAmount load newAccountResourceAmount into memory
func (uv *UtxoVM) LoadNewAccountResourceAmount() (int64, error) {
	newAccountResourceAmountBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.NewAccountResourceAmountKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(newAccountResourceAmountBuf, utxoMeta)
		return utxoMeta.GetNewAccountResourceAmount(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		return uv.ledger.GetNewAccountResourceAmount(), nil
	}

	return int64(0), findErr
}

// UpdateNewAccountResourceAmount ...
func (uv *UtxoVM) UpdateNewAccountResourceAmount(newAccountResourceAmount int64, batch kvdb.Batch) error {
	tmpMeta := &pb.UtxoMeta{}
	newMeta := proto.Clone(tmpMeta).(*pb.UtxoMeta)
	newMeta.NewAccountResourceAmount = newAccountResourceAmount
	newAccountResourceAmountBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		uv.xlog.Warn("failed to marshal pb meta")
		return pbErr
	}
	err := batch.Put([]byte(pb.MetaTablePrefix+ledger_pkg.NewAccountResourceAmountKey), newAccountResourceAmountBuf)
	if err == nil {
		uv.xlog.Info("Update newAccountResourceAmount succeed")
	}
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	uv.metaTmp.NewAccountResourceAmount = newAccountResourceAmount
	return err
}

// GetMaxBlockSize get max block size effective in Utxo
func (uv *UtxoVM) GetMaxBlockSize() (int64, error) {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.GetMaxBlockSize(), nil
}

// LoadMaxBlockSize load maxBlockSize into memory
func (uv *UtxoVM) LoadMaxBlockSize() (int64, error) {
	maxBlockSizeBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.MaxBlockSizeKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(maxBlockSizeBuf, utxoMeta)
		return utxoMeta.GetMaxBlockSize(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		return uv.ledger.GetMaxBlockSize(), nil
	}

	return int64(0), findErr
}

func (uv *UtxoVM) MaxTxSizePerBlock() (int, error) {
	maxBlkSize, err := uv.GetMaxBlockSize()
	return int(float64(maxBlkSize) * TxSizePercent), err
}

func (uv *UtxoVM) UpdateMaxBlockSize(maxBlockSize int64, batch kvdb.Batch) error {
	tmpMeta := &pb.UtxoMeta{}
	newMeta := proto.Clone(tmpMeta).(*pb.UtxoMeta)
	newMeta.MaxBlockSize = maxBlockSize
	maxBlockSizeBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		uv.xlog.Warn("failed to marshal pb meta")
		return pbErr

	}
	err := batch.Put([]byte(pb.MetaTablePrefix+ledger_pkg.MaxBlockSizeKey), maxBlockSizeBuf)
	if err == nil {
		uv.xlog.Info("Update maxBlockSize succeed")
	}
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	uv.metaTmp.MaxBlockSize = maxBlockSize
	return err
}

func (uv *UtxoVM) GetReservedContracts() ([]*pb.InvokeRequest, error) {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.ReservedContracts, nil
}

func (uv *UtxoVM) LoadReservedContracts() ([]*pb.InvokeRequest, error) {
	reservedContractsBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.ReservedContractsKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(reservedContractsBuf, utxoMeta)
		return utxoMeta.GetReservedContracts(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		return uv.ledger.GetReservedContracts()
	}
	return nil, findErr
}

func (uv *UtxoVM) UpdateReservedContracts(params []*pb.InvokeRequest, batch kvdb.Batch) error {
	if params == nil {
		return fmt.Errorf("invalid reservered contract requests")
	}
	tmpNewMeta := &pb.UtxoMeta{}
	newMeta := proto.Clone(tmpNewMeta).(*pb.UtxoMeta)
	newMeta.ReservedContracts = params
	paramsBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		uv.xlog.Warn("failed to marshal pb meta")
		return pbErr
	}
	err := batch.Put([]byte(pb.MetaTablePrefix+ledger_pkg.ReservedContractsKey), paramsBuf)
	if err == nil {
		uv.xlog.Info("Update reservered contract succeed")
	}
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	uv.metaTmp.ReservedContracts = params
	return err
}

func (uv *UtxoVM) GetForbiddenContract() (*pb.InvokeRequest, error) {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.GetForbiddenContract(), nil
}

func (uv *UtxoVM) LoadForbiddenContract() (*pb.InvokeRequest, error) {
	forbiddenContractBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.ForbiddenContractKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(forbiddenContractBuf, utxoMeta)
		return utxoMeta.GetForbiddenContract(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		requests, err := uv.ledger.GetForbiddenContract()
		if len(requests) > 0 {
			return requests[0], err
		}
		return nil, errors.New("unexpected error")
	}
	return nil, findErr
}

func (uv *UtxoVM) UpdateForbiddenContract(param *pb.InvokeRequest, batch kvdb.Batch) error {
	if param == nil {
		return fmt.Errorf("invalid forbidden contract request")
	}
	tmpNewMeta := &pb.UtxoMeta{}
	newMeta := proto.Clone(tmpNewMeta).(*pb.UtxoMeta)
	newMeta.ForbiddenContract = param
	paramBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		uv.xlog.Warn("failed to marshal pb meta")
		return pbErr
	}
	err := batch.Put([]byte(pb.MetaTablePrefix+ledger_pkg.ForbiddenContractKey), paramBuf)
	if err == nil {
		uv.xlog.Info("Update forbidden contract succeed")
	}
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	uv.metaTmp.ForbiddenContract = param
	return err
}

func (uv *UtxoVM) LoadIrreversibleBlockHeight() (int64, error) {
	irreversibleBlockHeightBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.IrreversibleBlockHeightKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(irreversibleBlockHeightBuf, utxoMeta)
		return utxoMeta.GetIrreversibleBlockHeight(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		return int64(0), nil
	}
	return int64(0), findErr
}

func (uv *UtxoVM) LoadIrreversibleSlideWindow() (int64, error) {
	irreversibleSlideWindowBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.IrreversibleSlideWindowKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(irreversibleSlideWindowBuf, utxoMeta)
		return utxoMeta.GetIrreversibleSlideWindow(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		return uv.ledger.GetIrreversibleSlideWindow(), nil
	}
	return int64(0), findErr
}

func (uv *UtxoVM) GetIrreversibleBlockHeight() int64 {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.IrreversibleBlockHeight
}

func (uv *UtxoVM) GetIrreversibleSlideWindow() int64 {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.IrreversibleSlideWindow
}

func (uv *UtxoVM) UpdateIrreversibleBlockHeight(nextIrreversibleBlockHeight int64, batch kvdb.Batch) error {
	tmpMeta := &pb.UtxoMeta{}
	newMeta := proto.Clone(tmpMeta).(*pb.UtxoMeta)
	newMeta.IrreversibleBlockHeight = nextIrreversibleBlockHeight
	irreversibleBlockHeightBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		uv.xlog.Warn("failed to marshal pb meta")
		return pbErr
	}
	err := batch.Put([]byte(pb.MetaTablePrefix+ledger_pkg.IrreversibleBlockHeightKey), irreversibleBlockHeightBuf)
	if err != nil {
		return err
	}
	uv.xlog.Info("Update irreversibleBlockHeight succeed")
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	uv.metaTmp.IrreversibleBlockHeight = nextIrreversibleBlockHeight
	return nil
}

func (uv *UtxoVM) updateNextIrreversibleBlockHeight(blockHeight int64, curIrreversibleBlockHeight int64, curIrreversibleSlideWindow int64, batch kvdb.Batch) error {
	if curIrreversibleSlideWindow <= 0 {
		return nil
	}
	nextIrreversibleBlockHeight := blockHeight - curIrreversibleSlideWindow
	// case1: slideWindow不变或变小
	if nextIrreversibleBlockHeight >= 0 {
		err := uv.UpdateIrreversibleBlockHeight(nextIrreversibleBlockHeight, batch)
		return err
	}
	// case2: slide变大或区块发生回滚
	if curIrreversibleBlockHeight < 0 {
		uv.xlog.Warn("update irreversible block height error, should be here")
		return errors.New("curIrreversibleBlockHeight is less than 0")
	}
	return nil
}
