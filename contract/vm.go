package contract

import (
	"errors"
	"sync"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/xmodel"
)

const (
	// MaxGasLimit FIXME: 设置一个更合理的
	MaxGasLimit = 0xFFFFFFFF
)

var (
	// ErrVMNotExist is returned when found vm not exist
	ErrVMNotExist = errors.New("Vm not exist in vm manager")
)

// ContextConfig define the config of context
type ContextConfig struct {
	XMCache      *xmodel.XMCache
	Initiator    string
	AuthRequire  []string
	ContractName string
	GasLimit     int64
}

// VirtualMachine define virtual machine interface
type VirtualMachine interface {
	GetName() string
	NewContext(*ContextConfig) (Context, error)
}

// Context define context interface
type Context interface {
	Invoke(method string, args map[string][]byte) ([]byte, error)
	GasUsed() int64
	Release() error
}

// VMManager define VMManager type
type VMManager struct {
	lock   *sync.Mutex
	vms    map[string]VirtualMachine
	logger log.Logger
}

// NewVMManager new an instance of VMManager
func NewVMManager(logger log.Logger) (*VMManager, error) {
	vmMgr := &VMManager{
		lock:   new(sync.Mutex),
		vms:    map[string]VirtualMachine{},
		logger: logger,
	}
	return vmMgr, nil
}

// RegisterVM register an instance of VM into VMManager
func (vmMgr *VMManager) RegisterVM(module string, vm VirtualMachine) error {
	vmMgr.lock.Lock()
	defer vmMgr.lock.Unlock()
	vmMgr.vms[module] = vm
	return nil
}

// GetVM return specific virtual machine instance
func (vmMgr *VMManager) GetVM(module string) (VirtualMachine, error) {
	if vmMgr.vms[module] == nil {
		return nil, ErrVMNotExist
	}
	return vmMgr.vms[module], nil
}
