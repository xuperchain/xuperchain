package xmodel

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/xuperchain/xuperchain/core/pb"
)

// UtxoCache makes utxo rwset
type UtxoCache struct {
	utxovm      UtxoVM
	inputCache  []*pb.TxInput
	outputCache []*pb.TxOutput
	intputIdx   int
	isPenetrate bool
}

func NewUtxoCache(utxovm UtxoVM) *UtxoCache {
	return &UtxoCache{
		utxovm:      utxovm,
		isPenetrate: true,
	}
}

func NewUtxoCacheWithInputs(inputs []*pb.TxInput) *UtxoCache {
	return &UtxoCache{
		inputCache:  inputs,
		isPenetrate: false,
	}
}

func (u *UtxoCache) selectUtxos(from string, amount *big.Int) (*big.Int, error) {
	if u.isPenetrate {
		inputs, _, total, err := u.utxovm.SelectUtxos(from, "", amount, false, false)
		if err != nil {
			return nil, err
		}
		u.inputCache = append(u.inputCache, inputs...)
		return total, nil
	}

	fromBytes := []byte(from)
	inputCache := u.inputCache[u.intputIdx:]
	sum := new(big.Int)
	n := 0
	for _, input := range inputCache {
		n++
		// Since contract calls bridge serially, a mismatched from address is an error
		if !bytes.Equal(input.GetFromAddr(), fromBytes) {
			return nil, errors.New("from address mismatch in utxo cache")
		}
		sum.Add(sum, new(big.Int).SetBytes(input.GetAmount()))
		if sum.Cmp(amount) >= 0 {
			break
		}
	}
	if sum.Cmp(amount) < 0 {
		return nil, errors.New("utxo not enough in utxo cache")
	}
	u.intputIdx += n
	return sum, nil
}

func (u *UtxoCache) Transfer(from, to string, amount *big.Int) error {
	if amount.Cmp(new(big.Int)) == 0 {
		return nil
	}
	total, err := u.selectUtxos(from, amount)
	if err != nil {
		return err
	}
	u.outputCache = append(u.outputCache, &pb.TxOutput{
		Amount: amount.Bytes(),
		ToAddr: []byte(to),
	})
	// make change
	if total.Cmp(amount) > 0 {
		u.outputCache = append(u.outputCache, &pb.TxOutput{
			Amount: new(big.Int).Sub(total, amount).Bytes(),
			ToAddr: []byte(from),
		})
	}
	return nil
}

func (u *UtxoCache) GetRWSets() ([]*pb.TxInput, []*pb.TxOutput) {
	if u.isPenetrate {
		return u.inputCache, u.outputCache
	}
	return u.inputCache[:u.intputIdx], u.outputCache
}
