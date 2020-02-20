package xvm

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xvm/exec"
)

type memCodeProvider struct {
	code []byte
	desc *pb.WasmCodeDesc
}

func (m *memCodeProvider) GetContractCodeDesc(name string) (*pb.WasmCodeDesc, error) {
	return m.desc, nil
}
func (m *memCodeProvider) GetContractCode(name string) ([]byte, error) {
	return m.code, nil
}

type fakeCode struct {
}

func (f *fakeCode) NewContext(cfg *exec.ContextConfig) (exec.Context, error) {
	return nil, nil
}

func (f *fakeCode) Release() {}

func TestGetCacheExecCode(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "xvm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	compileFunc := func(code []byte, output string) error {
		return ioutil.WriteFile(output, code, 0700)
	}

	makeExecCodeFunc := func(libpath string) (exec.Code, error) {
		return new(fakeCode), nil
	}

	cp := &memCodeProvider{
		code: []byte("binary code"),
		desc: &pb.WasmCodeDesc{
			Digest: []byte("digest1"),
		},
	}
	cm, err := newCodeManager(tmpdir, compileFunc, makeExecCodeFunc)
	if err != nil {
		t.Fatal(err)
	}
	code, err := cm.GetExecCode("c1", cp)
	if err != nil {
		t.Fatal(err)
	}
	// 期待从内存中获取
	codeMemory, err := cm.GetExecCode("c1", cp)
	if err != nil {
		t.Fatal(err)
	}
	if code != codeMemory {
		t.Fatalf("expect same exec code:%p, %p", code, codeMemory)
	}

	// digest改变之后需要重新填充cache
	cp.desc.Digest = []byte("digest2")
	code1, _ := cm.GetExecCode("c1", cp)
	if code1 == code {
		t.Fatalf("expect none equal code:%p, %p", code1, code)
	}

	// 期待从磁盘中获取
	cm1, err := newCodeManager(tmpdir, compileFunc, makeExecCodeFunc)
	if err != nil {
		t.Fatal(err)
	}
	codeDisk, err := cm1.GetExecCode("c1", cp)
	if err != nil {
		t.Fatal(err)
	}
	if code1 == codeDisk {
		t.Fatalf("expect none same exec code address:%p, %p", code1, codeMemory)
	}

}

func TestMakeCacheBlocking(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "xvm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	compileFunc := func(code []byte, output string) error {
		time.Sleep(time.Second)
		return ioutil.WriteFile(output, code, 0700)
	}

	makeExecCodeFunc := func(libpath string) (exec.Code, error) {
		return new(fakeCode), nil
	}

	cp := &memCodeProvider{
		code: []byte("binary code"),
		desc: &pb.WasmCodeDesc{
			Digest: []byte("digest1"),
		},
	}
	cm, err := newCodeManager(tmpdir, compileFunc, makeExecCodeFunc)
	if err != nil {
		t.Fatal(err)
	}

	// fill cache
	cm.GetExecCode("c1", cp)
	// making a blocking contract for c2
	go cm.GetExecCode("blocking1", cp)
	c1 := make(chan int)
	go func() {
		// c1 should return immediately
		cm.GetExecCode("c1", cp)
		close(c1)
	}()
	select {
	case <-time.After(100 * time.Millisecond):
		t.Error("wait timeout")
	case <-c1:
	}

	go cm.GetExecCode("blocking2", cp)
	c2 := make(chan int)
	go func() {
		// c1 should return immediately
		cm.GetExecCode("blocking2", cp)
		close(c2)
	}()
	select {
	case <-time.After(100 * time.Millisecond):
	case <-c2:
		t.Error("should block for more than 100ms")
	}

}
