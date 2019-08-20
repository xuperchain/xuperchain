package wasm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/pb"
)

func makeXvmDeployArgs(t testing.TB) map[string][]byte {
	codepath := "testdata/counter.wasm"
	if _, err := os.Stat(codepath); err != nil {
		t.Skip()
	}
	codebuf := loadWasmBinary(t, codepath)
	desc := &pb.WasmCodeDesc{
		Runtime:  "c",
		Compiler: "emcc",
	}
	descbuf, _ := proto.Marshal(desc)

	args := map[string][]byte{
		"creator": []byte("icexin"),
	}
	argsbuf, _ := json.Marshal(args)
	return map[string][]byte{
		"contract_name": []byte("counter"),
		"contract_code": codebuf,
		"init_args":     argsbuf,
		"contract_desc": descbuf,
	}
}

func TestXvmDeploy(t *testing.T) {
	WithTestContext(t, "xvm", func(tctx *FakeWASMContext) {
		deployArgs := makeXvmDeployArgs(t)
		out, _, err := tctx.vmm.DeployContract(&contract.ContextConfig{
			XMCache:        tctx.Cache,
			ResourceLimits: contract.MaxLimits,
		}, deployArgs)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%v", out)
	})
}

func BenchmarkXVMInvoke(b *testing.B) {
	WithTestContext(b, "xvm", func(tctx *FakeWASMContext) {
		deployArgs := makeXvmDeployArgs(b)
		_, _, err := tctx.vmm.DeployContract(&contract.ContextConfig{
			XMCache:        tctx.Cache,
			ResourceLimits: contract.MaxLimits,
		}, deployArgs)
		if err != nil {
			b.Fatal(err)
		}
		err = tctx.CommitCache()
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx, err := tctx.vm.NewContext(&contract.ContextConfig{
				ContractName:   "counter",
				ResourceLimits: contract.MaxLimits,
				XMCache:        tctx.Cache,
			})
			if err != nil {
				b.Fatal(err)
			}
			_, err = ctx.Invoke("increase", map[string][]byte{"key": []byte("icexin")})
			if err != nil {
				b.Fatal(err)
			}
			ctx.Release()
		}
	})
}
