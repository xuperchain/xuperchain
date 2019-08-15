package utxo

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xuperchain/xuperunion/pb"
)

func (uv *UtxoVM) verifyBlockTxs(block *pb.InternalBlock, isRootTx bool, unconfirmToConfirm map[string]bool) error {
	var err error
	var once sync.Once
	txs := block.Transactions

	_, txGraph, buildTxDepsErr := uv.buildTxDeps(txs)
	if buildTxDepsErr != nil {
		return buildTxDepsErr
	}
	//order, cyclic, childDAGSizeArr := TopSortDFS(txGraph)
	_, cyclic, childDAGSizeArr := TopSortDFS(txGraph)
	//if cyclic != nil {
	if cyclic {
		return errors.New("there is a cycle in block's transactions")
	}
	wg := sync.WaitGroup{}
	count, length := 0, len(childDAGSizeArr)
	start, end := 0, 0
	for count < length {
		if err != nil {
			break
		}
		end += childDAGSizeArr[count]
		wg.Add(1)
		go func(start int, end int, txs []*pb.Transaction) {
			defer wg.Done()
			verifyErr := uv.verifyDAGTxs(txs[start:end], isRootTx, unconfirmToConfirm)
			onceBody := func() {
				err = verifyErr
			}
			// err 只被赋值一次
			if verifyErr != nil {
				once.Do(onceBody)
			}
		}(start, end, txs)
		start = end
		count++
	}
	wg.Wait()
	return err
}

func (uv *UtxoVM) verifyDAGTxs(txs []*pb.Transaction, isRootTx bool, unconfirmToConfirm map[string]bool) error {
	for _, tx := range txs {
		if tx == nil {
			return errors.New("verifyTx error, tx is nil")
		}
		txid := string(tx.GetTxid())
		if unconfirmToConfirm[txid] == false {
			if !uv.verifyAutogenTx(tx) {
				return ErrInvalidAutogenTx
			}
			if !tx.Autogen && !tx.Coinbase {
				if ok, err := uv.ImmediateVerifyTx(tx, isRootTx); !ok {
					uv.xlog.Warn("dotx failed to ImmediateVerifyTx", "txid", fmt.Sprintf("%x", tx.Txid), "err", err)
					return errors.New("dotx failed to ImmediateVerifyTx error")
				}
			}
		}
	}

	return nil
}
