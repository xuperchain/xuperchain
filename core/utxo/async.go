package utxo

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	XModel "github.com/xuperchain/xuperchain/core/xmodel"
)

// Async settings
const (
	AsyncMaxWaitMS   = 10
	AsyncMaxWaitSize = 7000
	AsyncQueueBuffer = 500000
)

type AsyncResult struct {
	mailBox sync.Map
}

func (ar *AsyncResult) Open(txid []byte) error {
	_, loaded := ar.mailBox.LoadOrStore(string(txid), make(chan error, 1))
	if loaded {
		return ErrDuplicatedTx
	}
	return nil
}

func (ar *AsyncResult) Send(txid []byte, err error) {
	ch, _ := ar.mailBox.Load(string(txid))
	if ch != nil {
		ch.(chan error) <- err
	}
}

func (ar *AsyncResult) Wait(txid []byte) error {
	ch, _ := ar.mailBox.Load(string(txid))
	err := <-ch.(chan error)
	ar.mailBox.Delete(string(txid))
	return err
}

// StartAsyncWriter start the Asynchronize writer
func (uv *UtxoVM) StartAsyncWriter() {
	uv.asyncMode = true
	ctx, cancel := context.WithCancel(context.Background())
	uv.asyncCancel = cancel
	ledger.DisableTxDedup = true
	go uv.asyncWriter(ctx)
	go uv.asyncVerify(ctx)
}

func (uv *UtxoVM) StartAsyncBlockMode() {
	uv.asyncBlockMode = true
	ctx, cancel := context.WithCancel(context.Background())
	uv.asyncCancel = cancel
	ledger.DisableTxDedup = true
	go uv.asyncWriter(ctx)
	go uv.asyncVerify(ctx)
}

func (uv *UtxoVM) verifyTxWorker(itxlist []*InboundTx) error {
	if len(itxlist) == 0 {
		return nil
	}
	uv.xlog.Debug("async tx list size", "size", len(itxlist))
	//校验tx合法性
	for _, itx := range itxlist {
		// 去重判断
		tx := itx.tx
		ok, xerr := uv.ImmediateVerifyTx(tx, false)
		if !ok {
			uv.xlog.Warn("invalid transaction found", "txid", global.F(tx.Txid), "err", xerr)
			uv.asyncResult.Send(tx.Txid, xerr)
		} else {
			uv.verifiedTxChan <- tx
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
		for _, txOutputExt := range tx.TxOutputsExt {
			writeSetKey := XModel.GenWriteKeyWithPrefix(txOutputExt)
			if !conflictUtxos[writeSetKey] {
				conflictUtxos[writeSetKey] = true
			} else {
				innerConflict = true
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
	uv.xlog.Warn("async tx list size", "size", len(txList), "L1", len(uv.inboundTxChan), "L2", len(uv.verifiedTxChan))
	pbTxList := make([][]byte, len(txList))
	// 异步阻塞模式需要将交易数据持久化落盘
	if uv.asyncBlockMode {
		for i, tx := range txList {
			pbTxBuf, pbErr := proto.Marshal(tx)
			if pbErr != nil {
				uv.xlog.Warn("    fail to marshal tx", "pbErr", pbErr)
				pbTxList[i] = nil
				continue
			}
			pbTxList[i] = pbTxBuf
		}
	}
	conflictedTxs := uv.checkConflictTxs(txList)
	uv.mutex.Lock()
	defer uv.mutex.Unlock()
	for uv.asyncTryBlockGen { //避让出块的线程
		uv.asyncCond.Wait() //会临时让出锁
	}
	batch := uv.asyncBatch
	batch.Reset()
	for i, tx := range txList {
		if uv.asyncBlockMode && pbTxList[i] == nil {
			uv.asyncResult.Send(tx.Txid, errors.New("marshal failed"))
			continue
		}
		if conflictedTxs[string(tx.Txid)] {
			uv.asyncResult.Send(tx.Txid, errors.New("conflict tx"))
			pbTxList[i] = nil
			continue
		}
		doErr := uv.doTxInternal(tx, batch, nil)
		if doErr != nil {
			uv.xlog.Warn("doTxInternal failed, when DoTx", "doErr", doErr)
			uv.asyncResult.Send(tx.Txid, doErr)
			pbTxList[i] = nil
			continue
		}
		uv.unconfirmTxInMem.Store(string(tx.Txid), tx)
		if uv.asyncBlockMode {
			batch.Put(append([]byte(pb.UnconfirmedTablePrefix), tx.Txid...), pbTxList[i])
		}
		// uv.xlog.Debug("print tx size when DoTx", "tx_size", batch.ValueSize(), "txid", fmt.Sprintf("%x", tx.Txid))
	}
	writeErr := batch.Write()
	if writeErr != nil {
		uv.ClearCache()
		uv.xlog.Warn("fail to save to ldb", "writeErr", writeErr)
	}
	// 对于异步阻塞模式，有必要在执行完一个交易后同步PostTx，并将结果返回给客户端
	if uv.asyncBlockMode {
		go func() {
			for i, tx := range txList {
				if pbTxList[i] == nil {
					continue
				}
				uv.asyncResult.Send(tx.Txid, nil)
			}
		}()
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
func (uv *UtxoVM) asyncVerify(ctx context.Context) {
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
			//uv.RollBackUnconfirmedTx()
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
	if !uv.asyncMode && !uv.asyncBlockMode {
		return
	}
	uv.asyncTryBlockGen = false
	uv.asyncCond.Broadcast()
}

// IsAsync return current async state
func (uv *UtxoVM) IsAsync() bool {
	return uv.asyncMode
}

// IsAsyncBlock return current async state
func (uv *UtxoVM) IsAsyncBlock() bool {
	return uv.asyncBlockMode
}
