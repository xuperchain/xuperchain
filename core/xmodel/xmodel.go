package xmodel

import (
	"encoding/hex"
	"fmt"
	"sync"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	xmodel_pb "github.com/xuperchain/xuperchain/core/xmodel/pb"
)

const (
	bucketExtUTXOCacheSize = 1024

	// TransientBucket is the name of bucket that only appears in tx output set
	// but does't persists in xmodel
	TransientBucket = "$transient"
)

// XModel xmodel data structure
type XModel struct {
	ledger          *ledger.Ledger
	stateDB         kvdb.Database
	unconfirmTable  kvdb.Database
	extUtxoTable    kvdb.Database
	extUtxoDelTable kvdb.Database
	logger          log.Logger
	batchCache      *sync.Map
	lastBatch       kvdb.Batch
	// extUtxoCache caches per bucket key-values using version as key
	extUtxoCache sync.Map // map[string]*LRUCache
}

// NewXuperModel new an instance of XModel
func NewXuperModel(ledger *ledger.Ledger, stateDB kvdb.Database, logger log.Logger) (*XModel, error) {
	return &XModel{
		ledger:          ledger,
		stateDB:         stateDB,
		unconfirmTable:  kvdb.NewTable(stateDB, pb.UnconfirmedTablePrefix),
		extUtxoTable:    kvdb.NewTable(stateDB, pb.ExtUtxoTablePrefix),
		extUtxoDelTable: kvdb.NewTable(stateDB, pb.ExtUtxoDelTablePrefix),
		logger:          logger,
		batchCache:      &sync.Map{},
	}, nil
}

func (s *XModel) CreateSnapshot(blkId []byte) (XMReader, error) {
	// 查询快照区块高度
	blkInfo, err := s.ledger.QueryBlockHeader(blkId)
	if err != nil {
		return nil, fmt.Errorf("query block header fail.block_id:%s, err:%v",
			hex.EncodeToString(blkId), err)
	}

	xms := &xModSnapshot{
		xmod:      s,
		logger:    s.logger,
		blkHeight: blkInfo.Height,
		blkId:     blkId,
	}
	return xms, nil
}

func (s *XModel) updateExtUtxo(tx *pb.Transaction, batch kvdb.Batch) error {
	for offset, txOut := range tx.TxOutputsExt {
		if txOut.Bucket == TransientBucket {
			continue
		}
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
		if len(tx.Blockid) > 0 {
			s.batchCache.Store(string(bucketAndKey), valueVersion)
		}
		s.bucketCacheStore(txOut.Bucket, valueVersion, &xmodel_pb.VersionedData{
			RefTxid:   tx.Txid,
			RefOffset: int32(offset),
			PureData: &xmodel_pb.PureData{
				Key:    txOut.Key,
				Value:  txOut.Value,
				Bucket: txOut.Bucket,
			},
		})
	}
	return nil
}

// DoTx running a transaction and update extUtxoTable
func (s *XModel) DoTx(tx *pb.Transaction, batch kvdb.Batch) error {
	if len(tx.Blockid) > 0 {
		s.cleanCache(batch)
	}
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
		if txOut.Bucket == TransientBucket {
			continue
		}
		bucketAndKey := makeRawKey(txOut.Bucket, txOut.Key)
		previousVersion := inputVersionMap[string(bucketAndKey)]
		if previousVersion == "" {
			delKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
			batch.Delete(delKey)
			s.logger.Trace("    undo xmodel del", "delkey", string(delKey))
			s.batchCache.Store(string(bucketAndKey), "")
		} else {
			verData, err := s.fetchVersionedData(txOut.Bucket, previousVersion)
			if err != nil {
				return err
			}
			if isDelFlag(verData.PureData.Value) { //previous version is del
				putKey := append([]byte(pb.ExtUtxoDelTablePrefix), bucketAndKey...)
				batch.Put(putKey, []byte(previousVersion))
				delKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
				batch.Delete(delKey)
				s.logger.Trace("    undo xmodel put gc", "putkey", string(putKey), "prever", previousVersion)
				s.logger.Trace("    undo xmodel del", "del key", string(delKey), "prever", previousVersion)
			} else {
				putKey := append([]byte(pb.ExtUtxoTablePrefix), bucketAndKey...)
				batch.Put(putKey, []byte(previousVersion))
				s.logger.Trace("    undo xmodel put", "putkey", string(putKey), "prever", previousVersion)
				if isDelFlag(txOut.Value) { //current version is del
					delKey := append([]byte(pb.ExtUtxoDelTablePrefix), bucketAndKey...)
					batch.Delete(delKey) //remove garbage in gc table
				}
			}
			s.batchCache.Store(string(bucketAndKey), previousVersion)
		}
	}
	return nil
}

