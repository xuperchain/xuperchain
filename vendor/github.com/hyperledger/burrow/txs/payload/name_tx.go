package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
)

func NewNameTx(st acmstate.AccountGetter, from crypto.PublicKey, name, data string, amt, fee uint64) (*NameTx, error) {
	addr := from.GetAddress()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("NewNameTx: could not find account with address %v", addr)
	}

	sequence := acc.Sequence + 1
	return NewNameTxWithSequence(from, name, data, amt, fee, sequence), nil
}

func NewNameTxWithSequence(from crypto.PublicKey, name, data string, amt, fee, sequence uint64) *NameTx {
	input := &TxInput{
		Address:  from.GetAddress(),
		Amount:   amt,
		Sequence: sequence,
	}

	return &NameTx{
		Input: input,
		Name:  name,
		Data:  data,
		Fee:   fee,
	}
}

func (tx *NameTx) Type() Type {
	return TypeName
}

func (tx *NameTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *NameTx) String() string {
	return fmt.Sprintf("NameTx{%v -> %s: %s}", tx.Input, tx.Name, tx.Data)
}

func (tx *NameTx) Any() *Any {
	return &Any{
		NameTx: tx,
	}
}
