package binary

import hex "github.com/tmthrgd/go-hex"

type HexBytes []byte

func (hb *HexBytes) UnmarshalText(hexBytes []byte) error {
	bs, err := hex.DecodeString(string(hexBytes))
	if err != nil {
		return err
	}
	*hb = bs
	return nil
}

func (hb HexBytes) MarshalText() ([]byte, error) {
	return []byte(hb.String()), nil
}

func (hb HexBytes) String() string {
	return hex.EncodeUpperToString(hb)
}

// Protobuf support
func (hb HexBytes) Marshal() ([]byte, error) {
	return hb, nil
}

func (hb *HexBytes) Unmarshal(data []byte) error {
	*hb = data
	return nil
}

func (hb HexBytes) MarshalTo(data []byte) (int, error) {
	return copy(data, hb), nil
}

func (hb HexBytes) Size() int {
	return len(hb)
}

func (hb HexBytes) Bytes() []byte {
	return hb
}
