package xmodel

import (
	"bytes"
	"encoding/hex"
	"fmt"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	xmodpb "github.com/xuperchain/xuperchain/core/xmodel/pb"
)

type xModSnapshot struct {
	xmod      *XModel
	logger    log.Logger
	blkHeight int64
	blkId     []byte
}

type xModListCursor struct {
	txid   []byte
	offset int32
}

func NewSnapshot(blkId []byte, ledger *ledger.Ledger, stateDB kvdb.Database,
	logger log.Logger) (*xModSnapshot, error) {

	// check param
	if len(blkId) < 1 || ledger == nil || stateDB == nil || logger == nil {
		return nil, fmt.Errorf("param set error")
	}

	// 查询快照区块高度
	blkInfo, err := ledger.QueryBlockHeader(blkId)
	if err != nil {
		return nil, fmt.Errorf("query block header fail.err:%v", err)
	}

	// 实例化xmodel
	xmod, err := NewXuperModel(ledger, stateDB, logger)
	if err != nil {
		return nil, fmt.Errorf("new xmodel fail.err:%v", err)
	}

	xmsObj := &xModSnapshot{
		xmod:      xmod,
		logger:    logger,
		blkHeight: blkInfo.Height,
		blkId:     blkId,
	}

	return xmsObj, nil
}

func (t *xModSnapshot) Get(bucket string, key []byte) (*xmodpb.VersionedData, error) {
	if !t.isInit() || bucket == "" || len(key) < 1 {
		return nil, fmt.Errorf("xmod snapshot not init or param set error")
	}

	// 通过xmodel.Get()获取到最新版本数据
	newestVD, err := t.xmod.Get(bucket, key)
	if err != nil {
		return nil, fmt.Errorf("get newest version data fail.err:%v", err)
	}

	// 通过txid串联查询，直到找到<=blkHeight的交易
	var verValue *xmodpb.VersionedData
	cursor := &xModListCursor{newestVD.RefTxid, newestVD.RefOffset}
	for {
		// 最初的InputExt是空值，只设置了Bucket和Key
		if len(cursor.txid) < 1 {
			break
		}

		// 通过txid查询交易信息
		txInfo, _, err := t.xmod.QueryTx(cursor.txid)
		if err != nil {
			return nil, fmt.Errorf("query tx fail.err:%v", err)
		}
		// 更新游标，input和output的索引没有关系
		tmpOffset := cursor.offset
		cursor.txid, cursor.offset, err = t.getPreOutExt(txInfo.TxInputsExt, bucket, key)
		if err != nil {
			return nil, fmt.Errorf("get previous output fail.err:%v", err)
		}
		if txInfo.Blockid == nil {
			// 没有Blockid就是未确认交易，未确认交易直接更新游标
			continue
		}

		// 查询交易所在区块高度
		blkHeight, err := t.getBlockHeight(txInfo.Blockid)
		if err != nil {
			return nil, fmt.Errorf("query block height fail.err:%v", err)
		}
		// 当前块高度<=blkHeight，遍历结束
		if blkHeight <= t.blkHeight {
			verValue = t.genVerDataByTx(txInfo, tmpOffset)
			break
		}
	}

	if verValue == nil {
		return makeEmptyVersionedData(bucket, key), nil
	}
	return verValue, nil
}

func (t *xModSnapshot) isInit() bool {
	if t.xmod == nil || t.logger == nil || len(t.blkId) < 1 || t.blkHeight < 0 {
		return false
	}

	return true
}

func (t *xModSnapshot) getBlockHeight(blockid []byte) (int64, error) {
	blkInfo, err := t.xmod.QueryBlock(blockid)
	if err != nil {
		return 0, fmt.Errorf("query block info fail. block_id:%s err:%v",
			hex.EncodeToString(blockid), err)
	}

	return blkInfo.Height, nil
}

func (t *xModSnapshot) genVerDataByTx(tx *pb.Transaction, offset int32) *xmodpb.VersionedData {
	if tx == nil || int(offset) >= len(tx.TxOutputsExt) || offset < 0 {
		return nil
	}

	txOutputsExt := tx.TxOutputsExt[offset]
	value := &xmodpb.VersionedData{
		RefTxid:   tx.Txid,
		RefOffset: offset,
		PureData: &xmodpb.PureData{
			Key:    txOutputsExt.Key,
			Value:  txOutputsExt.Value,
			Bucket: txOutputsExt.Bucket,
		},
	}
	return value
}

// 根据bucket和key从inputsExt中查找对应的outputsExt索引
func (t *xModSnapshot) getPreOutExt(inputsExt []*pb.TxInputExt,
	bucket string, key []byte) ([]byte, int32, error) {
	for _, inExt := range inputsExt {
		if inExt.Bucket == bucket && bytes.Compare(inExt.Key, key) == 0 {
			return inExt.RefTxid, inExt.RefOffset, nil
		}
	}

	return nil, 0, fmt.Errorf("bucket and key not exist.bucket:%s key:%s", bucket, string(key))
}
