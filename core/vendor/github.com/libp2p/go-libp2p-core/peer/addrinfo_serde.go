package peer

import (
	"encoding/json"

	ma "github.com/multiformats/go-multiaddr"
)

func (pi AddrInfo) MarshalJSON() ([]byte, error) {
	out := make(map[string]interface{})
	out["ID"] = pi.ID.Pretty()
	var addrs []string
	for _, a := range pi.Addrs {
		addrs = append(addrs, a.String())
	}
	out["Addrs"] = addrs
	return json.Marshal(out)
}

func (pi *AddrInfo) UnmarshalJSON(b []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	pid, err := IDB58Decode(data["ID"].(string))
	if err != nil {
		return err
	}
	pi.ID = pid
	addrs, ok := data["Addrs"].([]interface{})
	if ok {
		for _, a := range addrs {
			pi.Addrs = append(pi.Addrs, ma.StringCast(a.(string)))
		}
	}
	return nil
}
