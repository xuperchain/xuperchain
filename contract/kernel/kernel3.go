package kernel

import (
	"fmt"

	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/wasm"
	"github.com/xuperchain/xuperunion/xmodel"
)

// ModuleName moudle name
const ModuleName = "xkernel"

// Method define method interface needed
type Method interface {
	Invoke(ctx *KContext, args map[string][]byte) ([]byte, error)
}

// XuperKernel define kernel contract method type
type XuperKernel struct {
	methods map[string]Method
}

// KContext define kernel contract context type
type KContext struct {
	xk          *XuperKernel
	gasUsed     int64
	ModelCache  *xmodel.XMCache
	GasLimit    int64
	Initiator   string
	AuthRequire []string
}

// NewKernel new an instance of XuperKernel, initialized with kernel contract method
func NewKernel(vmm *wasm.VMManager) (*XuperKernel, error) {
	return &XuperKernel{
		methods: map[string]Method{
			"Get":           &GetMethod{},
			"NewAccount":    &NewAccountMethod{},
			"SetAccountAcl": &SetAccountACLMethod{},
			"SetMethodAcl":  &SetMethodACLMethod{},
			"Deploy":        &DeployMethod{vmm: vmm},
		},
	}, nil
}

// GetName get moudle name
func (xk *XuperKernel) GetName() string {
	return ModuleName
}

// NewContext new a context, initialized with KernelContext
func (xk *XuperKernel) NewContext(ctxCfg *contract.ContextConfig) (contract.Context, error) {
	return &KContext{
		ModelCache:  ctxCfg.XMCache,
		xk:          xk,
		GasLimit:    ctxCfg.GasLimit,
		Initiator:   ctxCfg.Initiator,
		AuthRequire: ctxCfg.AuthRequire,
	}, nil
}

// Invoke entrance for kernel contract method invoke
func (kc *KContext) Invoke(methodName string, args map[string][]byte) ([]byte, error) {
	method := kc.xk.methods[methodName]
	if method == nil {
		return nil, fmt.Errorf("Mehotd %s not exists in %s", methodName, ModuleName)
	}
	res, err := method.Invoke(kc, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// AddGasUsed set gas used when invoking kernel contract method
func (kc *KContext) AddGasUsed(delta int64) {
	if delta < 0 {
		panic(fmt.Sprintf("bad gas delta %d", delta))
	}
	kc.gasUsed += delta
}

// GasUsed return gas used for calling specific kernel contract method
func (kc *KContext) GasUsed() int64 {
	return kc.gasUsed
}

// Release release context
func (kc *KContext) Release() error {
	return nil
}
