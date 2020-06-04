package bridge

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/test/util"
)

type counter struct{}

func (c *counter) Initialize(ctx code.Context) code.Response {
	creator, ok := ctx.Args()["creator"]
	if !ok {
		return code.Errors("missing creator")
	}
	ctx.Logf("creator:%s", creator)
	err := ctx.PutObject([]byte("creator"), creator)
	if err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok"))
}

func (c *counter) Increase(ctx code.Context) code.Response {
	key, ok := ctx.Args()["key"]
	if !ok {
		return code.Errors("missing key")
	}
	value, err := ctx.GetObject(key)
	cnt := 0
	if err == nil {
		cnt, _ = strconv.Atoi(string(value))
	}

	cntstr := strconv.Itoa(cnt + 1)

	err = ctx.PutObject(key, []byte(cntstr))
	if err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(cntstr))
}

func (c *counter) Get(ctx code.Context) code.Response {
	key, ok := ctx.Args()["key"]
	if !ok {
		return code.Errors("missing key")
	}
	value, err := ctx.GetObject(key)
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)
}

type newCounter struct {
}

func (c *newCounter) Initialize(ctx code.Context) code.Response {
	return code.OK([]byte("ok"))
}

func (c *newCounter) Increase(ctx code.Context) code.Response {
	return code.OK([]byte("0"))
}

type testHelper struct {
	bridge *XBridge
	model  *util.XModelContext
}

func newTestHelper(model *util.XModelContext) (*testHelper, error) {
	bridge, err := New(&XBridgeConfig{
		Basedir:   model.Basedir,
		XModel:    model.Model,
		LogWriter: os.Stderr,
		VMConfigs: map[ContractType]VMConfig{
			TypeNative: &config.NativeConfig{
				Enable: true,
				Driver: "memory",
			},
		},
		Config: config.ContractConfig{
			EnableUpgrade: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return &testHelper{
		bridge: bridge,
		model:  model,
	}, nil
}

func (t *testHelper) DeployContract(name string, c code.Contract, initArgs map[string][]byte) (*contract.Response, error) {
	codebuf := memoryEncode(c)
	desc := &pb.WasmCodeDesc{
		ContractType: "native",
	}
	descbuf, _ := proto.Marshal(desc)
	argsbuf, _ := json.Marshal(initArgs)
	deployArgs := map[string][]byte{
		"contract_name": []byte(name),
		"contract_code": codebuf,
		"init_args":     argsbuf,
		"contract_desc": descbuf,
	}

	resp, _, err := t.bridge.DeployContract(&contract.ContextConfig{
		XMCache:        t.model.Cache,
		ResourceLimits: contract.MaxLimits,
	}, deployArgs)

	if err != nil {
		return nil, err
	}
	if resp.Status == 200 {
		t.model.CommitCache()
	}
	return resp, nil
}

func (t *testHelper) UpgradeContract(name string, c code.Contract) error {
	codebuf := memoryEncode(c)
	upgradeArgs := map[string][]byte{
		"contract_name": []byte(name),
		"contract_code": codebuf,
	}

	_, _, err := t.bridge.UpgradeContract(&contract.ContextConfig{
		XMCache:        t.model.Cache,
		ResourceLimits: contract.MaxLimits,
	}, upgradeArgs)

	if err != nil {
		return err
	}
	t.model.CommitCache()
	return nil
}

func (t *testHelper) InvokeContract(cname, method string, args map[string][]byte) (*contract.Response, error) {
	ctxCfg := &contract.ContextConfig{
		XMCache:        t.model.Cache,
		Initiator:      "",
		AuthRequire:    []string{},
		ContractName:   cname,
		ResourceLimits: contract.MaxLimits,
		CanInitialize:  false,
	}
	vm, _ := t.bridge.GetVirtualMachine(string(TypeNative))
	ctx, err := vm.NewContext(ctxCfg)
	if err != nil {
		return nil, err
	}
	defer ctx.Release()
	resp, err := ctx.Invoke(method, args)
	if err != nil {
		return nil, err
	}
	if resp.Status == 200 {
		t.model.CommitCache()
	}
	return resp, nil
}

func TestDeploy(t *testing.T) {
	util.WithXModelContext(t, func(model *util.XModelContext) {
		helper, err := newTestHelper(model)
		if err != nil {
			t.Fatal(err)
		}

		contract := new(counter)
		resp, err := helper.DeployContract("counter", contract, map[string][]byte{
			"creator": []byte("icexin"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if string(resp.Body) != "ok" {
			t.Errorf("expect ok got %s", resp.Body)
		}
	})
}

func TestUpgrade(t *testing.T) {
	util.WithXModelContext(t, func(model *util.XModelContext) {
		helper, err := newTestHelper(model)
		if err != nil {
			t.Fatal(err)
		}

		contract := new(counter)
		_, err = helper.DeployContract("counter", contract, map[string][]byte{
			"creator": []byte("icexin"),
		})
		if err != nil {
			t.Fatal(err)
		}

		newContract := new(newCounter)
		err = helper.UpgradeContract("counter", newContract)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := helper.InvokeContract("counter", "increase", map[string][]byte{
			"key": []byte("icexin"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if string(resp.Body) != "0" {
			t.Fatalf("expect 0 got %s", resp.Body)
		}
	})
}

func TestInvoke(t *testing.T) {
	util.WithXModelContext(t, func(model *util.XModelContext) {
		helper, err := newTestHelper(model)
		if err != nil {
			t.Fatal(err)
		}

		contract := new(counter)
		_, err = helper.DeployContract("counter", contract, map[string][]byte{
			"creator": []byte("icexin"),
		})
		if err != nil {
			t.Fatal(err)
		}

		resp, err := helper.InvokeContract("counter", "increase", map[string][]byte{
			"key": []byte("icexin"),
		})
		if err != nil {
			t.Fatal(err)
		}
		if resp.Status != 200 {
			t.Fatal(resp.Message)
		}
		if string(resp.Body) != "1" {
			t.Errorf("expect 1 got %s", resp.Body)
		}
	})
}
