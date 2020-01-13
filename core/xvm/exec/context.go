package exec

// #include "xvm.h"
// #include "stdlib.h"
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

const (
	// MaxGasLimit is the maximum gas limit
	MaxGasLimit = 0xFFFFFFFF
)

type ErrFuncNotFound struct {
	Name string
}

func (e *ErrFuncNotFound) Error() string {
	return fmt.Sprintf("%s not found", e.Name)
}

// ContextConfig configures an execution context
type ContextConfig struct {
	GasLimit int64
}

// DefaultContextConfig returns the default configuration of ContextConfig
func DefaultContextConfig() *ContextConfig {
	return &ContextConfig{
		GasLimit: MaxGasLimit,
	}
}

// Context hold the context data when running a wasm instance
type Context interface {
	Exec(name string, param []int64) (ret int64, err error)
	GasUsed() int64
	ResetGasUsed()
	Memory() []byte
	StaticTop() uint32
	SetUserData(key string, value interface{})
	GetUserData(key string) interface{}
	Release()
}

type Code interface {
	NewContext(cfg *ContextConfig) (ictx Context, err error)
	Release()
}

type aotContext struct {
	context  C.xvm_context_t
	gasUsed  int64
	cfg      ContextConfig
	userData map[string]interface{}
}

// NewContext instances a Context from Code
func (code *aotCode) NewContext(cfg *ContextConfig) (ictx Context, err error) {
	ctx := &aotContext{
		cfg:      *cfg,
		userData: make(map[string]interface{}),
	}
	defer func() {
		if err != nil {
			ctx.Release()
			ctx = nil
		}
	}()
	defer CaptureTrap(&err)
	ret := C.xvm_init_context(&ctx.context, code.code, C.uint64_t(cfg.GasLimit))
	if ret == 0 {
		return nil, errors.New("init context error")
	}
	ictx = ctx
	return ictx, nil
}

// Release releases resources hold by Context
func (c *aotContext) Release() {
	C.xvm_release_context(&c.context)
}

func isalpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isalnum(c byte) bool {
	return isalpha(c) || (c >= '0' && c <= '9')
}

// legalizeName makes a name a legail c identifier
func legalizeName(name string) string {
	if len(name) == 0 {
		return "_"
	}
	result := make([]byte, 1, len(name))
	result[0] = name[0]
	if !isalpha(name[0]) {
		result[0] = '_'
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !isalnum(c) {
			c = '_'
		}
		result = append(result, c)
	}
	return string(result)

}

// Exec executes a wasm function by given name and param
func (c *aotContext) Exec(name string, param []int64) (ret int64, err error) {
	defer CaptureTrap(&err)

	exportName := "export_" + legalizeName(name)
	cname := C.CString(exportName)
	defer C.free(unsafe.Pointer(cname))

	var args *C.int64_t
	if len(param) != 0 {
		args = (*C.int64_t)(unsafe.Pointer(&param[0]))
	}
	var cret C.int64_t
	ok := C.xvm_call(&c.context, cname, args, C.int64_t(len(param)), &cret)
	if ok == 0 {
		return 0, &ErrFuncNotFound{
			Name: name,
		}
	}
	ret = int64(cret)
	return
}

// GasUsed returns the gas used by Exec
func (c *aotContext) GasUsed() int64 {
	return int64(C.xvm_gas_used(&c.context))
}

// ResetGasUsed reset the gas counter
func (c *aotContext) ResetGasUsed() {
	C.xvm_reset_gas_used(&c.context)
}

// Memory returns the memory of current context, nil will be returned if wasm code has no memory
func (c *aotContext) Memory() []byte {
	if c.context.mem == nil || c.context.mem.size == 0 {
		return nil
	}
	ptr := c.context.mem.data
	n := int(c.context.mem.size)
	return (*[1 << 30]byte)(unsafe.Pointer(ptr))[:n:n]
}

// StaticTop returns the static data's top offset of memory
func (c *aotContext) StaticTop() uint32 {
	return uint32(C.xvm_mem_static_top(&c.context))
}

// SetUserData store key-value pair to Context which can be retrieved by GetUserData
func (c *aotContext) SetUserData(key string, value interface{}) {
	c.userData[key] = value
}

// GetUserData retrieves user data stored by SetUserData
func (c *aotContext) GetUserData(key string) interface{} {
	return c.userData[key]
}
