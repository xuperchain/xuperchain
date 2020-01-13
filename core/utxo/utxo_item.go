package utxo

import (
	"bytes"
	"encoding/json"
	"math/big"
)

// UtxoItem the data structure of an UTXO item
type UtxoItem struct {
	Amount       *big.Int //utxo的面值
	FrozenHeight int64    //锁定until账本高度超过
}

// Loads load UTXO item from JSON encoded data
func (item *UtxoItem) Loads(data []byte) error {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	return decoder.Decode(item)
}

// Dumps dump UTXO item into JSON encoded data
func (item *UtxoItem) Dumps() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(item)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
