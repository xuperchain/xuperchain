package pb

import "math/big"
import (
	"bytes"
	"fmt"
)

// FeePlaceholder fee identifier to miner
const FeePlaceholder = "$"

// GetFrozenAmount 获得交易output中超过某height才能解冻的金额
func (tx *Transaction) GetFrozenAmount(height int64) *big.Int {
	sum := big.NewInt(0)
	for _, txOutput := range tx.TxOutputs {
		if txOutput.FrozenHeight > height {
			amount := big.NewInt(0)
			amount.SetBytes(txOutput.Amount)
			sum = sum.Add(sum, amount)
		}
	}
	return sum
}

// GetAmountByAddress 获得交易的Output中某个address的收益
func (tx *Transaction) GetAmountByAddress(address string) *big.Int {
	sum := big.NewInt(0)
	for _, txOutput := range tx.TxOutputs {
		if string(txOutput.ToAddr) == address {
			amount := big.NewInt(0)
			amount.SetBytes(txOutput.Amount)
			sum = sum.Add(sum, amount)
		}
	}
	return sum
}

// FromAddrInList 判断交易的发起人是否在白名单
func (tx *Transaction) FromAddrInList(whiteList map[string]bool) bool {
	if whiteList[tx.Initiator] {
		return true
	}
	return false
}

// HexTxid get txid in hex string
func (tx *Transaction) HexTxid() string {
	return fmt.Sprintf("%x", tx.Txid)
}

// GetFee get fee in tx output
func (tx *Transaction) GetFee() *big.Int {
	fee := big.NewInt(0)
	for _, txOutput := range tx.TxOutputs {
		addr := txOutput.ToAddr
		if !bytes.Equal(addr, []byte(FeePlaceholder)) {
			continue
		}
		fee.SetBytes(txOutput.Amount)
	}
	return fee
}

// GetCoinbaseTotal get total coinbase amount
func (ib *InternalBlock) GetCoinbaseTotal() *big.Int {
	total := big.NewInt(0)
	for _, tx := range ib.Transactions {
		if tx.Coinbase {
			total = total.Add(total, tx.GetFrozenAmount(-1))
		}
	}
	return total
}

// GetVersion get refid and offset as version string
func (txIn *TxInputExt) GetVersion() string {
	if txIn.RefTxid == nil {
		return ""
	}
	return fmt.Sprintf("%x_%d", txIn.RefTxid, txIn.RefOffset)
}

// ContainsTx returns whether a tx is included in a block
func (ib *InternalBlock) ContainsTx(txid []byte) bool {
	for _, tx := range ib.Transactions {
		if bytes.Equal(txid, tx.Txid) {
			return true
		}
	}
	return false
}

// GetTx returns a tx included in a block
func (ib *InternalBlock) GetTx(txid []byte) *Transaction {
	for _, tx := range ib.Transactions {
		if bytes.Equal(txid, tx.Txid) {
			return tx
		}
	}
	return nil
}
