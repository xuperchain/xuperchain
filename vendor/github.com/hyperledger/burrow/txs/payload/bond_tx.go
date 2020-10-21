package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
)

func NewBondTx(address crypto.Address, amount uint64) *BondTx {
	return &BondTx{
		Input: &TxInput{
			Address: address,
			Amount:  amount,
		},
	}
}

func (tx *BondTx) Type() Type {
	return TypeBond
}

func (tx *BondTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *BondTx) String() string {
	return fmt.Sprintf("BondTx{%v}", tx.Input)
}

func (tx *BondTx) AddInput(st acmstate.AccountGetter, pubkey crypto.PublicKey, amt uint64) error {
	addr := pubkey.GetAddress()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("invalid address %s from pubkey %s", addr, pubkey)
	}
	return tx.AddInputWithSequence(pubkey, amt, acc.Sequence+uint64(1))
}

func (tx *BondTx) AddInputWithSequence(pubkey crypto.PublicKey, amt uint64, sequence uint64) error {
	tx.Input = &TxInput{
		Address:  pubkey.GetAddress(),
		Amount:   amt,
		Sequence: sequence,
	}
	return nil
}

func (tx *BondTx) Any() *Any {
	return &Any{
		BondTx: tx,
	}
}
