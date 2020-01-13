// +build wasm

package driver

import (
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver/wasm"
)

// Serve run contract in wasm environment
func Serve(contract code.Contract) {
	driver := wasm.New()
	driver.Serve(contract)
}
