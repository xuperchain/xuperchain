package bridge

import "testing"

func TestParseCrossChainURI(t *testing.T) {
	testCases := map[string]struct {
		uri    string
		result bool
	}{
		"test1": {
			uri: "xuper://chain1?module=wasm&bcname=xuper&contract_name=counter&method_name=increase",
		},
	}
	for _, v := range testCases {
		res, _ := ParseCrossChainURI(v.uri)
		if res.GetScheme() != "xuper" || res.GetChainName() != "chain1" || res.GetQuery().Get("method_name") != "increase" {
			t.Error("Parse uri error")
		}
	}
}
