package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

func NewUnbondTx(address crypto.Address, amount uint64) *UnbondTx {
	return &UnbondTx{
		Input: &TxInput{
			Address: address,
		},
		Output: &TxOutput{
			Address: address,
			Amount:  amount,
		},
	}
}

func (tx *UnbondTx) Type() Type {
	return TypeUnbond
}

func (tx *UnbondTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *UnbondTx) String() string {
	return fmt.Sprintf("UnbondTx{%v}", tx.Input.Address)
}

func (tx *UnbondTx) Any() *Any {
	return &Any{
		UnbondTx: tx,
	}
}
