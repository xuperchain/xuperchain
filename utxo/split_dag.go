package utxo

import (
	"github.com/xuperchain/xuperunion/pb"
)

func (uv *UtxoVM) buildTxDeps(txs []*pb.Transaction) (map[string]*pb.Transaction, TxGraph, error) {
	txMap := map[string]*pb.Transaction{}
	txGraph := TxGraph{}
	for _, tx := range txs {
		txid := string(tx.Txid)
		txMap[txid] = tx
		txGraph[txid] = []string{}
	}

	for txid, tx := range txMap {
		// 填写正常的utxo的引用
		for _, refTx := range tx.TxInputs {
			refTxid := string(refTx.RefTxid)
			// 依赖的交易不在同一个区块中
			if _, exist := txMap[refTxid]; !exist {
				continue
			}
			txGraph[refTxid] = append(txGraph[refTxid], txid)
		}
		// 填写读写集
		for _, txIn := range tx.TxInputsExt {
			refTxid := string(txIn.RefTxid)
			// 依赖的交易不在同一个区块中
			if _, exist := txMap[refTxid]; !exist {
				continue
			}
			txGraph[refTxid] = append(txGraph[refTxid], txid)
		}
	}
	return txMap, txGraph, nil
}
