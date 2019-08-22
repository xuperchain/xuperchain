package xmodel

import (
	"fmt"
	"sync"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
	xmodel_pb "github.com/xuperchain/xuperunion/xmodel/pb"
)

// XModel xmodel data structure
type XModel struct {
	bcname          string
	ledger          *ledger.Ledger
	stateDB         kvdb.Database
	unconfirmTable  kvdb.Database
	extUtxoTable    kvdb.Database
	extUtxoDelTable kvdb.Database
	logger          log.Logger
	batchCache      *sync.Map
	lastBatch       kvdb.Batch
}

// NewXuperModel new an instance of XModel
func NewXuperModel(bcname string, ledger *ledger.Ledger, stateDB kvdb.Database, logger log.Logger) (*XModel, error) {
	return &XModel{
		bcname:          bcname,
		ledger:          ledger,
		stateDB:         stateDB,
		unconfirmTable:  kvdb.NewTable(stateDB, pb.UnconfirmedTablePrefix),
		extUtxoTable:    kvdb.NewTable(stateDB, pb.ExtUtxoTablePrefix),
		extUtxoDelTable: kvdb.NewTable(stateDB, pb.ExtUtxoDelTablePrefix),
		logger:          logger,
		batchCache:      &sync.Map{},
	}, nil
}

func (s *XModel) updateExtUtxo(tx *pb.Transaction, batch kvdb.Batch) error {
	for offset, txOut := range tx.TxOutputsExt {
		bucketAndKey := makeRawKey(txOut.Bucket, txOut.Key)
		valueVersion := MakeVersion(tx.Txid, int32(offset))
		if isDelFlag(txOut.Value) {
			putKey := append([]byte(pb.ExtUtxoDelTablePrefix), bucketAndKey...)
			delKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
			batch.Delete(delKey)
			batch.Put(putKey, []byte(valueVersion))
			s.logger.Trace("    xmodel put gc", "putkey", string(putKey), "version", valueVersion)
			s.logger.Trace("    xmodel del", "delkey", string(delKey), "version", valueVersion)
		} else {
			putKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
			batch.Put(putKey, []byte(valueVersion))
			s.logger.Trace("    xmodel put", "putkey", string(putKey), "version", valueVersion)
		}
		s.batchCache.Store(string(bucketAndKey), valueVersion)
	}
	return nil
}

// DoTx running a transaction and update extUtxoTable
func (s *XModel) DoTx(tx *pb.Transaction, batch kvdb.Batch) error {
	s.cleanCache(batch)
	err := s.verifyInputs(tx)
	if err != nil {
		return err
	}
	err = s.verifyOutputs(tx)
	if err != nil {
		return err
	}
	err = s.updateExtUtxo(tx, batch)
	if err != nil {
		return err
	}
	return nil
}

// UndoTx rollback a transaction and update extUtxoTable
func (s *XModel) UndoTx(tx *pb.Transaction, batch kvdb.Batch) error {
	s.cleanCache(batch)
	inputVersionMap := map[string]string{}
	for _, txIn := range tx.TxInputsExt {
		rawKey := string(makeRawKey(txIn.Bucket, txIn.Key))
		version := GetVersionOfTxInput(txIn)
		inputVersionMap[rawKey] = version
	}
	for _, txOut := range tx.TxOutputsExt {
		bucketAndKey := makeRawKey(txOut.Bucket, txOut.Key)
		previousVersion := inputVersionMap[string(bucketAndKey)]
		if previousVersion == "" {
			delKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
			batch.Delete(delKey)
			s.logger.Trace("    undo xmodel del", "delkey", string(delKey))
			s.batchCache.Store(string(bucketAndKey), "")
		} else {
			verData, err := s.fetchVersionedData(previousVersion)
			if err != nil {
				return err
			}
			if isDelFlag(verData.PureData.Value) {
				putKey := append([]byte(pb.ExtUtxoDelTablePrefix), bucketAndKey...)
				batch.Put(putKey, verData.PureData.Value)
				delKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
				batch.Delete(delKey)
				s.logger.Trace("    undo xmodel put gc", "putkey", string(putKey), "prever", previousVersion)
				s.logger.Trace("    undo xmodel del", "del key", string(delKey), "prever", previousVersion)
			} else {
				putKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
				batch.Put(putKey, []byte(previousVersion))
				s.logger.Trace("    undo xmodel put", "putkey", string(putKey), "prever", previousVersion)
			}
			s.batchCache.Store(string(bucketAndKey), previousVersion)
		}
	}
	return nil
}

