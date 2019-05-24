package xchaincore

import (
	"bytes"
	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
	"time"
)

const (
	// PrePlayInterval prepaly interval
	PrePlayInterval = 300 //ms
	// PrePlayMaxNum preplay max num for each block
	PrePlayMaxNum = 30000
)

// PipelineMiner ...
type PipelineMiner struct {
	batch   kvdb.Batch
	queue   []*pb.Transaction
	inflag  map[string]bool
	xc      *XChainCore
	pause   bool
	txsSize int64
	//合约执行超时之后，不能继续小范围的预执行，必须等待下一个块开始
	contractExecutionTimeoutWait bool
	failedTxs                    map[string]string //失败的合约 txid -> err
}

// NewPipelineMiner ...
func NewPipelineMiner(xc *XChainCore) *PipelineMiner {
	pm := &PipelineMiner{
		batch:     xc.Utxovm.NewBatch(),
		queue:     []*pb.Transaction{},
		inflag:    map[string]bool{},
		xc:        xc,
		txsSize:   0,
		failedTxs: map[string]string{},
	}
	return pm
}

func (pm *PipelineMiner) doPrePlay() (needWait bool) {
	needWait = true
	maxSizeB := pm.xc.Ledger.GetMaxBlockSize()
	if pm.txsSize >= maxSizeB || pm.contractExecutionTimeoutWait {
		return
	}
	pm.xc.mutex.Lock()
	defer pm.xc.mutex.Unlock()
	txsUnconf, err := pm.xc.Utxovm.GetUnconfirmedTx(true)
	if err != nil {
		pm.xc.log.Warn("[Minning] fail to get unconfirmedtx")
		return
	}
	ledgerLastID := pm.xc.Ledger.GetMeta().TipBlockid
	utxovmLastID := pm.xc.Utxovm.GetLatestBlockid()
	if !bytes.Equal(ledgerLastID, utxovmLastID) {
		pm.xc.log.Warn("can not preplay, because ledger last blockid is not equal utxovm last id")
		return
	}
	t := time.Now()
	txs := []*pb.Transaction{}
	for _, tx := range txsUnconf {
		if !pm.inflag[string(tx.Txid)] {
			txs = append(txs, tx)
		}
		// 按照指标要求, 这里限制了一下每个小区间内的最大tx个数, 避免交易过于繁忙时导致不能及时出块
		if len(txs) >= PrePlayMaxNum {
			needWait = false
			break
		}
	}
	if len(txs) == 0 {
		return
	}
	meta := pm.xc.Ledger.GetMeta()
	// todo 下面两步操作时间之和要小于出块周期
	// make fake block
	fakeBlock, err := pm.xc.Ledger.FormatFakeBlock(txs, pm.xc.address, pm.xc.privateKey,
		t.UnixNano(), 0, 0, meta.TipBlockid, pm.xc.Utxovm.GetTotal())
	if err != nil {
		pm.xc.log.Warn("[Minning] format block error", "logid")
		return
	}
	needInitCtx := pm.NeedInitCtx() //是否需要初始化合约机context
	// pre-execute the contract
	if txs, _, err = pm.xc.Utxovm.TxOfRunningContractGenerate(txs, fakeBlock, pm.batch, needInitCtx); err != nil {
		if err.Error() == common.ErrContractExecutionTimeout.Error() {
			pm.contractExecutionTimeoutWait = true
		} else {
			pm.xc.log.Warn("PrePlay failed", "error", err)
			return
		}
	}
	for _, tx := range txs {
		pm.queue = append(pm.queue, tx)
		pm.inflag[string(tx.Txid)] = true
		if failedErr, failed := fakeBlock.FailedTxs[global.F(tx.Txid)]; failed {
			pm.failedTxs[global.F(tx.Txid)] = failedErr
		}
		txSize, err := common.GetTxSerializedSize(tx)
		if err != nil {
			pm.xc.log.Warn("[Minning] fail to get tx size")
			return
		}
		pm.txsSize += txSize
		if pm.txsSize >= maxSizeB {
			return
		}
	}
	return
}

// NeedInitCtx return true if queue is empty
func (pm *PipelineMiner) NeedInitCtx() bool {
	return len(pm.queue) == 0 //队列空，说明拿去打包了，因此可以重新初始化合约机们的SetContext
}

// Start start preplay service
func (pm *PipelineMiner) Start() error {
	for {
		// 当交易非常繁忙时不需要等待300ms, 可以直接进行下一步打包
		var needWait = true
		if !pm.pause {
			needWait = pm.doPrePlay()
		}
		if needWait {
			time.Sleep(PrePlayInterval * time.Millisecond)
		}
	}
}

// Resume resume preplay service
func (pm *PipelineMiner) Resume() {
	pm.pause = false
}

// Pause stop preplay service
func (pm *PipelineMiner) Pause() {
	pm.pause = true
}

// FetchTxs fetch txs from unconfirmed table
func (pm *PipelineMiner) FetchTxs() (kvdb.Batch, []*pb.Transaction, map[string]string) {
	batch := pm.batch
	txs := []*pb.Transaction{}
	failedTxs := pm.failedTxs
	for _, tx := range pm.queue {
		if ok, _ := pm.xc.Utxovm.HasTx(tx.Txid); ok { //确认一下没被回滚
			txs = append(txs, tx)
		} else {
			pm.xc.Utxovm.RollbackContract([]byte(""), tx) //回滚合约预执行的影响
		}
	}
	pm.batch = pm.xc.Utxovm.NewBatch()
	pm.queue = []*pb.Transaction{}
	pm.inflag = map[string]bool{}
	pm.txsSize = 0
	pm.contractExecutionTimeoutWait = false
	pm.failedTxs = map[string]string{}
	return batch, txs, failedTxs
}

// RollbackPrePlay rollback pre-execution contract
func (pm *PipelineMiner) RollbackPrePlay() error {
	_, txs, _ := pm.FetchTxs()
	for i := len(txs) - 1; i >= 0; i-- {
		err := pm.xc.Utxovm.RollbackContract([]byte(""), txs[i])
		if err != nil {
			return err
		}
	}
	return nil
}
