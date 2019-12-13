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
	ErrProposalParamsIsNegativeNumber    = errors.New("negative number for proposal parameter is not allowed")
	ErrProposalParamsIsNotPositiveNumber = errors.New("negative number of zero for proposal parameter is not allowed")
	// TxSizePercent max percent of txs' size in one block
	TxSizePercent = 0.8
)

// GetNewAccountResourceAmount get account for creating an account
func (uv *UtxoVM) GetNewAccountResourceAmount() int64 {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.GetNewAccountResourceAmount()
}

// LoadNewAccountResourceAmount load newAccountResourceAmount into memory
func (uv *UtxoVM) LoadNewAccountResourceAmount() (int64, error) {
	newAccountResourceAmountBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.NewAccountResourceAmountKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(newAccountResourceAmountBuf, utxoMeta)
		return utxoMeta.GetNewAccountResourceAmount(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		genesisNewAccountResourceAmount := uv.ledger.GetNewAccountResourceAmount()
		if genesisNewAccountResourceAmount < 0 {
			return genesisNewAccountResourceAmount, ErrProposalParamsIsNegativeNumber
		}
		return genesisNewAccountResourceAmount, nil
	}

	return int64(0), findErr
}

// UpdateNewAccountResourceAmount ...
func (uv *UtxoVM) UpdateNewAccountResourceAmount(newAccountResourceAmount int64, batch kvdb.Batch) error {
	if newAccountResourceAmount < 0 {
		return ErrProposalParamsIsNegativeNumber
	}
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
func (uv *UtxoVM) GetMaxBlockSize() int64 {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.GetMaxBlockSize()
}

// LoadMaxBlockSize load maxBlockSize into memory
func (uv *UtxoVM) LoadMaxBlockSize() (int64, error) {
	maxBlockSizeBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.MaxBlockSizeKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(maxBlockSizeBuf, utxoMeta)
		return utxoMeta.GetMaxBlockSize(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		genesisMaxBlockSize := uv.ledger.GetMaxBlockSize()
		if genesisMaxBlockSize <= 0 {
			return genesisMaxBlockSize, ErrProposalParamsIsNotPositiveNumber
		}
		return genesisMaxBlockSize, nil
	}

	return int64(0), findErr
}

func (uv *UtxoVM) MaxTxSizePerBlock() (int, error) {
	maxBlkSize := uv.GetMaxBlockSize()
	return int(float64(maxBlkSize) * TxSizePercent), nil
}

func (uv *UtxoVM) UpdateMaxBlockSize(maxBlockSize int64, batch kvdb.Batch) error {
	if maxBlockSize <= 0 {
		return ErrProposalParamsIsNotPositiveNumber
	}
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

func (uv *UtxoVM) GetReservedContracts() []*pb.InvokeRequest {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.ReservedContracts
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

func (uv *UtxoVM) GetForbiddenContract() *pb.InvokeRequest {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.GetForbiddenContract()
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
		genesisSlideWindow := uv.ledger.GetIrreversibleSlideWindow()
		// negative number is not meaningful
		if genesisSlideWindow < 0 {
			return genesisSlideWindow, ErrProposalParamsIsNegativeNumber
		}
		return genesisSlideWindow, nil
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
	// negative number for irreversible slide window is not allowed.
	if curIrreversibleSlideWindow < 0 {
		return ErrProposalParamsIsNegativeNumber
	}
	// slideWindow为开启,不需要更新IrreversibleBlockHeight
	if curIrreversibleSlideWindow == 0 {
		return nil
	}
	// curIrreversibleBlockHeight小于0, 不符合预期，报警
	if curIrreversibleBlockHeight < 0 {
		uv.xlog.Warn("update irreversible block height error, should be here")
		return errors.New("curIrreversibleBlockHeight is less than 0")
	}
	nextIrreversibleBlockHeight := blockHeight - curIrreversibleSlideWindow
	// 下一个不可逆高度小于当前不可逆高度，直接返回
	// slideWindow变大或者发生区块回滚
	if nextIrreversibleBlockHeight <= curIrreversibleBlockHeight {
		return nil
	}
	// 正常升级
	// slideWindow不变或变小
	if nextIrreversibleBlockHeight > curIrreversibleBlockHeight {
		err := uv.UpdateIrreversibleBlockHeight(nextIrreversibleBlockHeight, batch)
		return err
	}

	return errors.New("unexpected error")
}

func (uv *UtxoVM) updateNextIrreversibleBlockHeightForPrune(blockHeight int64, curIrreversibleBlockHeight int64, curIrreversibleSlideWindow int64, batch kvdb.Batch) error {
	// negative number for irreversible slide window is not allowed.
	if curIrreversibleSlideWindow < 0 {
		return ErrProposalParamsIsNegativeNumber
	}
	// slideWindow为开启,不需要更新IrreversibleBlockHeight
	if curIrreversibleSlideWindow == 0 {
		return nil
	}
	// curIrreversibleBlockHeight小于0, 不符合预期，报警
	if curIrreversibleBlockHeight < 0 {
		uv.xlog.Warn("update irreversible block height error, should be here")
		return errors.New("curIrreversibleBlockHeight is less than 0")
	}
	nextIrreversibleBlockHeight := blockHeight - curIrreversibleSlideWindow
	if nextIrreversibleBlockHeight <= 0 {
		nextIrreversibleBlockHeight = 0
	}
	err := uv.UpdateIrreversibleBlockHeight(nextIrreversibleBlockHeight, batch)
	return err
}

func (uv *UtxoVM) UpdateIrreversibleSlideWindow(nextIrreversibleSlideWindow int64, batch kvdb.Batch) error {
	if nextIrreversibleSlideWindow < 0 {
		return ErrProposalParamsIsNegativeNumber
	}
	tmpMeta := &pb.UtxoMeta{}
	newMeta := proto.Clone(tmpMeta).(*pb.UtxoMeta)
	newMeta.IrreversibleSlideWindow = nextIrreversibleSlideWindow
	irreversibleSlideWindowBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		uv.xlog.Warn("failed to marshal pb meta")
		return pbErr
	}
	err := batch.Put([]byte(pb.MetaTablePrefix+ledger_pkg.IrreversibleSlideWindowKey), irreversibleSlideWindowBuf)
	if err != nil {
		return err
	}
	uv.xlog.Info("Update irreversibleSlideWindow succeed")
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	uv.metaTmp.IrreversibleSlideWindow = nextIrreversibleSlideWindow
	return nil
}

// GetGasPrice get gas rate to utxo
func (uv *UtxoVM) GetGasPrice() *pb.GasPrice {
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	return uv.meta.GetGasPrice()
}

// LoadGasPrice load gas rate
func (uv *UtxoVM) LoadGasPrice() (*pb.GasPrice, error) {
	gasPriceBuf, findErr := uv.metaTable.Get([]byte(ledger_pkg.GasPriceKey))
	if findErr == nil {
		utxoMeta := &pb.UtxoMeta{}
		err := proto.Unmarshal(gasPriceBuf, utxoMeta)
		return utxoMeta.GetGasPrice(), err
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		gasPrice := uv.ledger.GetGasPrice()
		cpuRate := gasPrice.CpuRate
		memRate := gasPrice.MemRate
		diskRate := gasPrice.DiskRate
		xfeeRate := gasPrice.XfeeRate
		if cpuRate < 0 || memRate < 0 || diskRate < 0 || xfeeRate < 0 {
			return nil, ErrProposalParamsIsNegativeNumber
		}
		// To be compatible with the old version v3.3
		// If GasPrice configuration is missing or value euqals 0, support a default value
		if cpuRate == 0 && memRate == 0 && diskRate == 0 && xfeeRate == 0 {
			gasPrice = &pb.GasPrice{
				CpuRate:  1000,
				MemRate:  1000000,
				DiskRate: 1,
				XfeeRate: 1,
			}
		}
		return gasPrice, nil
	}
	return nil, findErr
}

// UpdateGasPrice update gasPrice parameters
func (uv *UtxoVM) UpdateGasPrice(nextGasPrice *pb.GasPrice, batch kvdb.Batch) error {
	// check if the parameters are valid
	cpuRate := nextGasPrice.GetCpuRate()
	memRate := nextGasPrice.GetMemRate()
	diskRate := nextGasPrice.GetDiskRate()
	xfeeRate := nextGasPrice.GetXfeeRate()
	if cpuRate < 0 || memRate < 0 || diskRate < 0 || xfeeRate < 0 {
		return ErrProposalParamsIsNegativeNumber
	}
	tmpMeta := &pb.UtxoMeta{}
	newMeta := proto.Clone(tmpMeta).(*pb.UtxoMeta)
	newMeta.GasPrice = nextGasPrice
	gasPriceBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		uv.xlog.Warn("failed to marshal pb meta")
		return pbErr
	}
	err := batch.Put([]byte(pb.MetaTablePrefix+ledger_pkg.GasPriceKey), gasPriceBuf)
	if err != nil {
		return err
	}
	uv.xlog.Info("Update gas price succeed")
	uv.mutexMeta.Lock()
	defer uv.mutexMeta.Unlock()
	uv.metaTmp.GasPrice = nextGasPrice
	return nil
}
