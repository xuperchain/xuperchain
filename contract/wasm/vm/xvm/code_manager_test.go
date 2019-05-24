package xvm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/xvm/exec"
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

func TestGetCacheExecCode(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "xvm-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	compileFunc := func(code []byte, output string) error {
		return ioutil.WriteFile(output, code, 0700)
	}

	makeExecCodeFunc := func(libpath string) (*exec.Code, error) {
		return new(exec.Code), nil
	}

	cp := &memCodeProvider{
		code: []byte("binary code"),
		desc: &pb.WasmCodeDesc{
			Digest: []byte("digest1"),
		},
	}
	cm := newCodeManager(tmpdir, compileFunc, makeExecCodeFunc)
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
	cm1 := newCodeManager(tmpdir, compileFunc, makeExecCodeFunc)
	codeDisk, err := cm1.GetExecCode("c1", cp)
	if err != nil {
		t.Fatal(err)
	}
	if code1 == codeDisk {
		t.Fatalf("expect none same exec code address:%p, %p", code1, codeMemory)
	}

}
