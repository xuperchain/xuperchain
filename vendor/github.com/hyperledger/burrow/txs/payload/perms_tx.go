package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
)

func NewPermsTx(st acmstate.AccountGetter, from crypto.PublicKey, args permission.PermArgs) (*PermsTx, error) {
	addr := from.GetAddress()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("NewPermsTx: could not find account with address %v", addr)
	}

	sequence := acc.Sequence + 1
	return NewPermsTxWithSequence(from, args, sequence), nil
}

func NewPermsTxWithSequence(from crypto.PublicKey, args permission.PermArgs, sequence uint64) *PermsTx {
	input := &TxInput{
		Address:  from.GetAddress(),
		Amount:   1, // NOTE: amounts can't be 0 ...
		Sequence: sequence,
	}

	return &PermsTx{
		Input:    input,
		PermArgs: args,
	}
}

func (tx *PermsTx) Type() Type {
	return TypePermissions
}

func (tx *PermsTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *PermsTx) String() string {
	return fmt.Sprintf("PermsTx{%v -> %v}", tx.Input, tx.PermArgs)
}

func (tx *PermsTx) Any() *Any {
	return &Any{
		PermsTx: tx,
	}
}
