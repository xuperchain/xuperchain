// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package txs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/acm/balance"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/encoding/rlp"
	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/txs/payload"
)

const (
	HashLength    = 32
	HashLengthHex = HashLength * 2
)

// Tx is the canonical object that we serialise to produce the SignBytes that we sign
type Tx struct {
	ChainID string
	payload.Payload
	txHash []byte
}

// Wrap the Payload in Tx required for signing and serialisation
func NewTx(payload payload.Payload) *Tx {
	return &Tx{
		Payload: payload,
	}
}

// Enclose this Tx in an Envelope to be signed
func (tx *Tx) Enclose() *Envelope {
	return &Envelope{
		Tx: tx,
	}
}

// Encloses in Envelope and signs envelope
func (tx *Tx) Sign(signingAccounts ...acm.AddressableSigner) (*Envelope, error) {
	env := tx.Enclose()
	err := env.Sign(signingAccounts...)
	if err != nil {
		return nil, err
	}
	tx.Rehash()
	return env, nil
}

// Generate SignBytes, panicking on any failure
func (tx *Tx) MustSignBytes() []byte {
	bs, err := tx.SignBytes(Envelope_JSON)
	if err != nil {
		panic(err)
	}
	return bs
}

// Produces the canonical SignBytes (the Tx message that will be signed) for a Tx
func (tx *Tx) SignBytes(enc Envelope_EncodingType) ([]byte, error) {
	switch enc {
	case Envelope_JSON:
		bs, err := json.Marshal(tx)
		if err != nil {
			return nil, fmt.Errorf("could not generate canonical SignBytes for Payload %v: %v", tx.Payload, err)
		}
		return bs, nil
	case Envelope_RLP:
		switch pay := tx.Payload.(type) {
		case *payload.CallTx:
			input := pay.Input
			return RLPEncode(
				input.Sequence-1,
				pay.GasPrice,
				pay.GasLimit,
				pay.Address.Bytes(),
				balance.NativeToWei(input.Amount).Bytes(),
				pay.Data.Bytes(),
			)
		default:
			return nil, fmt.Errorf("tx type %v not supported for rlp encoding", tx.Payload.Type())
		}
	default:
		return nil, fmt.Errorf("encoding type %s not supported", enc.String())
	}
}

func RLPEncode(seq, gasPrice, gasLimit uint64, address, amount, data []byte) ([]byte, error) {
	return rlp.Encode([]interface{}{
		seq,       // nonce
		gasPrice,  // gasPrice
		gasLimit,  // gasLimit
		address,   // to
		amount,    // value
		data,      // data
		uint64(1), // chainID
		uint(0), uint(0),
	})
}

// Serialisation intermediate for switching on type
type wrapper struct {
	ChainID string
	Type    payload.Type
	Payload json.RawMessage
}

func (tx *Tx) MarshalJSON() ([]byte, error) {
	bs, err := json.Marshal(tx.Payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(wrapper{
		ChainID: tx.ChainID,
		Type:    tx.Type(),
		Payload: bs,
	})
}

func (tx *Tx) UnmarshalJSON(data []byte) error {
	w := new(wrapper)
	err := json.Unmarshal(data, w)
	if err != nil {
		return err
	}
	tx.ChainID = w.ChainID
	// Now we know the Type we can deserialise the Payload
	tx.Payload, err = payload.New(w.Type)
	if err != nil {
		return err
	}
	return json.Unmarshal(w.Payload, tx.Payload)
}

// Protobuf support
func (tx *Tx) Marshal() ([]byte, error) {
	if tx == nil {
		return nil, nil
	}
	return tx.MarshalJSON()
}

func (tx *Tx) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return tx.UnmarshalJSON(data)
}

func (tx *Tx) MarshalTo(data []byte) (int, error) {
	bs, err := tx.Marshal()
	if err != nil {
		return 0, err
	}
	return copy(data, bs), nil
}

func (tx *Tx) Size() int {
	bs, _ := tx.Marshal()
	return len(bs)
}

func (tx *Tx) Type() payload.Type {
	if tx == nil {
		return payload.TypeUnknown
	}
	return tx.Payload.Type()
}

// Generate a Hash for this transaction based on the SignBytes. The hash is memoized over the lifetime
// of the Tx so repeated calls to Hash() are effectively free
func (tx *Tx) Hash() binary.HexBytes {
	if tx == nil {
		return nil
	}
	if tx.txHash == nil {
		return tx.Rehash()
	}
	return tx.txHash
}

func (tx *Tx) String() string {
	if tx == nil {
		return "Tx{nil}"
	}
	return fmt.Sprintf("Tx{ChainID: %s; TxHash: %s; Payload: %s}", tx.ChainID, tx.Hash(), tx.MustSignBytes())
}

// Regenerate the Tx hash if it has been mutated or as called by Hash() in first instance
func (tx *Tx) Rehash() []byte {
	hasher := sha256.New()
	hasher.Write(tx.MustSignBytes())
	tx.txHash = hasher.Sum(nil)
	tx.txHash = tx.txHash[:HashLength]
	return tx.txHash
}

func (tx *Tx) Get(key string) (interface{}, bool) {
	v, ok := query.GetReflect(reflect.ValueOf(tx), key)
	if ok {
		return v, true
	}
	return query.GetReflect(reflect.ValueOf(tx.Payload), key)
}

// Generate a transaction Receipt containing the Tx hash and other information if the Tx is call.
// Returned by ABCI methods.
func (tx *Tx) GenerateReceipt() *Receipt {
	receipt := &Receipt{
		TxType: tx.Type(),
		TxHash: tx.Hash(),
	}
	if callTx, ok := tx.Payload.(*payload.CallTx); ok {
		receipt.CreatesContract = callTx.Address == nil
		if receipt.CreatesContract {
			receipt.ContractAddress = crypto.NewContractAddress(callTx.Input.Address, tx.Hash())
		} else {
			receipt.ContractAddress = *callTx.Address
		}
	}
	return receipt
}

func DecodeReceipt(bs []byte) (*Receipt, error) {
	receipt := new(Receipt)
	err := encoding.Decode(bs, receipt)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (receipt *Receipt) Encode() ([]byte, error) {
	return encoding.Encode(receipt)
}

func EnvelopeFromAny(chainID string, p *payload.Any) *Envelope {
	if p.CallTx != nil {
		return Enclose(chainID, p.CallTx)
	}
	if p.SendTx != nil {
		return Enclose(chainID, p.SendTx)
	}
	if p.NameTx != nil {
		return Enclose(chainID, p.NameTx)
	}
	if p.PermsTx != nil {
		return Enclose(chainID, p.PermsTx)
	}
	if p.GovTx != nil {
		return Enclose(chainID, p.GovTx)
	}
	if p.ProposalTx != nil {
		return Enclose(chainID, p.ProposalTx)
	}
	if p.BatchTx != nil {
		return Enclose(chainID, p.BatchTx)
	}
	if p.BondTx != nil {
		return Enclose(chainID, p.BondTx)
	}
	if p.UnbondTx != nil {
		return Enclose(chainID, p.UnbondTx)
	}
	if p.IdentifyTx != nil {
		return Enclose(chainID, p.IdentifyTx)
	}
	return nil
}
