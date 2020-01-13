package xmodel

import (
	"errors"

	"github.com/xuperchain/xuperchain/core/pb"
)

// GetNearestConfirmedValue get the confirmed value of a specific key
// step1: get the value rewritten by the tx executed recently and successfully
// step2: check if the tx has been confirmed, and if not, check the reftix recursively
func (s *XModel) GetNearestConfirmedValue(bucket string, key []byte) ([]byte, bool, error) {
	versionData, confirmed, err := s.GetWithTxStatus(bucket, key)
	if err != nil {
		return nil, false, err
	}
	// 从xmodel拿到的数据直接已经confirmed, 那么直接返回
	if confirmed {
		return versionData.GetPureData().GetValue(), confirmed, nil
	}
	// 从它引用的reftxid获取已经confirmed的value
	refTxid := versionData.GetRefTxid()
	for {
		// no confirmed value and mission failed
		if refTxid == nil {
			return nil, false, errors.New("no confirmed value")
		}
		// s.queryTx会从confirmed/unconfirmed中查询tx
		tx, confirmed, err := s.queryTx(refTxid)
		if err != nil {
			return nil, false, err
		}
		// 被引用的tx还未确认
		if !confirmed {
			refTxid = getRefTxidFromTxWithBucketAndKey(tx, bucket, key)
			continue
		}
		// 被引用的tx已经被确认
		// 直接拿到被引用tx的写集
		value := getWriteSetFromTxWithBucketAndKey(tx, bucket, key)
		return value, true, nil
	}
}

// getRefTxidFromTxWithBucketAndKey 从tx中获取包含特定bucket,key引用的txid
func getRefTxidFromTxWithBucketAndKey(tx *pb.Transaction, bucket string, key []byte) []byte {
	for _, inputsExt := range tx.GetTxInputsExt() {
		if inputsExt.GetBucket() == bucket && string(inputsExt.GetKey()) == string(key) {
			return inputsExt.GetRefTxid()
		}
	}
	// 读集中不包含对应的bucket和key
	return nil
}

// getWriteSetFromTxWithBucketAndKey 从tx中获取包含特定bucket,key的写集
func getWriteSetFromTxWithBucketAndKey(tx *pb.Transaction, bucket string, key []byte) []byte {
	for _, outputsExt := range tx.GetTxOutputsExt() {
		if outputsExt.GetBucket() == bucket && string(outputsExt.GetKey()) == string(key) {
			return outputsExt.GetValue()
		}
	}
	// 没有获取到value
	return nil
}
