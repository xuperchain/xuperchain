package xmodel

import (
	"fmt"

	"github.com/xuperchain/xuperchain/core/pb"
)

func (s *XModel) verifyInputs(tx *pb.Transaction) error {
	//确保tx.TxInputs里面声明的版本和本地model是match的
	for _, txIn := range tx.TxInputsExt {
		verData, err := s.GetUncommited(txIn.Bucket, txIn.Key) //because previous txs in the same block write into batch cache
		if err != nil {
			return err
		}
		localVer := GetVersion(verData)
		remoteVer := GetVersionOfTxInput(txIn)
		if localVer != remoteVer {
			return fmt.Errorf("verifyInputs failed, version missmatch: %s / %s, local: %s, remote:%s",
				txIn.Bucket, txIn.Key,
				localVer, remoteVer)
		}
	}
	return nil
}

func (s *XModel) verifyOutputs(tx *pb.Transaction) error {
	//outputs中不能出现inputs没有的key
	inputKeys := map[string]bool{}
	for _, txIn := range tx.TxInputsExt {
		rawKey := string(makeRawKey(txIn.Bucket, txIn.Key))
		inputKeys[rawKey] = true
	}
	for _, txOut := range tx.TxOutputsExt {
		if txOut.Bucket == TransientBucket {
			continue
		}
		rawKey := string(makeRawKey(txOut.Bucket, txOut.Key))
		if !inputKeys[rawKey] {
			return fmt.Errorf("verifyOutputs failed, not such key in txInputsExt: %s", rawKey)
		}
		if txOut.Value == nil {
			return fmt.Errorf("verifyOutputs failed, value can't be null")
		}
	}
	return nil
}
