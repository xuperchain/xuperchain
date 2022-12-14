package models

import (
	"math/big"
	"strconv"

	"github.com/xuperchain/xupercore/lib/crypto/hash"
)

// LockedUtxo UTXO to be locked for operation
type LockedUtxo struct {
	bcName  string // blockchain name
	address string // address for UTXO belongs to
	amount  *big.Int
}

// NewLockedUtxo creates a given amount UTXO of the address in the blockchain
func NewLockedUtxo(bcName, address string, amount *big.Int) *LockedUtxo {
	return &LockedUtxo{
		bcName:  bcName,
		address: address,
		amount:  amount,
	}
}

// NewLockedUtxoAll creates a locked UTXO denotes all UTXO of the address in the blockchain
func NewLockedUtxoAll(bcName, address string) *LockedUtxo {
	return NewLockedUtxo(bcName, address, big.NewInt(0))
}

// Hash gets hash value of locked UTXO for signature
func (o *LockedUtxo) Hash() []byte {
	hashKey := o.bcName + o.address + o.amount.String() + strconv.FormatBool(true)
	hashValue := hash.DoubleSha256([]byte(hashKey))
	return hashValue
}
