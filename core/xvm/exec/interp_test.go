package exec

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuperchain/xuperchain/core/xvm/compile"
)

func withInterpCode(t testing.TB, watCode string, r Resolver, f func(code *InterpCode)) {
	tmpdir, err := ioutil.TempDir("", "xvm-exec-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	wasmpath := filepath.Join(tmpdir, "wasm.wasm")
	cfg := &compile.Config{
		Wasm2cPath:   "../compile/wabt/build/wasm2c",
		Wat2wasmPath: "../compile/wabt/build/wat2wasm",
		OptLevel:     0,
	}

	err = compile.CompileWatSource(cfg, wasmpath, watCode)
	if err != nil {
		t.Fatal(err)
	}

	codebuf, err := ioutil.ReadFile(wasmpath)
	if err != nil {
		t.Fatal(err)
	}
	code, err := NewInterpCode(codebuf, r)
	if err != nil {
		t.Fatal(err)
	}
	f(code)
}

func TestNewInterContext(t *testing.T) {
	withInterpCode(t, "testdata/add.wat", nil, func(code *InterpCode) {
		ctx, err := code.NewContext(DefaultContextConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Release()
		ret, err := ctx.Exec("_add", []int64{1, 2})
		if err != nil {
			t.Fatal(err)
		}
		if ret != 3 {
			t.Errorf("expect 3 got %d", ret)
		}
	})
}

func TestInterpResolveFunc(t *testing.T) {
	r := MapResolver(map[string]interface{}{
		"env._print": func(ctx Context, addr uint32) uint32 {
			c := NewCodec(ctx)
			if c.CString(addr) != "hello world" {
				panic("not equal")
			}
			return 0
		},
	})
	withInterpCode(t, "testdata/extern_func.wat", r, func(code *InterpCode) {
		ctx, err := code.NewContext(DefaultContextConfig())
		if err != nil {
			t.Fatal(err)
		}
		_, err = ctx.Exec("_run", nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		ctx.Release()
	})
}
