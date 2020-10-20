package balance

type Type uint32

const (
	TypeNative Type = 1
	TypePower  Type = 2
)

var nameFromType = map[Type]string{
	TypeNative: "Native",
	TypePower:  "Power",
}

var typeFromName = make(map[string]Type)

func init() {
	for t, n := range nameFromType {
		typeFromName[n] = t
	}
}

func TypeFromString(name string) Type {
	return typeFromName[name]
}

func (typ Type) String() string {
	name, ok := nameFromType[typ]
	if ok {
		return name
	}
	return "UnknownBalanceType"
}

func (typ Type) MarshalText() ([]byte, error) {
	return []byte(typ.String()), nil
}

func (typ *Type) UnmarshalText(data []byte) error {
	*typ = TypeFromString(string(data))
	return nil
}

// Protobuf support
func (typ Type) Marshal() ([]byte, error) {
	return typ.MarshalText()
}

func (typ *Type) Unmarshal(data []byte) error {
	return typ.UnmarshalText(data)
}
