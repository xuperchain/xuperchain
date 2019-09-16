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
	Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error)
}

// XuperKernel define kernel contract method type
type XuperKernel struct {
	methods map[string]Method
}

// KContext define kernel contract context type
type KContext struct {
	xk            *XuperKernel
	resourceUsed  contract.Limits
	ResourceLimit contract.Limits
	ModelCache    *xmodel.XMCache
	Initiator     string
	AuthRequire   []string
	ContextConfig *contract.ContextConfig
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
		ModelCache:    ctxCfg.XMCache,
		xk:            xk,
		Initiator:     ctxCfg.Initiator,
		AuthRequire:   ctxCfg.AuthRequire,
		ResourceLimit: ctxCfg.ResourceLimits,
		ContextConfig: ctxCfg,
	}, nil
}

// Invoke entrance for kernel contract method invoke
func (kc *KContext) Invoke(methodName string, args map[string][]byte) (*contract.Response, error) {
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
func (kc *KContext) AddXFeeUsed(delta int64) {
	if delta < 0 {
		panic(fmt.Sprintf("bad xfee delta %d", delta))
	}
	kc.resourceUsed.XFee += delta
}

func (kc *KContext) AddResourceUsed(delta contract.Limits) {
	kc.resourceUsed.Cpu += delta.Cpu
	kc.resourceUsed.Memory += delta.Memory
	kc.resourceUsed.Disk += delta.Disk
	kc.resourceUsed.XFee += delta.XFee
}

func (kc *KContext) ResourceUsed() contract.Limits {
	return kc.resourceUsed
}

// Release release context
func (kc *KContext) Release() error {
	return nil
}
