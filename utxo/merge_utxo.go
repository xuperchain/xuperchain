package utxo

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperunion/pb"
)

func (uv *UtxoVM) MergeUtxos(fromAddr string, fromPubKey string, needLock, excludeUnconfirmed bool) ([]*pb.TxInput, [][]byte, *big.Int, error) {
	uv.xlog.Trace("start to merge utxos", "address", fromAddr)

	// Total amount selected
	amount := big.NewInt(0)
	maxTxSizePerBlock, _ := uv.MaxTxSizePerBlock()
	maxTxSize := big.NewInt(int64(maxTxSizePerBlock / 2))
	willLockKeys := make([][]byte, 0)
	txInputs := []*pb.TxInput{}
	txInputSize := int64(0)

	addrPrefix := fmt.Sprintf("%s%s_", pb.UTXOTablePrefix, fromAddr)
	it := uv.ldb.NewIteratorWithPrefix([]byte(addrPrefix))
	defer it.Release()

	for it.Next() {
		key := append([]byte{}, it.Key()...)
		utxoItem := new(UtxoItem)
		// 反序列化utxoItem
		uErr := utxoItem.Loads(it.Value())
		if uErr != nil {
			uv.xlog.Warn("load utxo failed, skipped", "key", key)
			continue
		}
		if needLock {
			if uv.tryLockKey(key) {
				willLockKeys = append(willLockKeys, key)
			} else {
				uv.xlog.Debug("can not lock the utxo key, conflict", "key", key)
				continue
			}
		}
		// If the utxo has been locked
		if uv.isLocked(key) {
			uv.xlog.Debug("utxo locked, skipped", "key", key)
			continue
		}
		if utxoItem.FrozenHeight > uv.ledger.GetMeta().GetTrunkHeight() || utxoItem.FrozenHeight == -1 {
			uv.xlog.Debug("utxo still frozen, skipped", "key", key)
			continue
		}

		realKey := bytes.Split(key[len(pb.UTXOTablePrefix):], []byte("_"))
		// build a tx input
		txInput := &pb.TxInput{
			FromAddr:     []byte(fromAddr),
			Amount:       utxoItem.Amount.Bytes(),
			FrozenHeight: utxoItem.FrozenHeight,
		}
		txInput.RefTxid, _ = hex.DecodeString(string(realKey[1]))
		offset, _ := strconv.Atoi(string(realKey[2]))
		txInput.RefOffset = int32(offset)

		txInputs = append(txInputs, txInput)
		amount.Add(amount, utxoItem.Amount)
		txInputSize += int64(proto.Size(txInput))

		// check size
		txInputSize := big.NewInt(txInputSize)
		if txInputSize.Cmp(maxTxSize) == 1 {
			txInputs = txInputs[:len(txInputs)-1]
			amount.Sub(amount, utxoItem.Amount)
			break
		} else {
			continue
		}
	}
	if it.Error() != nil {
		return nil, nil, nil, it.Error()
	}

	return txInputs, willLockKeys, amount, nil
}
