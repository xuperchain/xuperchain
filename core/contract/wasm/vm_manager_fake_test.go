package wasm

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/xuperchain/xuperunion/common/config"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/test/util"
)

type FakeWASMContext struct {
	*util.XModelContext
	vmm *VMManager
	vm  contract.VirtualMachine
}

func WithTestContext(t testing.TB, driver string, callback func(tctx *FakeWASMContext)) {
	util.WithXModelContext(t, func(x *util.XModelContext) {
		basedir := filepath.Join(x.Basedir, "wasm")
		xbridge := bridge.New()
		vmm, err := New(&config.WasmConfig{
			Driver: driver,
			XVM: config.XVMConfig{
				OptLevel: 0,
			},
		}, basedir, xbridge, x.Model)
		if err != nil {
			t.Fatal(err)
		}
		exec := xbridge.RegisterExecutor("wasm", vmm)

		callback(&FakeWASMContext{
			vmm:           vmm,
			vm:            exec,
			XModelContext: x,
		})

	})
}

func loadWasmBinary(t testing.TB, filepath string) []byte {
	by, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}
	return by
}
