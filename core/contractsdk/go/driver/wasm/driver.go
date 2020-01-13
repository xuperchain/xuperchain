// +build wasm

package wasm

import (
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/exec"
)

type driver struct {
}

// New returns a wasm driver
func New() code.Driver {
	return new(driver)
}

func (d *driver) Serve(contract code.Contract) {
	initDebugLog()
	exec.RunContract(0, contract, syscall)
}
