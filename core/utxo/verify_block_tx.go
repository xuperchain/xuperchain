package utxo

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xuperchain/xuperchain/core/pb"
)

func (uv *UtxoVM) verifyBlockTxs(block *pb.InternalBlock, isRootTx bool, unconfirmToConfirm map[string]bool) error {
	var err error
	var once sync.Once
	wg := sync.WaitGroup{}
	dags := splitToDags(block)
	for _, txs := range dags {
		wg.Add(1)
		go func(txs []*pb.Transaction) {
			defer wg.Done()
			verifyErr := uv.verifyDAGTxs(txs, isRootTx, unconfirmToConfirm)
			onceBody := func() {
				err = verifyErr
			}
			// err 只被赋值一次
			if verifyErr != nil {
				once.Do(onceBody)
			}
		}(txs)
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
					ok, isRelyOnMarkedTx, err := uv.verifyMarked(tx)
					if isRelyOnMarkedTx {
						if !ok || err != nil {
							uv.xlog.Warn("tx verification failed because it is blocked tx", "err", err)
						} else {
							uv.xlog.Trace("blocked tx verification succeed")
						}
						return err
					}
					return errors.New("dotx failed to ImmediateVerifyTx error")
				}
			}
		}
	}

	return nil
}
