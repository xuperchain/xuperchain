package utxo

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xuperchain/xuperunion/pb"
)

func (uv *UtxoVM) verifyBlockTxs(block *pb.InternalBlock, isRootTx bool, unconfirmToConfirm map[string]bool) error {
	var err error
	txs := block.Transactions
	wg := sync.WaitGroup{}
	for _, tx := range txs {
		if err != nil {
			break
		}
		wg.Add(1)
		go func(tx *pb.Transaction, isRootTx bool, unconfirmToConfirm map[string]bool) {
			defer wg.Done()
			verifyErr := uv.verifyTx(tx, isRootTx, unconfirmToConfirm)
			if verifyErr != nil {
				err = verifyErr
			}
		}(tx, isRootTx, unconfirmToConfirm)
	}
	wg.Wait()
	return err
}

func (uv *UtxoVM) verifyTx(tx *pb.Transaction, isRootTx bool, unconfirmToConfirm map[string]bool) error {
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
	return nil
}
