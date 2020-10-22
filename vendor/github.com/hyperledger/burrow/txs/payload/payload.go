package payload

import (
	"fmt"
	"strings"
)

/*
Payload (Transaction) is an atomic operation on the ledger state.

Account Txs:
 - SendTx         Send coins to address
 - CallTx         Send a msg to a contract that runs in the vm
 - NameTx	  Store some value under a name in the global namereg

Validation Txs:
 - BondTx         New validator posts a bond
 - UnbondTx       Validator leaves

Admin Txs:
 - PermsTx
*/

type Type uint32

// Types of Payload implementations
const (
	TypeUnknown = Type(0x00)
	// Account transactions
	TypeSend  = Type(0x01)
	TypeCall  = Type(0x02)
	TypeName  = Type(0x03)
	TypeBatch = Type(0x04)

	// Validation transactions
	TypeBond   = Type(0x11)
	TypeUnbond = Type(0x12)

	// Admin transactions
	TypePermissions = Type(0x21)
	TypeGovernance  = Type(0x22)
	TypeProposal    = Type(0x23)
	TypeIdentify    = Type(0x24)
)

type Payload interface {
	String() string
	GetInputs() []*TxInput
	Type() Type
	Any() *Any
	// The serialised size in bytes
	Size() int
}

var nameFromType = map[Type]string{
	TypeUnknown:     "UnknownTx",
	TypeSend:        "SendTx",
	TypeCall:        "CallTx",
	TypeName:        "NameTx",
	TypeBatch:       "BatchTx",
	TypePermissions: "PermsTx",
	TypeGovernance:  "GovTx",
	TypeProposal:    "ProposalTx",
	TypeBond:        "BondTx",
	TypeUnbond:      "UnbondTx",
	TypeIdentify:    "IdentifyTx",
}

var typeFromName = make(map[string]Type)

func init() {
	for t, n := range nameFromType {
		typeFromName[n] = t
	}
}

func TxTypeFromString(name string) Type {
	return typeFromName[name]
}

func (typ Type) String() string {
	name, ok := nameFromType[typ]
	if ok {
		return name
	}
	return "UnknownTx"
}

func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.String()), nil
}

func (typ *Type) UnmarshalText(data []byte) error {
	*typ = TxTypeFromString(string(data))
	return nil
}

// Protobuf support
func (typ Type) Marshal() ([]byte, error) {
	return typ.MarshalText()
}

func (typ *Type) Unmarshal(data []byte) error {
	return typ.UnmarshalText(data)
}

func InputsString(inputs []*TxInput) string {
	strs := make([]string, len(inputs))
	for i, in := range inputs {
		strs[i] = in.Address.String()
	}
	return strings.Join(strs, ",")
}

func New(txType Type) (Payload, error) {
	switch txType {
	case TypeSend:
		return &SendTx{}, nil
	case TypeCall:
		return &CallTx{}, nil
	case TypeName:
		return &NameTx{}, nil
	case TypeBatch:
		return &BatchTx{}, nil
	case TypePermissions:
		return &PermsTx{}, nil
	case TypeGovernance:
		return &GovTx{}, nil
	case TypeBond:
		return &BondTx{}, nil
	case TypeUnbond:
		return &UnbondTx{}, nil
	case TypeProposal:
		return &ProposalTx{}, nil
	case TypeIdentify:
		return &IdentifyTx{}, nil
	}
	return nil, fmt.Errorf("unknown payload type: %d", txType)
}
