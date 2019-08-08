package utxo

import (
	"fmt"

	"github.com/xuperchain/xuperunion/pb"
)

func (uv *UtxoVM) buildTxDeps(txs []*pb.Transaction) (map[string]*pb.Transaction, TxGraph, error) {
	txMap := map[string]*pb.Transaction{}
	txGraph := TxGraph{}
	for _, tx := range txs {
		txid := fmt.Sprintf("%x", tx.Txid)
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

// SplitChildDAGs split child dags
func SplitChildDAGs(txGraph TxGraph, outputTxList []string) []int {
	step := 0
	idx := 0
	// 存放每个子DAG的大小,作为拆分子DAG的索引
	res := []int{}
	size := len(outputTxList)
	for idx < size {
		step = 0
		// 不被别人依赖
		if len(txGraph[outputTxList[idx]]) <= 0 {
			step = 1
		} else if len(txGraph[outputTxList[idx]]) > 0 {
			// 被别人依赖
			tmp := SplitChildDAGs(txGraph, txGraph[outputTxList[idx]])
			// 该子DAG的元素个数
			for _, v := range tmp {
				step += v
			}
		}
		// 当前子DAG的元素个数
		res = append(res, step)
		// 跳过已经经过拣选的元素
		idx += step
	}
	return res
}
