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

// MergerUTXO merge all the UTXOs of an account, resulting into one or more transactions
func (uv *UtxoVM) MergeUtxos(fromAddr string, fromPubKey string, needLock, excludeUnconfirmed bool) ([]*pb.TxInput, [][]byte, *big.Int, error) {
	uv.xlog.Trace("start to merge utxos", "address", fromAddr)

	// Total amount selected
	amount := big.NewInt(0)
	maxTxSize := big.NewInt(uv.GetMaxBlockSize() / 100)
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
		// If the utxo has been frozen
		// case1: utxo's frozenHeight is greater than current ledger height
		// case2: utxo's frozenHeiht equals negative one
		if utxoItem.FrozenHeight > uv.ledger.GetMeta().GetTrunkHeight() || utxoItem.FrozenHeight == -1 {
			uv.xlog.Debug("utxo still frozen, skipped", "key", key)
			continue
		}
		// ignore UTXOTablePrefix and split the key with _
		realKey := bytes.Split(key[len(pb.UTXOTablePrefix):], []byte("_"))

		// build a tx input
		txInput := &pb.TxInput{}
		txInput.RefTxid, _ = hex.DecodeString(string(realKey[1]))
		offset, _ := strconv.Atoi(string(realKey[2]))
		txInput.RefOffset = int32(offset)
		txInput.FromAddr = []byte(fromAddr)
		txInput.Amount = utxoItem.Amount.Bytes()
		txInput.FrozenHeight = utxoItem.FrozenHeight

		txInputs = append(txInputs, txInput)
		amount.Add(amount, utxoItem.Amount)
		txInputSize += int64(proto.Size(txInput))

		// check size
		bs := big.NewInt(txInputSize)
		// If maxTxSize is bigger than maxTxSize, remove the biggest utxoItem
		if 1 == bs.Cmp(maxTxSize) {
			// remove the last one
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

func (uv *UtxoVM) QueryUtxoRecord(accountName string) (string, error) {
	defaultUtxoRecord := strconv.FormatInt(int64(0), 10)
	utxoRecord := int64(0)

	addrPrefix := fmt.Sprintf("%s%s_", pb.UTXOTablePrefix, accountName)
	it := uv.ldb.NewIteratorWithPrefix([]byte(addrPrefix))
	defer it.Release()

	for it.Next() {
		utxoRecord++
	}
	if it.Error() != nil {
		return defaultUtxoRecord, it.Error()
	}

	return strconv.FormatInt(utxoRecord, 10), nil
}