func (s *XModel) fetchVersionedData(bucket, version string) (*xmodel_pb.VersionedData, error) {
	value, ok := s.bucketCacheGet(bucket, version)
	if ok {
		return value, nil
	}
	txid, offset, err := parseVersion(version)
	if err != nil {
		return nil, err
	}
	tx, _, err := s.queryTx(txid)
	if err != nil {
		return nil, err
	}
	if offset >= len(tx.TxOutputsExt) {
		return nil, fmt.Errorf("xmodel.Get failed, offset overflow: %d, %d", offset, len(tx.TxOutputsExt))
	}
	txOutputs := tx.TxOutputsExt[offset]
	value = &xmodel_pb.VersionedData{
		RefTxid:   txid,
		RefOffset: int32(offset),
		PureData: &xmodel_pb.PureData{
			Key:    txOutputs.Key,
			Value:  txOutputs.Value,
			Bucket: txOutputs.Bucket,
		},
	}
	s.bucketCacheStore(bucket, version, value)
	return value, nil
}

// GetUncommited get value for specific key, return the value with version, even it is in batch cache
func (s *XModel) GetUncommited(bucket string, key []byte) (*xmodel_pb.VersionedData, error) {
	rawKey := makeRawKey(bucket, key)
	cacheObj, cacheHit := s.batchCache.Load(string(rawKey))
	if cacheHit {
		version := cacheObj.(string)
		if version == "" {
			return makeEmptyVersionedData(bucket, key), nil
		}
		return s.fetchVersionedData(bucket, version)
	}
	return s.Get(bucket, key)
}

// GetFromLedger get data directely from ledger
func (s *XModel) GetFromLedger(txin *pb.TxInputExt) (*xmodel_pb.VersionedData, error) {
	if txin.RefTxid == nil {
		return makeEmptyVersionedData(txin.Bucket, txin.Key), nil
	}
	version := MakeVersion(txin.RefTxid, txin.RefOffset)
	return s.fetchVersionedData(txin.Bucket, version)
}

// Get get value for specific key, return value with version
func (s *XModel) Get(bucket string, key []byte) (*xmodel_pb.VersionedData, error) {
	rawKey := makeRawKey(bucket, key)
	version, err := s.extUtxoTable.Get(rawKey)
	if err != nil {
		if kvdb.ErrNotFound(err) {
			//从回收站Get, 因为这个utxo可能是被删除了，RefTxid需要引用
			version, err = s.extUtxoDelTable.Get(rawKey)
			if err != nil {
				if kvdb.ErrNotFound(err) {
					return makeEmptyVersionedData(bucket, key), nil
				}
				return nil, err
			}
			return s.fetchVersionedData(bucket, string(version))
		}
		return nil, err
	}
	return s.fetchVersionedData(bucket, string(version))
}

// GetWithTxStatus likes Get but also return tx status information
func (s *XModel) GetWithTxStatus(bucket string, key []byte) (*xmodel_pb.VersionedData, bool, error) {
	data, err := s.Get(bucket, key)
	if err != nil {
		return nil, false, err
	}
	exists, err := s.ledger.HasTransaction(data.RefTxid)
	if err != nil {
		return nil, false, err
	}
	return data, exists, nil
}

// Select select all kv from a bucket, can set key range, left closed, right opend
func (s *XModel) Select(bucket string, startKey []byte, endKey []byte) (Iterator, error) {
	rawStartKey := makeRawKey(bucket, startKey)
	rawEndKey := makeRawKey(bucket, endKey)
	iter := &XMIterator{
		bucket: bucket,
		iter:   s.extUtxoTable.NewIteratorWithRange(rawStartKey, rawEndKey),
		model:  s,
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
func (s *XModel) QueryTx(txid []byte) (*pb.Transaction, bool, error) {
	tx, status, err := s.queryTx(txid)
	if err != nil {
		return nil, status, err
	}
	return tx, status, nil
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

func (s *XModel) bucketCache(bucket string) *common.LRUCache {
	icache, ok := s.extUtxoCache.Load(bucket)
	if ok {
		return icache.(*common.LRUCache)
	}
	cache := common.NewLRUCache(bucketExtUTXOCacheSize)
	s.extUtxoCache.Store(bucket, cache)
	return cache
}

func (s *XModel) bucketCacheStore(bucket, version string, value *xmodel_pb.VersionedData) {
	cache := s.bucketCache(bucket)
	cache.Add(version, value)
}

func (s *XModel) bucketCacheGet(bucket, version string) (*xmodel_pb.VersionedData, bool) {
	cache := s.bucketCache(bucket)
	value, ok := cache.Get(version)
	if !ok {
		return nil, false
	}
	return value.(*xmodel_pb.VersionedData), true
}

// BucketCacheDelete gen write key with perfix
func (s *XModel) BucketCacheDelete(bucket, version string) {
	cache := s.bucketCache(bucket)
	cache.Del(version)
}

// GenWriteKeyWithPrefix gen write key with perfix
func GenWriteKeyWithPrefix(txOutputExt *pb.TxOutputExt) string {
	bucket := txOutputExt.GetBucket()
	key := txOutputExt.GetKey()
	baseWriteSetKey := bucket + fmt.Sprintf("%s", key)
	return pb.ExtUtxoTablePrefix + baseWriteSetKey
}
