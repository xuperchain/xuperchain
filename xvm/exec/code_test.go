package exec

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuperchain/xuperunion/xvm/compile"
)

func withCode(t testing.TB, watCode string, r Resolver, f func(code Code)) {
	tmpdir, err := ioutil.TempDir("", "xvm-exec-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	wasmpath := filepath.Join(tmpdir, "wasm.wasm")
	libpath := filepath.Join(tmpdir, "wasm.so")
	cfg := &compile.Config{
		Wasm2cPath:   "../compile/wabt/build/wasm2c",
		Wat2wasmPath: "../compile/wabt/build/wat2wasm",
		OptLevel:     0,
	}

	err = compile.CompileWatSource(cfg, wasmpath, watCode)
	if err != nil {
		t.Fatal(err)
	}

	err = compile.CompileNativeLibrary(cfg, libpath, wasmpath)
	if err != nil {
		t.Fatal(err)
	}
	code, err := NewAOTCode(libpath, r)
	if err != nil {
		t.Fatal(err)
	}
	f(code)
	code.Release()
}

func TestNewCode(t *testing.T) {
	withCode(t, "testdata/sum.wat", nil, func(code Code) {
	})
}
