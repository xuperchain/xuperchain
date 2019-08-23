package wasm

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/wasm/vm/memory"
	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/xmodel"
)

type contractCode struct {
}

func (c *contractCode) Initialize(ctx code.Context) code.Response {
	creator := ctx.Args()["creator"]
	err := ctx.PutObject([]byte("creator"), []byte(creator))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(nil)
}

func (c *contractCode) Invoke(ctx code.Context) code.Response {
	key, ok := ctx.Args()["key"]
	if !ok {
		return code.Errors("missing key")
	}
	value, err := ctx.GetObject([]byte(key))
	cnt := 0
	if err == nil {
		cnt, _ = strconv.Atoi(string(value))
	}

	cntstr := strconv.Itoa(cnt + 1)

	err = ctx.PutObject([]byte(key), []byte(cntstr))
	if err != nil {
		return code.Error(err)
	}
	return code.Response{
		Status:  200,
		Message: cntstr,
	}
}

func (c *contractCode) Query(ctx code.Context) code.Response {
	return code.OK(nil)
}

func makeDeployArgs(t *testing.T) map[string][]byte {
	codebuf := memory.Encode(new(contractCode))
	desc := &pb.WasmCodeDesc{
		Runtime:  "go",
		Compiler: "none",
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

func TestWasmDeploy(t *testing.T) {
	WithTestContext(t, "memory", func(tctx *FakeWASMContext) {
		deployArgs := makeDeployArgs(t)
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

func TestWasmInvoke(t *testing.T) {
	WithTestContext(t, "memory", func(tctx *FakeWASMContext) {
		deployArgs := makeDeployArgs(t)
		_, _, err := tctx.vmm.DeployContract(&contract.ContextConfig{
			XMCache:        tctx.Cache,
			ResourceLimits: contract.MaxLimits,
		}, deployArgs)
		if err != nil {
			t.Fatal(err)
		}
		err = tctx.CommitCache()
		if err != nil {
			t.Fatal(err)
		}
		{
			cache, err := xmodel.NewXModelCache(tctx.Model, true)
			if err != nil {
				t.Fatal(err)
			}
			ctxCfg := &contract.ContextConfig{
				XMCache:        cache,
				Initiator:      "",
				AuthRequire:    []string{},
				ContractName:   "counter",
				ResourceLimits: contract.MaxLimits,
			}
			ctx, err := tctx.vm.NewContext(ctxCfg)
			if err != nil {
				t.Fatal(err)
			}
			defer ctx.Release()
			out, err := ctx.Invoke("invoke", map[string][]byte{
				"key": []byte("mycounter"),
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("out:%v", out)
		}
	})
}

func TestWasmContractMissing(t *testing.T) {
	WithTestContext(t, "memory", func(tctx *FakeWASMContext) {
		ctxCfg := &contract.ContextConfig{
			XMCache:        tctx.Cache,
			Initiator:      "",
			AuthRequire:    []string{},
			ContractName:   "counter",
			ResourceLimits: contract.MaxLimits,
		}
		ctx, err := tctx.vm.NewContext(ctxCfg)
		if err == nil {
			ctx.Release()
			t.Fatal("expect none nil error, go nil")
			return
		}
	})
}
