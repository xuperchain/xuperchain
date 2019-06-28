package bridge

import (
	"github.com/xuperchain/xuperunion/contract"
)

// Executor 为用户态虚拟机工厂类
type Executor interface {
	// RegisterSyscallService 用于虚拟机把系统调用链接到合约代码上，类似vdso
	// 注册到Registry的时候被调用一次
	RegisterSyscallService(*SyscallService)
	// NewInstance 根据合约Context返回合约虚拟机的一个实例
	NewInstance(ctx *Context) (Instance, error)
}

// Instance is an instance of a contract run
type Instance interface {
	// Exec根据ctx里面的参数执行合约代码
	Exec() error
	// ResourceUsed returns the resource used by contract
	ResourceUsed() contract.Limits
	// Release releases contract instance
	Release()
}

// XBridge 用于注册用户虚拟机以及向Xchain Core注册可被识别的vm.VirtualMachine
type XBridge struct {
	ctxmgr         *ContextManager
	syscallService *SyscallService
	vms            map[string]contract.VirtualMachine
}

// New instances a new XBridge
func New() *XBridge {
	ctxmgr := NewContextManager()
	syscallService := NewSyscallService(ctxmgr)
	return &XBridge{
		ctxmgr:         ctxmgr,
		syscallService: syscallService,
		vms:            make(map[string]contract.VirtualMachine),
	}
}

func (v *XBridge) convertToVM(name string, exec Executor) contract.VirtualMachine {
	wraper := &vmImpl{
		ctxmgr: v.ctxmgr,
		name:   name,
		exec:   exec,
	}
	return wraper
}

// RegisterExecutor register a Executor to XBridge
func (v *XBridge) RegisterExecutor(name string, exec Executor) contract.VirtualMachine {
	wraper := v.convertToVM(name, exec)
	exec.RegisterSyscallService(v.syscallService)
	v.vms[name] = wraper
	return wraper
}

// GetVirtualMachine returns a contract.VirtualMachine from the given name
func (v *XBridge) GetVirtualMachine(name string) (contract.VirtualMachine, bool) {
	vm, ok := v.vms[name]
	return vm, ok
}

// RegisterToXCore register VirtualMachines to xchain core
func (v *XBridge) RegisterToXCore(regfunc func(name string, vm contract.VirtualMachine) error) {
	for _, vm := range v.vms {
		err := regfunc(vm.GetName(), vm)
		if err != nil {
			panic(err)
		}
	}
}
