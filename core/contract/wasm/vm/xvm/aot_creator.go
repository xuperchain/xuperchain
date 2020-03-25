package xvm

import (
	"fmt"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"path/filepath"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	"github.com/xuperchain/xuperchain/core/contract/teevm"
	"github.com/xuperchain/xuperchain/core/contract/wasm/vm"
	"github.com/xuperchain/xuperchain/core/xvm/compile"
	"github.com/xuperchain/xuperchain/core/xvm/exec"
	"github.com/xuperchain/xuperchain/core/xvm/runtime/emscripten"
	gowasm "github.com/xuperchain/xuperchain/core/xvm/runtime/go"
)

type xvmCreator struct {
	cm       *codeManager
	config   vm.InstanceCreatorConfig
	vmconfig config.XVMConfig

	wasm2cPath string
}

// 优先查找跟xchain同级目录的二进制，再在PATH里面找
func lookupWasm2c() (string, error) {
	// 首先查找跟xchain同级的目录
	wasm2cPath := filepath.Join(filepath.Dir(os.Args[0]), "wasm2c")
	stat, err := os.Stat(wasm2cPath)
	if err == nil {
		if m := stat.Mode(); !m.IsDir() && m&0111 != 0 {
			return filepath.Abs(wasm2cPath)
		}
	}
	// 再查找系统PATH目录
	return osexec.LookPath("wasm2c")
}

func newXVMCreator(creatorConfig *vm.InstanceCreatorConfig) (vm.InstanceCreator, error) {
	wasm2cPath, err := lookupWasm2c()
	if err != nil {
		return nil, err
	}
	creator := &xvmCreator{
		wasm2cPath: wasm2cPath,
		config:     *creatorConfig,
	}
	if creatorConfig.VMConfig != nil {
		creator.vmconfig = creatorConfig.VMConfig.(config.XVMConfig)
		optlevel := creator.vmconfig.OptLevel
		if optlevel < 0 || optlevel > 3 {
			return nil, fmt.Errorf("bad xvm optlevel:%d", optlevel)
		}
	}
	creator.cm, err = newCodeManager(creator.config.Basedir,
		creator.CompileCode, creator.MakeExecCode)
	if err != nil {
		return nil, err
	}
	return creator, nil
}

func cpfile(dest, src string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, buf, 0700)
}

func (x *xvmCreator) CompileCode(buf []byte, outputPath string) error {
	tmpdir, err := ioutil.TempDir("", "xvm-compile")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)
	wasmpath := filepath.Join(tmpdir, "code.wasm")
	err = ioutil.WriteFile(wasmpath, buf, 0600)
	if err != nil {
		return err
	}

	libpath := filepath.Join(tmpdir, "code.so")

	cfg := &compile.Config{
		Wasm2cPath: x.wasm2cPath,
		OptLevel:   x.vmconfig.OptLevel,
	}
	err = compile.CompileNativeLibrary(cfg, libpath, wasmpath)
	if err != nil {
		return err
	}
	return cpfile(outputPath, libpath)
}

func (x *xvmCreator) getContractCodeCache(name string, cp vm.ContractCodeProvider) (*contractCode, error) {
	return x.cm.GetExecCode(name, cp)
}

func (x *xvmCreator) MakeExecCode(libpath string) (exec.Code, error) {
	resolvers := []exec.Resolver{
		gowasm.NewResolver(),
		emscripten.NewResolver(),
		newSyscallResolver(x.config.SyscallService),
		builtinResolver,
	}
	//AOT only for experiment;
	if x.config.TEEConfig.Enable {
		teeResolver, err := teevm.NewTrustFunctionResolver(x.config.TEEConfig)
		if err != nil {
			return nil, err
		}
		resolvers = append(resolvers, teeResolver)
	}
	resolver := exec.NewMultiResolver(
		resolvers...,
	)
	return exec.NewAOTCode(libpath, resolver)
}

func (x *xvmCreator) CreateInstance(ctx *bridge.Context, cp vm.ContractCodeProvider) (vm.Instance, error) {
	code, err := x.getContractCodeCache(ctx.ContractName, cp)
	if err != nil {
		log.Error("get contract cache error", "error", err, "contract", ctx.ContractName)
		return nil, err
	}

	return createInstance(ctx, code, x.config.DebugLogger)
}

func (x *xvmCreator) RemoveCache(contractName string) {
	x.cm.RemoveCode(contractName)
}

func init() {
	vm.Register("xvm", newXVMCreator)
}
