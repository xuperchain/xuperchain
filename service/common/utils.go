package common

import (
	"fmt"

	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo/txhash"
	crypto_base "github.com/xuperchain/xupercore/lib/crypto/client/base"
)

// 适配原结构计算txid
func MakeTxId(tx *pb.Transaction) ([]byte, error) {
	// 转化结构
	xldgTx := TxToXledger(tx)
	if xldgTx == nil {
		return nil, fmt.Errorf("tx convert fail")
	}
	// 计算txid
	txId, err := txhash.MakeTransactionID(xldgTx)
	if err != nil {
		return nil, err
	}
	return txId, nil
}

// 适配原结构签名
func ComputeTxSign(cryptoClient crypto_base.CryptoClient, tx *pb.Transaction, jsonSK []byte) ([]byte, error) {
	// 转换结构
	xldgTx := TxToXledger(tx)
	if xldgTx == nil {
		return nil, fmt.Errorf("tx convert fail")
	}
	txSign, err := txhash.ProcessSignTx(cryptoClient, xldgTx, jsonSK)
	if err != nil {
		return nil, err
	}
	return txSign, nil
}

func MakeTxDigestHash(tx *pb.Transaction) ([]byte, error) {
	// 转换结构
	xldgTx := TxToXledger(tx)
	if xldgTx == nil {
		return nil, fmt.Errorf("tx convert fail")
	}
	digestHash, err := txhash.MakeTxDigestHash(xldgTx)
	if err != nil {
		return nil, err
	}
	return digestHash, nil
}
