package utxo

import (
	"errors"

	"github.com/xuperchain/xuperunion/pb"
)

// GetConfirmedValue get the confirmed value of a specific key
// step1: get the value rewritten by the tx executed recently and successfully
// step2: check if the tx has been confirmed, and if not, check the reftix recursively
func (uv *UtxoVM) GetConfirmedValue(bucket string, key []byte) ([]byte, bool, error) {
	versionData, err := uv.model3.Get(bucket, key)
	if err != nil {
		return nil, false, err
	}
	confirmed := versionData.GetConfirmed()
	// 从xmodel拿到的数据直接已经confirmed, 那么直接返回
	if confirmed {
		return versionData.GetPureData().GetValue(), confirmed, nil
	}
	// 从它引用的reftxid获取已经confirmed的value
	refTxid := versionData.GetRefTxid()
	tx := &pb.Transaction{}
	for {
		// no confirmed value and mission failed
		if refTxid == nil {
			return nil, false, errors.New("no confirmed value")
		}
		// 因为HasTx永远不会返回error,这里就不判断error的返回值
		exist, _ := uv.HasTx(refTxid)
		// 被引用的tx还未确认
		if exist {
			tx, err = uv.QueryTx(refTxid)
			if err != nil {
				return nil, false, err
			}
			refTxid = getRefTxidFromTxWithBucketAndKey(tx, bucket, key)
			continue
		} else {
			// 被引用的tx已经被确认
			// 直接拿到被引用tx的写集
			tx, err = uv.ledger.QueryTransaction(refTxid)
			if err != nil {
				return nil, false, err
			}
			value := getWriteSetFromTxWithBucketAndKey(tx, bucket, key)
			return value, true, nil
		}
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
