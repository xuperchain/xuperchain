package payload

import (
	"fmt"
)

func (input *TxInput) String() string {
	return fmt.Sprintf("TxInput{%s, Amount: %v, Sequence:%v}", input.Address, input.Amount, input.Sequence)
}
