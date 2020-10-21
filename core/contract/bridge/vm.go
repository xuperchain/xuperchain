package bridge

import (
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/pb"
)

type VMConfig interface {
	DriverName() string
	IsEnable() bool
}

// InstanceCreatorConfig configures InstanceCreator
type InstanceCreatorConfig struct {
	Basedir        string
	SyscallService *SyscallService
	// VMConfig is the config of vm driver
	VMConfig VMConfig
}

// NewInstanceCreatorFunc instances a new InstanceCreator from InstanceCreatorConfig
type NewInstanceCreatorFunc func(config *InstanceCreatorConfig) (InstanceCreator, error)

// ContractCodeProvider provides source code and desc of contract
type ContractCodeProvider interface {
	GetContractCodeDesc(name string) (*pb.WasmCodeDesc, error)
	GetContractCode(name string) ([]byte, error)
	GetContractAbi(name string) ([]byte, error)
}

// InstanceCreator is the creator of contract virtual machine instance
type InstanceCreator interface {
	// CreateInstance instances a wasm virtual machine instance which can run a single contract call
	CreateInstance(ctx *Context, cp ContractCodeProvider) (Instance, error)
	RemoveCache(name string)
}

// Instance is a contract virtual machine instance which can run a single contract call
type Instance interface {
	Exec() error
	ResourceUsed() contract.Limits
	Release()
	Abort(msg string)
}
