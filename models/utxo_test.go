package models

import (
	"encoding/hex"
	"math/big"
	"testing"
)

func TestLockedUtxo_Hash(t *testing.T) {

	type fields struct {
		bcName  string
		address string
		amount  *big.Int
	}
	tests := []struct {
		name   string
		fields fields
		want   string // hex string format of hashed value
	}{
		{
			name: "compatibility",
			fields: fields{
				bcName:  "xuper",
				address: "TeyyPLpp9L7QAcxHangtcHTu7HUZ6iydY",
				amount:  big.NewInt(0),
			},
			want: "21eb1c4c51abeb7c19cfffdb21b1cf685deee636d0aae77a6b73fa1ff2c4855c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &LockedUtxo{
				bcName:  tt.fields.bcName,
				address: tt.fields.address,
				amount:  tt.fields.amount,
			}
			if got := o.Hash(); hex.EncodeToString(got) != tt.want {
				t.Errorf("LockedUtxo.Hash() = %v,\n"+
					"want %v",
					hex.EncodeToString(got), tt.want)
			}
		})
	}
}
