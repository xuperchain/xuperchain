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
	code = new(Code)
	code.bridge = newResolverBridge(resolver)
	code.bridgePointer = pointer.Save(code.bridge)
	// xvm_new_code执行期间可能会抛出Trap，导致资源泄露
	// 如果CaptureTrap捕获了Trap则释放所有已经初始化的资源
	defer func() {
		if err != nil {
			code.Release()
			code = nil
		}
	}()
	defer CaptureTrap(&err)

	cpath := C.CString(module)
	defer C.free(unsafe.Pointer(cpath))
	resolvert := C.make_resolver_t(unsafe.Pointer(code.bridgePointer))
	code.code = C.xvm_new_code(cpath, resolvert)

	if code.code == nil {
		err = fmt.Errorf("open module %s error", module)
		return
	}
	return
}

// Release releases resources hold by Code
func (c *Code) Release() {
	if c.code != nil {
		C.xvm_release_code(c.code)
	}
	if c.bridgePointer != 0 {
		pointer.Delete(c.bridgePointer)
	}
	*c = Code{}
}
