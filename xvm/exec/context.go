package exec

// #include "xvm.h"
// #include "stdlib.h"
import "C"
import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

const (
	// MaxGasLimit is the maximum gas limit
	MaxGasLimit = 0xFFFFFFFF
)

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
type Context struct {
	context  C.xvm_context_t
	gasUsed  int64
	cfg      ContextConfig
	userData map[string]interface{}
}

// NewContext instances a Context from Code
func NewContext(code *Code, cfg *ContextConfig) (ctx *Context, err error) {
	ctx = &Context{
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
	ret := C.xvm_init_context(&ctx.context, code.code)
	if ret == 0 {
		return nil, errors.New("init context error")
	}
	return ctx, nil
}

// Release releases resources hold by Context
func (c *Context) Release() {
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
func (c *Context) Exec(name string, param []int64) (ret int64, err error) {
	defer CaptureTrap(&err)

	exportName := "export_" + legalizeName(name)
	cname := C.CString(exportName)
	defer C.free(unsafe.Pointer(cname))

	var args *C.int64_t
	if len(param) != 0 {
		args = (*C.int64_t)(unsafe.Pointer(&param[0]))
	}
	var cgas C.wasm_rt_gas_t
	cgas.limit = C.int64_t(c.cfg.GasLimit)
	var cret C.int64_t
	ok := C.xvm_call(&c.context, cname, args, C.int64_t(len(param)), &cgas, &cret)
	if ok == 0 {
		return 0, fmt.Errorf("%s not found", name)
	}
	ret = int64(cret)
	c.gasUsed = int64(cgas.used)
	return
}

// GasUsed returns the gas used by Exec
func (c *Context) GasUsed() int64 {
	return c.gasUsed
}

// Memory returns the memory of current context, nil will be returned if wasm code has no memory
func (c *Context) Memory() []byte {
	if c.context.mem == nil || c.context.mem.size == 0 {
		return nil
	}
	var mem []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&mem))
	header.Data = uintptr(unsafe.Pointer(c.context.mem.data))
	header.Len = int(c.context.mem.size)
	header.Cap = int(c.context.mem.size)
	return mem
}

// SetUserData store key-value pair to Context which can be retrieved by GetUserData
func (c *Context) SetUserData(key string, value interface{}) {
	c.userData[key] = value
}

// GetUserData retrieves user data stored by SetUserData
func (c *Context) GetUserData(key string) interface{} {
	return c.userData[key]
}
