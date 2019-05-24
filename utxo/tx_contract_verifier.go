package utxo

import (
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	ledger_pkg "github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

// TxOfRunningContractVerify run contract and verify the tx
func (uv *UtxoVM) TxOfRunningContractVerify(batch kvdb.Batch, block *pb.InternalBlock, tx *pb.Transaction, autogenTxList *[]*pb.Transaction, idx int) (int, error) {
	if txDesc, ok := uv.isSmartContract(tx.Desc); ok {
		txDesc.Tx = tx
		err := uv.runContract(block.Blockid, tx, autogenTxList, 0)
		if block.Version >= ledger_pkg.BlockVersion {
			//进入v2才开始处理失败的情况
			minerErr := block.FailedTxs[global.F(tx.Txid)]
			if err != nil {
				uv.xlog.Warn("run contract failed, when handleContractForNonMiner", "err", err)
				if minerErr != "" { //验证节点和矿工都认为合约有错误
					uv.xlog.Warn("contranct also failed on minner, ignore it")
				} else {
					return idx, err
				}
			} else {
				if minerErr != "" {
					uv.xlog.Warn("local success, but miner mark it failed", "err", err)
				}
			}
		}
	}
	return idx + 1, nil
}

func (uv *UtxoVM) isConfirmed(tx *pb.Transaction) bool {
	b, err := uv.ledger.QueryBlockHeader(tx.Blockid)
	if err != nil {
		//查询不到直接放过
		return false
	}
	return b.InTrunk
}
