package exec

type CallType uint32

const (
	CallTypeCall     = CallType(0x00)
	CallTypeCode     = CallType(0x01)
	CallTypeDelegate = CallType(0x02)
	CallTypeStatic   = CallType(0x03)
)

var nameFromCallType = map[CallType]string{
	CallTypeCall:     "Call",
	CallTypeCode:     "CallCode",
	CallTypeDelegate: "DelegateCall",
	CallTypeStatic:   "StaticCall",
}

var callTypeFromName = make(map[string]CallType)

func init() {
	for t, n := range nameFromCallType {
		callTypeFromName[n] = t
	}
}

func CallTypeFromString(name string) CallType {
	return callTypeFromName[name]
}

func (ct CallType) String() string {
	name, ok := nameFromCallType[ct]
	if ok {
		return name
	}
	return "UnknownCallType"
}

func (ct CallType) MarshalText() ([]byte, error) {
	return []byte(ct.String()), nil
}

func (ct *CallType) UnmarshalText(data []byte) error {
	*ct = CallTypeFromString(string(data))
	return nil
}
