package abi

import (
	"encoding/json"
	"reflect"

	"github.com/hyperledger/burrow/event/query"
	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/sha3"
)

// Argument is a decoded function parameter, return or event field
type Argument struct {
	Name        string
	EVM         EVMType
	IsArray     bool
	Indexed     bool
	Hashed      bool
	ArrayLength uint64
}

type argumentJSON struct {
	Name       string
	Type       string
	Components []argumentJSON
	Indexed    bool
}

// EventIDSize is the length of the event selector
const EventIDSize = 32

type EventSpec struct {
	ID        EventID
	Inputs    []Argument
	Name      string
	Anonymous bool
}

func (e *EventSpec) Get(key string) (interface{}, bool) {
	return query.GetReflect(reflect.ValueOf(e), key)
}

func (e *EventSpec) UnmarshalJSON(data []byte) error {
	s := new(specJSON)
	err := json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	return e.unmarshalSpec(s)
}

func (e *EventSpec) unmarshalSpec(s *specJSON) error {
	inputs, err := readArgSpec(s.Inputs)
	if err != nil {
		return err
	}
	// Get signature before we deal with hashed types
	sig := Signature(s.Name, inputs)
	for i := range inputs {
		if inputs[i].Indexed && inputs[i].EVM.Dynamic() {
			// For Dynamic types, the hash is stored in stead
			inputs[i].EVM = EVMBytes{M: 32}
			inputs[i].Hashed = true
		}
	}
	e.Name = s.Name
	e.ID = GetEventID(sig)
	e.Inputs = inputs
	e.Anonymous = s.Anonymous
	return nil
}

type EventID [EventIDSize]byte

func GetEventID(signature string) (id EventID) {
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(signature))
	copy(id[:], hash.Sum(nil))
	return
}

func (e *EventSpec) String() string {
	str := e.Name + argsToSignature(e.Inputs, true)
	if e.Anonymous {
		str += " anonymous"
	}

	return str
}

func (id EventID) String() string {
	return hex.EncodeUpperToString(id[:])
}

func (id EventID) Bytes() []byte {
	return id[:]
}
