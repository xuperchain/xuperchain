package utxo

import (
	"context"
	"time"

	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

// Async settings
const (
	AsyncMaxWaitMS   = 100
	AsyncMaxWaitSize = 7000
	AsyncQueueBuffer = 500000
)

// StartAsyncWriter start the Asynchronize writer
func (uv *UtxoVM) StartAsyncWriter() {
	uv.asyncMode = true
	ctx, cancel := context.WithCancel(context.Background())
	uv.asyncCancel = cancel
	ledger.DisableTxDedup = true
	go uv.asyncWriter(ctx)
	go uv.asyncVerifiy(ctx)
}

func (uv *UtxoVM) verifyTxWorker(itxlist []*InboundTx) error {
	if len(itxlist) == 0 {
		return nil
	}
	uv.xlog.Debug("async tx list size", "size", len(itxlist))
	//校验tx合法性
	for _, itx := range itxlist {
		ok, xerr := uv.ImmediateVerifyTx(itx.tx, false)
		if !ok {
			uv.xlog.Warn("invalid transaction found", "txid", global.F(itx.tx.Txid), "err", xerr)
		} else {
			uv.verifiedTxChan <- itx.tx
		}
	}
	return nil
}

// checkConflictTxs 检测一个batch内部的utxo引用冲突的txList
func (uv *UtxoVM) checkConflictTxs(txList []*pb.Transaction) map[string]bool {
	conflictUtxos := map[string]bool{}
	conflictTxs := map[string]bool{}
	for _, tx := range txList {
		innerConflict := false
		for _, txInput := range tx.TxInputs {
			utxoKey := GenUtxoKeyWithPrefix(txInput.FromAddr, txInput.RefTxid, txInput.RefOffset)
			if !conflictUtxos[utxoKey] {
				conflictUtxos[utxoKey] = true
			} else {
				innerConflict = true
				uv.xlog.Warn("utxo has been used by previous tx in the batch", "utxo", utxoKey)
				break
			}
		}
		if innerConflict {
			conflictTxs[string(tx.Txid)] = true
		}
	}
	return conflictTxs
}

// 一次刷一批tx到存储
func (uv *UtxoVM) flushTxList(txList []*pb.Transaction) error {
	if len(txList) == 0 {
		return nil
	}
	uv.xlog.Warn("async tx list size", "size", len(txList))
	batch := uv.ldb.NewBatch()
	conflictedTxs := uv.checkConflictTxs(txList)
	uv.mutex.Lock()
	defer uv.mutex.Unlock()
	for uv.asyncTryBlockGen { //避让出块的线程
		uv.asyncCond.Wait() //会临时让出锁
	}
	for _, tx := range txList {
		if conflictedTxs[string(tx.Txid)] {
			continue
		}
		doErr := uv.doTxInternal(tx, batch)
		if doErr != nil {
			uv.xlog.Warn("doTxInternal failed, when DoTx", "doErr", doErr)
			continue
		}
		uv.unconfirmTxInMem.Store(string(tx.Txid), tx)
		// uv.xlog.Debug("print tx size when DoTx", "tx_size", batch.ValueSize(), "txid", fmt.Sprintf("%x", tx.Txid))
	}
	writeErr := batch.Write()
	if writeErr != nil {
		uv.ClearCache()
		uv.xlog.Warn("fail to save to ldb", "writeErr", writeErr)
	}
	return writeErr
}

// asyncWriter 异步批量执行tx, 在AsyncMode=true时开启
func (uv *UtxoVM) asyncWriter(ctx context.Context) {
	tick := time.Tick(time.Millisecond * AsyncMaxWaitMS)
	txList := []*pb.Transaction{}
	uv.asyncWriterWG.Add(1)
	for {
		select {
		case tx := <-uv.verifiedTxChan:
			txList = append(txList, tx)
			if len(txList) > AsyncMaxWaitSize {
				go uv.flushTxList(txList)
				txList = []*pb.Transaction{}
			}
		case <-tick:
			go uv.flushTxList(txList)
			txList = []*pb.Transaction{}
		case <-ctx.Done():
			uv.asyncWriterWG.Done()
			return
		}
	}
}

// asyncVerifiy 异步并行校验tx，在AsyncMode=true时开启
func (uv *UtxoVM) asyncVerifiy(ctx context.Context) {
	tick := time.Tick(time.Millisecond * AsyncMaxWaitMS)
	itxlist := []*InboundTx{}
	uv.asyncWriterWG.Add(1)
	for {
		select {
		case itx := <-uv.inboundTxChan:
			itxlist = append(itxlist, itx)
			if len(itxlist) > AsyncMaxWaitSize {
				go uv.verifyTxWorker(itxlist)
				itxlist = []*InboundTx{}
			}
		case <-tick:
			go uv.verifyTxWorker(itxlist)
			itxlist = []*InboundTx{}
		case <-ctx.Done():
			uv.RollBackUnconfirmedTx()
			uv.asyncWriterWG.Done()
			return
		}
	}
}

// SetBlockGenEvent set if try to generate block in async mode
func (uv *UtxoVM) SetBlockGenEvent() {
	uv.asyncTryBlockGen = true
}

// NotifyFinishBlockGen notify to finish generating block
func (uv *UtxoVM) NotifyFinishBlockGen() {
	if !uv.asyncMode {
		return
	}
	uv.asyncTryBlockGen = false
	uv.asyncCond.Broadcast()
}

// IsAsync return current async state
func (uv *UtxoVM) IsAsync() bool {
	return uv.asyncMode
}
