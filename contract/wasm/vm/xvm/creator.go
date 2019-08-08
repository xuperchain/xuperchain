package xvm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"path/filepath"

	"github.com/xuperchain/xuperunion/common/config"
	"github.com/xuperchain/xuperunion/common/log"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contract/wasm/vm"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/xvm/compile"
	"github.com/xuperchain/xuperunion/xvm/debug"
	"github.com/xuperchain/xuperunion/xvm/exec"
	"github.com/xuperchain/xuperunion/xvm/runtime/emscripten"
	gowasm "github.com/xuperchain/xuperunion/xvm/runtime/go"
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
	creator.cm = newCodeManager(creator.config.Basedir,
		creator.CompileCode, creator.MakeExecCode)
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

func (x *xvmCreator) MakeExecCode(libpath string) (*exec.Code, error) {
	resolver := exec.NewMultiResolver(
		gowasm.NewResolver(),
		emscripten.NewResolver(),
		newSyscallResolver(x.config.SyscallService))
	return exec.NewCode(libpath, resolver)
}

func (x *xvmCreator) CreateInstance(ctx *bridge.Context, cp vm.ContractCodeProvider) (vm.Instance, error) {
	code, err := x.getContractCodeCache(ctx.ContractName, cp)
	if err != nil {
		log.Error("get contract cache error", "error", err, "contract", ctx.ContractName)
		return nil, err
	}

	log.Info("instance resource limit", "limits", ctx.ResourceLimits)
	execCtx, err := exec.NewContext(code.ExecCode, &exec.ContextConfig{
		GasLimit: ctx.ResourceLimits.Cpu,
	})
	if err != nil {
		log.Error("create contract context error", "error", err, "contract", ctx.ContractName)
		return nil, err
	}
	switch code.Desc.GetRuntime() {
	case "go":
		gowasm.RegisterRuntime(execCtx)
	case "c":
		emscripten.Init(execCtx)
	}
	execCtx.SetUserData(contextIDKey, ctx.ID)
	instance := &xvmInstance{
		bridgeCtx: ctx,
		execCtx:   execCtx,
		desc:      code.Desc,
	}
	instance.InitDebugWriter(x.config.DebugLogger)
	return instance, nil
}

func (x *xvmCreator) RemoveCache(contractName string) {
	x.cm.RemoveCode(contractName)
}

type xvmInstance struct {
	bridgeCtx *bridge.Context
	execCtx   *exec.Context
	desc      pb.WasmCodeDesc
}

func (x *xvmInstance) Exec(function string) error {
	mem := x.execCtx.Memory()
	if mem == nil {
		return errors.New("bad contract, no memory")
	}
	_, err := x.execCtx.Exec(function, []int64{})
	if err != nil {
		log.Error("exec contract error", "error", err, "contract", x.bridgeCtx.ContractName)
	}
	return err
}

func (x *xvmInstance) ResourceUsed() contract.Limits {
	limits := contract.Limits{
		Cpu: x.execCtx.GasUsed(),
	}
	mem := x.execCtx.Memory()
	if mem != nil {
		limits.Memory = int64(len(mem))
	}
	return limits
}

func (x *xvmInstance) Release() {
	x.execCtx.Release()
}

func (x *xvmInstance) InitDebugWriter(logger *log.Logger) {
	if logger == nil {
		return
	}
	instanceLogger := logger.New("contract", x.bridgeCtx.ContractName, "ctxid", x.bridgeCtx.ID)
	instanceLogWriter := newDebugWriter(instanceLogger)
	debug.SetWriter(x.execCtx, instanceLogWriter)
}

func init() {
	vm.Register("xvm", newXVMCreator)
}
