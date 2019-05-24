package exec

// #include "xvm.h"
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/xuperchain/xuperunion/xvm/pointer"
)

// A Resolver resolves global and function symbols imported by wasm code
type Resolver interface {
	ResolveFunc(module, name string) (interface{}, bool)
	ResolveGlobal(module, name string) (float64, bool)
}

// MultiResolver chains multiple Resolvers, symbol looking up is according to the order of resolvers.
// The first found symbol will be returned.
type MultiResolver []Resolver

// NewMultiResolver instance a MultiResolver from resolves
func NewMultiResolver(resolvers ...Resolver) MultiResolver {
	return resolvers
}

// ResolveFunc implements Resolver interface
func (m MultiResolver) ResolveFunc(module, name string) (interface{}, bool) {
	for _, r := range m {
		if f, ok := r.ResolveFunc(module, name); ok {
			return f, true
		}
	}
	return nil, false
}

// ResolveGlobal implements Resolver interface
func (m MultiResolver) ResolveGlobal(module, name string) (float64, bool) {
	for _, r := range m {
		if v, ok := r.ResolveGlobal(module, name); ok {
			return v, true
		}
	}
	return 0, false
}

type importFunc struct {
	module, name string
	body         interface{}
}

type resolverBridge struct {
	resolver Resolver
	funcmap  map[string]int
	funcs    []importFunc
}

func newResolverBridge(r Resolver) *resolverBridge {
	return &resolverBridge{
		resolver: r,
		funcmap:  make(map[string]int),
		funcs:    make([]importFunc, 1),
	}
}

//export xvm_resolve_func
func xvm_resolve_func(env unsafe.Pointer, module, name *C.char) C.wasm_rt_func_handle_t {
	r := pointer.Restore(uintptr(env)).(*resolverBridge)
	moduleStr, nameStr := C.GoString(module), C.GoString(name)
	key := moduleStr + ":" + nameStr

	idx := r.funcmap[key]
	if idx != 0 {
		return C.wasm_rt_func_handle_t(uintptr(idx))
	}
	if r.resolver == nil {
		Throw(&TrapSymbolNotFound{
			Module: moduleStr,
			Name:   nameStr,
		})
	}

	f, ok := r.resolver.ResolveFunc(moduleStr, nameStr)
	if !ok {
		Throw(&TrapSymbolNotFound{
			Module: moduleStr,
			Name:   nameStr,
		})
	}
	r.funcs = append(r.funcs, importFunc{
		module: moduleStr,
		name:   nameStr,
		body:   f,
	})
	idx = len(r.funcs) - 1
	r.funcmap[key] = idx
	return C.wasm_rt_func_handle_t(uintptr(idx))
}

//export xvm_resolve_global
func xvm_resolve_global(env unsafe.Pointer, module, name *C.char) C.double {
	r := pointer.Restore(uintptr(env)).(*resolverBridge)
	moduleStr, nameStr := C.GoString(module), C.GoString(name)
	value, ok := r.resolver.ResolveGlobal(moduleStr, nameStr)
	if !ok {
		Throw(&TrapSymbolNotFound{
			Module: moduleStr,
			Name:   nameStr,
		})
	}
	return C.double(value)
}

//export xvm_call_func
func xvm_call_func(env unsafe.Pointer, handle C.wasm_rt_func_handle_t,
	ctxptr *C.xvm_context_t, params *C.uint32_t, paramLen C.uint32_t) C.uint32_t {
	r := pointer.Restore(uintptr(env)).(*resolverBridge)
	idx := int(uintptr(handle))
	if idx <= 0 || idx >= len(r.funcs) {
		Throw(NewTrap(fmt.Sprintf("bad func idx %d", idx)))
	}
	f := r.funcs[idx]
	args := make([]uint32, paramLen)
	for i := range args {
		args[i] = *(*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(params)) + uintptr(i*4)))
	}
	// TODO: 因为context字段是Context的第一个字段，可以强转，希望后续go的内存布局不会变化
	// FIXME: cgo应该不会有问题，如果有问题可以使用pointer package来转换
	ctx := (*Context)(unsafe.Pointer(ctxptr))
	ret, ok := applyFuncCall(ctx, f.body, args)
	if !ok {
		Throw(&TrapFuncSignatureNotMatch{
			Module: f.module,
			Name:   f.name,
		})
	}
	return C.uint32_t(ret)
}

func applyFuncCall(ctx *Context, f interface{}, params []uint32) (uint32, bool) {
	len := len(params)
	switch fun := f.(type) {
	case func(*Context) uint32:
		if len != 0 {
			return 0, false
		}
		return fun(ctx), true
	case func(*Context, uint32) uint32:
		if len != 1 {
			return 0, false
		}
		return fun(ctx, params[0]), true
	case func(*Context, uint32, uint32) uint32:
		if len != 2 {
			return 0, false
		}
		return fun(ctx, params[0], params[1]), true
	case func(*Context, uint32, uint32, uint32) uint32:
		if len != 3 {
			return 0, false
		}
		return fun(ctx, params[0], params[1], params[2]), true
	case func(*Context, uint32, uint32, uint32, uint32) uint32:
		if len != 4 {
			return 0, false
		}
		return fun(ctx, params[0], params[1], params[2], params[3]), true
	default:
		return 0, false
	}
}
