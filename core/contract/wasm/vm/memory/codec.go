package memory

import (
	"bytes"
	"encoding/gob"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
)

// Encode encodes a contract handler to bytes which can be later Decoded to contract
func Encode(contract code.Contract) []byte {
	gob.Register(contract)
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(&contract)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// Decode decodes bytes to contract
// The underlying type must be known to Decode function
func Decode(buf []byte) (code.Contract, error) {
	var contract code.Contract
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	err := dec.Decode(&contract)
	if err != nil {
		return nil, err
	}
	return contract, nil
}
