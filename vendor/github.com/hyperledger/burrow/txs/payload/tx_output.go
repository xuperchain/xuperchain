package payload

import (
	"fmt"
)

func (txOut *TxOutput) String() string {
	return fmt.Sprintf("TxOutput{%s, Amount: %v}", txOut.Address, txOut.Amount)
}
