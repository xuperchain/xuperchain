package payload

import (
	"fmt"

	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/crypto"
)

func NewCallTx(st acmstate.AccountGetter, from crypto.PublicKey, to *crypto.Address, data []byte,
	amt, gasLimit, fee uint64) (*CallTx, error) {

	addr := from.GetAddress()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, fmt.Errorf("NewCallTx: could not find account with address %v", addr)
	}

	sequence := acc.Sequence + 1
	return NewCallTxWithSequence(from, to, data, amt, gasLimit, fee, sequence), nil
}

func NewCallTxWithSequence(from crypto.PublicKey, to *crypto.Address, data []byte,
	amt, gasLimit, fee, sequence uint64) *CallTx {
	input := &TxInput{
		Address:  from.GetAddress(),
		Amount:   amt,
		Sequence: sequence,
	}

	return &CallTx{
		Input:    input,
		Address:  to,
		GasLimit: gasLimit,
		Fee:      fee,
		Data:     data,
	}
}

func (tx *CallTx) Type() Type {
	return TypeCall
}
func (tx *CallTx) GetInputs() []*TxInput {
	return []*TxInput{tx.Input}
}

func (tx *CallTx) String() string {
	return fmt.Sprintf("CallTx{%v -> %s: %v}", tx.Input, tx.Address, tx.Data)
}

func (tx *CallTx) Any() *Any {
	return &Any{
		CallTx: tx,
	}
}
