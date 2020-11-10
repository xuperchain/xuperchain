package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
)

func NewSendTx() *SendTx {
	return &SendTx{
		Inputs:  []*TxInput{},
		Outputs: []*TxOutput{},
	}
}

func (tx *SendTx) GetInputs() []*TxInput {
	return tx.Inputs
}

func (tx *SendTx) Type() Type {
	return TypeSend
}

func (tx *SendTx) String() string {
	return fmt.Sprintf("SendTx{%v -> %v}", tx.Inputs, tx.Outputs)
}

func (tx *SendTx) AddInput(st acmstate.AccountGetter, pubkey crypto.PublicKey, amt uint64) error {
	addr := pubkey.GetAddress()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("AddInput: could not find account with address %v", addr)
	}
	return tx.AddInputWithSequence(pubkey, amt, acc.Sequence+1)
}

func (tx *SendTx) AddInputWithSequence(pubkey crypto.PublicKey, amt uint64, sequence uint64) error {
	addr := pubkey.GetAddress()
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:  addr,
		Amount:   amt,
		Sequence: sequence,
	})
	return nil
}

func (tx *SendTx) AddOutput(addr crypto.Address, amt uint64) error {
	tx.Outputs = append(tx.Outputs, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

func (tx *SendTx) Any() *Any {
	return &Any{
		SendTx: tx,
	}
}
