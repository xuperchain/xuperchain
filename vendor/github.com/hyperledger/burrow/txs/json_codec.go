package txs

import (
	"encoding/json"
)

type jsonCodec struct{}

func NewJSONCodec() *jsonCodec {
	return &jsonCodec{}
}

func (gwc *jsonCodec) EncodeTx(env *Envelope) ([]byte, error) {
	return json.Marshal(env)
}

func (gwc *jsonCodec) DecodeTx(txBytes []byte) (*Envelope, error) {
	txEnv := new(Envelope)
	err := json.Unmarshal(txBytes, txEnv)
	if err != nil {
		return nil, err
	}
	return txEnv, nil
}
