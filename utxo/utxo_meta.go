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
