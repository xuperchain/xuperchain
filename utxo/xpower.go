package utxo

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
)

// CalcXPowerWithSelf ...
func (uv *UtxoVM) CalcXPower(address string, currHeight int64) (*big.Float, error) {
	totalNeedSelf, selfErr := uv.calcXPower(address, currHeight)
	if selfErr != nil {
		return big.NewFloat(0), selfErr
	}
	totalNeedLease, leaseErr := uv.calcXPower("lease#"+address, currHeight)
	if leaseErr != nil {
		return big.NewFloat(0), leaseErr
	}
	return totalNeedSelf.Add(totalNeedSelf, totalNeedLease), nil
}

// CalcXPower calc the xpower value
func (uv *UtxoVM) calcXPower(address string, currHeight int64) (*big.Float, error) {
	prefixKey := pb.AddressUTXOTablePrefix + address + "_"

	it := uv.ldb.NewIteratorWithPrefix([]byte(prefixKey))
	defer it.Release()

	totalNeed := big.NewFloat(0)
	tmp := big.NewInt(0)

	// TODO, 目前直接算余额
	for it.Next() {
		keyStr := string(it.Key())
		value := it.Value()
		// the formula of calculating the xpower
		// key: commonKey + address + "_" + startHeight
		startHeight := int64(0)
		splitStr := strings.Split(keyStr, "_")
		if len(splitStr) <= 0 {
			heightStr := splitStr[len(splitStr)-1]
			value, err := strconv.ParseInt(heightStr, 10, 64)
			if err != nil {
				return big.NewFloat(0), err
			}
			startHeight = value
		}
		// TODO, @ToWorld N as 10
		//beta := float64((currHeight - (startHeight + 10)) / (10 + 0.0))
		threshold := currHeight/100 - startHeight
		beta := float64(currHeight-startHeight*100) / float64(100)
		// 说明utxo已经恢复到实际价值
		if threshold > 0 {
			beta = 1.0
		}

		//tmpFloat := big.NewFloat(float64(tmp.SetBytes(value).Int64()) * beta)
		tmpFloat := big.NewFloat(0)
		tmpFloat.SetInt(tmp.SetBytes(value))
		tmpFloat.Mul(tmpFloat, big.NewFloat(beta))
		totalNeed.Add(totalNeed, tmpFloat)
	}
	if it.Error() != nil {
		return big.NewFloat(0), it.Error()
	}

	return totalNeed, nil
}

// SaveUTXOByHeightInterval save utxo
func (uv *UtxoVM) SaveUTXOByHeightInterval(addrToUtxo map[string]*big.Int, height int64, batchWrite kvdb.Batch) error {
	for address, utxoValue := range addrToUtxo {
		// TODO, @ToWorld 给出具体的N值，这个N值也是共识的一部分
		startHeight := height / 100
		// TODO, @ToWorld support an API to generate a complete key
		completeKey := pb.AddressUTXOTablePrefix + address + "_" + strconv.FormatInt(startHeight, 10)
		oldValue := big.NewInt(0)
		value, findErr := uv.ldb.Get([]byte(completeKey))
		if findErr != nil && common.NormalizedKVError(findErr) != common.ErrKVNotFound {
			return findErr
		}
		oldValue.SetBytes(value)
		oldValue.Add(oldValue, utxoValue)
		err := batchWrite.Put([]byte(completeKey), oldValue.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (uv *UtxoVM) StatUTXOWithBlock(transactions []*pb.Transaction, isUndo bool) map[string]*big.Int {
	addrToUtxo := map[string]*big.Int{}

	// if transactions is nil, return empty map directly
	if transactions == nil {
		return addrToUtxo
	}
	for _, tx := range transactions {
		if tx == nil {
			continue
		}
		// update TxInputs
		// address -> utxos, Sub
		txInputs := tx.GetTxInputs()
		amount := big.NewInt(0)
		for _, txInput := range txInputs {
			amount.SetBytes(txInput.GetAmount())
			if isUndo {
				amount.Mul(amount, big.NewInt(-1))
			}
			from := string(txInput.GetFromAddr())
			if addrToUtxo[from] == nil {
				addrToUtxo[from] = big.NewInt(0)
			}
			addrToUtxo[from].Sub(addrToUtxo[from], amount)
		}

		// update TxOutputs
		// address -> utxos, Add
		// TODO, @ToWorld 针对$做特殊处理
		txOutputs := tx.GetTxOutputs()
		for _, txOutput := range txOutputs {
			amount.SetBytes(txOutput.GetAmount())
			if isUndo {
				amount.Mul(amount, big.NewInt(-1))
			}
			to := string(txOutput.GetToAddr())
			if addrToUtxo[to] == nil {
				addrToUtxo[to] = big.NewInt(0)
			}
			addrToUtxo[to].Add(addrToUtxo[to], amount)
		}
	}

	return addrToUtxo
}
