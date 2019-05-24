package exec

// #include "xvm.h"
// #include "stdlib.h"
// extern xvm_resolver_t make_resolver_t(void* env);
// #cgo LDFLAGS: -ldl
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/xuperchain/xuperunion/xvm/pointer"
)

// Code represents the wasm code object
type Code struct {
	code   *C.xvm_code_t
	bridge *resolverBridge
	// 因为cgo不能持有go的pointer，这个指针是一个指向bridge的token，最后需要Delete
	bridgePointer uintptr
}

// NewCode instances a Code object from file path of native shared library
func NewCode(module string, resolver Resolver) (code *Code, err error) {
	bridge := newResolverBridge(resolver)
	bridgePointer := pointer.Save(bridge)
	defer CaptureTrap(&err)

	cpath := C.CString(module)
	defer C.free(unsafe.Pointer(cpath))
	resolvert := C.make_resolver_t(unsafe.Pointer(bridgePointer))
	ccode := C.xvm_new_code(cpath, resolvert)

	if ccode == nil {
		return nil, fmt.Errorf("open module %s error", module)
	}
	return &Code{
		code:          ccode,
		bridge:        bridge,
		bridgePointer: bridgePointer,
	}, nil
}

// Release releases resources hold by Code
func (c *Code) Release() {
	C.xvm_release_code(c.code)
	pointer.Delete(c.bridgePointer)
}
