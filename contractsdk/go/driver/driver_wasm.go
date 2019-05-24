// +build wasm

package driver

import (
	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/driver/wasm"
)

// Serve run contract in wasm environment
func Serve(contract code.Contract) {
	driver := wasm.New()
	driver.Serve(contract)
}
