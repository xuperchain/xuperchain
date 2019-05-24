// +build !wasm

package driver

import (
	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver/native"
)

// Serve run contract in native environment
func Serve(contract code.Contract) {
	driver := native.New()
	driver.Serve(contract)
}
