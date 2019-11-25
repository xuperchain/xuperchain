package xvm

import (
	"io/ioutil"

	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contract/wasm/vm"
	"github.com/xuperchain/xuperunion/xvm/exec"
	"github.com/xuperchain/xuperunion/xvm/runtime/emscripten"
	gowasm "github.com/xuperchain/xuperunion/xvm/runtime/go"
)

type xvmInterpCreator struct {
	cm     *codeManager
	config vm.InstanceCreatorConfig
}

func newXVMInterpCreator(creatorConfig *vm.InstanceCreatorConfig) (vm.InstanceCreator, error) {
	creator := &xvmInterpCreator{
		config: *creatorConfig,
	}
	creator.cm = newCodeManager(creator.config.Basedir,
		creator.compileCode, creator.makeExecCode)
	return creator, nil
}

func (x *xvmInterpCreator) compileCode(buf []byte, outputPath string) error {
	return ioutil.WriteFile(outputPath, buf, 0600)
}

func (x *xvmInterpCreator) makeExecCode(codepath string) (exec.Code, error) {
	codebuf, err := ioutil.ReadFile(codepath)
	if err != nil {
		return nil, err
	}
	resolver := exec.NewMultiResolver(
		gowasm.NewResolver(),
		emscripten.NewResolver(),
		newSyscallResolver(x.config.SyscallService))
	return exec.NewInterpCode(codebuf, resolver)
}

func (x *xvmInterpCreator) CreateInstance(ctx *bridge.Context, cp vm.ContractCodeProvider) (vm.Instance, error) {
	code, err := x.cm.GetExecCode(ctx.ContractName, cp)
	if err != nil {
		return nil, err
	}
	return createInstance(ctx, code, x.config.DebugLogger)
}

func (x *xvmInterpCreator) RemoveCache(contractName string) {
	x.cm.RemoveCode(contractName)
}

func init() {
	vm.Register("ixvm", newXVMInterpCreator)
}