func (s *XModel) fetchVersionedData(version string) (*xmodel_pb.VersionedData, error) {
	txid, offset, err := parseVersion(version)
	if err != nil {
		return nil, err
	}
	tx, confirmed, err := s.queryTx(txid)
	if err != nil {
		return nil, err
	}
	if offset >= len(tx.TxOutputsExt) {
		return nil, fmt.Errorf("xmodel.Get failed, offset overflow: %d, %d", offset, len(tx.TxOutputsExt))
	}
	txOutputs := tx.TxOutputsExt[offset]
	return &xmodel_pb.VersionedData{
		RefTxid:   txid,
		RefOffset: int32(offset),
		PureData: &xmodel_pb.PureData{
			Key:    txOutputs.Key,
			Value:  txOutputs.Value,
			Bucket: txOutputs.Bucket,
		},
		Confirmed: confirmed,
	}, nil
}

// Get get value for specific key, return value with version
func (s *XModel) Get(bucket string, key []byte) (*xmodel_pb.VersionedData, error) {
	rawKey := makeRawKey(bucket, key)
	cacheObj, cacheHit := s.batchCache.Load(string(rawKey))
	if cacheHit {
		version := cacheObj.(string)
		if version == "" {
			return s.makeEmptyVersionedData(bucket, key), nil
		}
		return s.fetchVersionedData(version)
	}
	version, err := s.extUtxoTable.Get(rawKey)
	if err != nil {
		if kvdb.ErrNotFound(err) {
			//从回收站Get, 因为这个utxo可能是被删除了，RefTxid需要引用
			version, err = s.extUtxoDelTable.Get(rawKey)
			if err != nil {
				if kvdb.ErrNotFound(err) {
					return s.makeEmptyVersionedData(bucket, key), nil
				}
				return nil, err
			}
			return s.fetchVersionedData(string(version))
		}
		return nil, err
	}
	return s.fetchVersionedData(string(version))
}

// Select select all kv from a bucket, can set key range, left closed, right opend
func (s *XModel) Select(bucket string, startKey []byte, endKey []byte) (Iterator, error) {
	rawStartKey := makeRawKey(bucket, startKey)
	rawEndKey := makeRawKey(bucket, endKey)
	iter := &XMIterator{
		iter:  s.extUtxoTable.NewIteratorWithRange(rawStartKey, rawEndKey),
		model: s,
	}
	return iter, nil
}

func (s *XModel) queryTx(txid []byte) (*pb.Transaction, bool, error) {
	unconfirmTx, err := queryUnconfirmTx(txid, s.unconfirmTable)
	if err != nil {
		if !kvdb.ErrNotFound(err) {
			return nil, false, err
		}
	} else {
		return unconfirmTx, false, nil
	}
	confirmedTx, err := s.ledger.QueryTransaction(txid)
	if err != nil {
		return nil, false, err
	}
	return confirmedTx, true, nil

}

// QueryTx query transaction including unconfirmed table and confirmed table
func (s *XModel) QueryTx(txid []byte) (*pb.TxStatus, error) {
	tx, isConfirmed, err := s.queryTx(txid)
	if err != nil {
		return &pb.TxStatus{Tx: nil, Status: pb.TransactionStatus_NOEXIST}, err
	}
	status := pb.TransactionStatus_UNCONFIRM
	if isConfirmed {
		status = pb.TransactionStatus_CONFIRM
	} else {
		//notice: can not access the unconfirmed tx in smart contract
		tx = nil
	}
	return &pb.TxStatus{Tx: tx, Status: status}, nil
}

// QueryBlock query block from ledger
func (s *XModel) QueryBlock(blockid []byte) (*pb.InternalBlock, error) {
	block, err := s.ledger.QueryBlock(blockid)
	if err != nil {
		return nil, err
	}
	return block, nil
}

// CleanCache clear batchCache and lastBatch
func (s *XModel) CleanCache() {
	s.cleanCache(nil)
}

func (s *XModel) cleanCache(newBatch kvdb.Batch) {
	if newBatch != s.lastBatch {
		s.batchCache = &sync.Map{}
		s.lastBatch = newBatch
	}
}
