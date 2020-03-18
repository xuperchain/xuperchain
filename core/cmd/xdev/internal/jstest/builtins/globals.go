package builtins

import (
	"encoding/base64"

	"github.com/robertkrimen/otto"
)

// Throw throw go error in js vm as an Exception
func throw(err error) {
	v, _ := otto.ToValue("Exception: " + err.Error())
	panic(v)
}

func bytesToString(s []byte) string {
	return string(s)
}

func btoa(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func atob(s string) string {
	ret, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		throw(err)
	}
	return string(ret)
}

// Globals will register to global object
var Globals = map[string]interface{}{
	"string": bytesToString,
	"btoa":   btoa,
	"atob":   atob,
}
