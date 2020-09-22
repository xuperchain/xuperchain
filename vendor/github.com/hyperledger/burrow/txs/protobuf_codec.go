package txs

import (
	"github.com/hyperledger/burrow/encoding"
)

type protobufCodec struct {
}

func NewProtobufCodec() *protobufCodec {
	return &protobufCodec{}
}

func (gwc *protobufCodec) EncodeTx(env *Envelope) ([]byte, error) {
	return encoding.Encode(env)
}

func (gwc *protobufCodec) DecodeTx(txBytes []byte) (*Envelope, error) {
	env := new(Envelope)
	err := encoding.Decode(txBytes, env)
	if err != nil {
		return nil, err
	}
	return env, nil
}
