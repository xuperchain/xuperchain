package wasm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperunion/pb"
)

func makeXvmDeployArgs(t *testing.T) map[string][]byte {
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
		out, _, err := tctx.vmm.DeployContract(tctx.Cache, deployArgs, 0)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", out)
	})
}
