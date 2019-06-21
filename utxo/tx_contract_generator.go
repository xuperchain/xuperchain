package utxo

// generator/verifier 主要操作合约的交易， 包括：
//     1. 执行和验证合约中的交易
//     2. 负责生成/验证合约中自动生成的交易

import (
	"bytes"
	"fmt"
	"time"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
)

func (uv *UtxoVM) isSmartContract(desc []byte) (*contract.TxDesc, bool) {
	if bytes.HasPrefix(desc, []byte("{")) {
		descObj, err := contract.Parse(string(desc))
		if err != nil {
			uv.xlog.Warn("parse contract failed", "desc", fmt.Sprintf("%s", desc))
			return nil, false
		}
		if descObj.Module == "" || descObj.Method == "" {
			return nil, false
		}
		// 判断合约是不是被注册
		allowedModules := uv.smartContract.GetAll()
		if _, ok := allowedModules[descObj.Module]; !ok {
			return nil, false
		}
		return descObj, err == nil
	}
	return nil, false
}

// TxOfRunningContractGenerate 预执行当前的交易里面的合约
func (uv *UtxoVM) TxOfRunningContractGenerate(txlist []*pb.Transaction, pendingBlock *pb.InternalBlock, outerBatch kvdb.Batch, ctxInit bool) ([]*pb.Transaction, kvdb.Batch, error) {
	var (
		newtxs []*pb.Transaction
	)
	var batch kvdb.Batch
	if outerBatch != nil {
		batch = outerBatch
	} else {
		batch = uv.ldb.NewBatch()
	}
	allowedModules := uv.smartContract.GetAll()
	txCtx := &contract.TxContext{
		Block:     pendingBlock,
		UtxoBatch: batch,
		LedgerObj: uv.ledger,
	}
	if ctxInit {
		for am := range allowedModules {
			if instance, ok := uv.smartContract.Get(am); ok {
				instance.SetContext(txCtx)
			}
		}
	}
	//统一单位是ns
	period := int64(uv.contractExectionTime * 1000 * 1000) // 0.5s
	addFailedNoRollback := func(tx *pb.Transaction, b *pb.InternalBlock, txErr error) {
		newtxs = append(newtxs, tx)
		b.FailedTxs[global.F(tx.Txid)] = txErr.Error()
	}
	addFailed := func(tx *pb.Transaction, b *pb.InternalBlock, txErr error) {
		//将交易的执行结果设置为contract execution fails
		//执行rollback
		err := uv.RollbackContract(b.Blockid, tx)
		if err != nil {
			uv.xlog.Error("rollback when addFailed", "error", err)
		}
		newtxs = append(newtxs, tx)
		b.FailedTxs[global.F(tx.Txid)] = txErr.Error()
	}
	addSucc := func(tx *pb.Transaction) {
		newtxs = append(newtxs, tx)
	}

	//如果第一个带合约的交易消耗了所有的合约执行时间，那么这个合约永远都执行不完，因此直接标记为失败。
	contractNo := 0
	//计算分配给合约的执行时间
	deadline := time.Now().UnixNano() + period
	for _, tx := range txlist {
		if txDesc, ok := uv.isSmartContract(tx.Desc); ok { // 交易需要执行智能合约
			txDesc.Tx = tx

			// 判断合约是否有效
			if _, ok := allowedModules[txDesc.Module]; !ok {
				//如果是没有注册的合约，直接当做一般的交易
				uv.xlog.Warn("module is not registered", "module", txDesc.Module)
				addSucc(tx)
				continue
			}
			contractNo++

			// 执行合约
			err := uv.runContract(pendingBlock.Blockid, tx, nil, deadline)
			if err != nil {
				if txDesc.Module == contract.KernelModuleName || txDesc.Module == contract.ConsensusModueName {
					addFailedNoRollback(tx, pendingBlock, err)
					continue
				}
				// 如果是超时,直接返回
				if err.Error() == common.ErrContractExecutionTimeout.Error() ||
					err.Error() == common.ErrContractConnectionError.Error() {
					//第一个就超时了, 打包交易，并且标记结果为失败
					if contractNo <= 1 && err.Error() == common.ErrContractExecutionTimeout.Error() {
						addFailed(tx, pendingBlock, err)
					}
					uv.xlog.Error("runContractWithTimeout", "error", err, "txid", fmt.Sprintf("%x", tx.Txid))
					//返回超时，让上层捕获处理
					return newtxs, batch, common.ErrContractExecutionTimeout
				}
				//其他错误,合约执行结果忽略，交易上链
				addFailed(tx, pendingBlock, err)
				continue
			}
			addSucc(tx)
		} else {
			addSucc(tx)
		}
	}
	return newtxs, batch, nil
}
