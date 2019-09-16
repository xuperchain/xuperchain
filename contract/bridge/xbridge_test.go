package bridge

import (
	"context"
	"testing"

	"github.com/xuperchain/xuperunion/contract"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
	"github.com/xuperchain/xuperunion/test/util"
)

type codeExecutor struct {
	syscall *SyscallService
}

func (c *codeExecutor) RegisterSyscallService(syscall *SyscallService) {
	c.syscall = syscall
}

func (c *codeExecutor) NewInstance(ctx *Context) (Instance, error) {
	return &codeInstance{
		ctx:     ctx,
		syscall: c.syscall,
	}, nil
}

type codeInstance struct {
	ctx     *Context
	syscall *SyscallService
}

func (c *codeInstance) Exec() error {
	switch c.ctx.Method {
	case "TestMethod":
		c.ctx.Output = &pb.Response{Body: []byte(c.ctx.ContractName + ":" + c.ctx.Method)}
		return nil
	case "Echo":
		c.ctx.Output = &pb.Response{Body: []byte("hello:" + string(c.ctx.Args["hello"]))}
		return nil
	case "Put":
		output, err := c.testPut(c.ctx.Args)
		if err != nil {
			return err
		}
		c.ctx.Output = &pb.Response{Body: output}
	}
	return nil
}

func (c *codeInstance) ResourceUsed() contract.Limits {
	return contract.Limits{}
}

func (c *codeInstance) Release() {
}

func (c *codeInstance) testPut(args map[string][]byte) ([]byte, error) {
	{
		_, err := c.syscall.PutObject(context.TODO(), &pb.PutRequest{
			Header: &pb.SyscallHeader{
				Ctxid: c.ctx.ID,
			},
			Key:   args["key"],
			Value: args["value"],
		})
		if err != nil {
			return nil, err
		}
	}
	{
		resp, err := c.syscall.GetObject(context.TODO(), &pb.GetRequest{
			Header: &pb.SyscallHeader{
				Ctxid: c.ctx.ID,
			},
			Key: args["key"],
		})
		if err != nil {
			return nil, err
		}
		return resp.Value, nil
	}
}

func TestExecutorMethod(t *testing.T) {
	xbridge := New()
	vm := xbridge.RegisterExecutor("code", new(codeExecutor))
	util.WithXModelContext(t, func(x *util.XModelContext) {
		ctxCfg := &contract.ContextConfig{
			XMCache:        x.Cache,
			Initiator:      "",
			AuthRequire:    []string{},
			ContractName:   "dummy",
			ResourceLimits: contract.MaxLimits,
		}

		ctx, err := vm.NewContext(ctxCfg)
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Release()
		args := map[string][]byte{}
		resp, err := ctx.Invoke("TestMethod", args)
		if err != nil {
			t.Fatal(err)
		}
		if string(resp.Body) != "dummy:TestMethod" {
			t.Errorf("expect dummy:TestMethod, got `%s`", resp.Body)
		}
	})
}

func TestExecutorArgs(t *testing.T) {
	xbridge := New()
	vm := xbridge.RegisterExecutor("code", new(codeExecutor))
	util.WithXModelContext(t, func(x *util.XModelContext) {

		ctxCfg := &contract.ContextConfig{
			XMCache:        x.Cache,
			Initiator:      "",
			AuthRequire:    []string{},
			ContractName:   "dummy",
			ResourceLimits: contract.MaxLimits,
		}

		ctx, err := vm.NewContext(ctxCfg)
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Release()
		args := map[string][]byte{
			"hello": []byte("world"),
		}
		resp, err := ctx.Invoke("Echo", args)
		if err != nil {
			t.Fatal(err)
		}
		if string(resp.Body) != "hello:world" {
			t.Errorf("expect hello:world, got `%s`", resp.Body)
		}
	})
}

func TestExecutorSyscall(t *testing.T) {
	xbridge := New()
	vm := xbridge.RegisterExecutor("code", new(codeExecutor))
	util.WithXModelContext(t, func(x *util.XModelContext) {
		ctxCfg := &contract.ContextConfig{
			XMCache:        x.Cache,
			Initiator:      "",
			AuthRequire:    []string{},
			ContractName:   "dummy",
			ResourceLimits: contract.MaxLimits,
		}

		ctx, err := vm.NewContext(ctxCfg)
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Release()
		args := map[string][]byte{
			"key":   []byte("hello"),
			"value": []byte("world"),
		}
		resp, err := ctx.Invoke("Put", args)
		if err != nil {
			t.Fatal(err)
		}
		if string(resp.Body) != "world" {
			t.Errorf("expect world, got `%s`", resp.Body)
		}
	})
}
