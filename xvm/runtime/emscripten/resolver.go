package emscripten

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/xuperchain/xuperunion/xvm/debug"
	"github.com/xuperchain/xuperunion/xvm/exec"
)

const (
	mutableGlobalsKey = "mutableGlobals"

	bytesPage = 65536

	mutableGlobalsBase = 63 * bytesPage
	stackTop           = 64 * bytesPage
	stackMax           = 128 * bytesPage
)

func unimplemented(symbol string) {
	exec.Throw(exec.NewTrap(fmt.Sprintf("%s not implemented", symbol)))
}

type mutableGlobals struct {
	DynamicTop uint32
}

func getMutableGlobals(ctx *exec.Context) *mutableGlobals {
	return ctx.GetUserData(mutableGlobalsKey).(*mutableGlobals)
}

// Init initialize global variables
func Init(ctx *exec.Context) {
	mem := ctx.Memory()
	if mem == nil {
		return
	}
	mg := (*mutableGlobals)(unsafe.Pointer(&mem[mutableGlobalsBase]))
	mg.DynamicTop = stackMax
	ctx.SetUserData(mutableGlobalsKey, mg)
}

// NewResolver return exec.Resolver which resolves symbols needed by emscripten environment
func NewResolver() exec.Resolver {
	return resolver
}

var resolver = exec.MapResolver(map[string]interface{}{
	"env.___setErrNo": func(ctx *exec.Context, addr uint32) uint32 {
		return 0
	},
	"env.abortOnCannotGrowMemory": func(ctx *exec.Context, code uint32) uint32 {
		exec.Throw(exec.NewTrap("cannot grow memory"))
		return 0
	},
	"env.getTotalMemory": func(ctx *exec.Context) uint32 {
		mem := ctx.Memory()
		if mem != nil {
			return uint32(len(mem))
		}
		return 0
	},
	"env.enlargeMemory": func(ctx *exec.Context) uint32 {
		return 0
	},
	"env._emscripten_memcpy_big": func(ctx *exec.Context, dest, src, len uint32) uint32 {
		codec := exec.NewCodec(ctx)
		destbuf := codec.Bytes(dest, len)
		srcbuf := codec.Bytes(src, len)
		copy(destbuf, srcbuf)
		return len
	},
	"env._emscripten_get_heap_size": func(ctx *exec.Context) uint32 {
		mem := ctx.Memory()
		if mem != nil {
			return uint32(len(mem))
		}
		return 0
	},
	"env._emscripten_resize_heap": func(ctx *exec.Context, size uint32) uint32 {
		unimplemented("emscripten_resize_heap")
		return 0
	},
	"env.abort": func(ctx *exec.Context, code uint32) uint32 {
		exec.Throw(exec.NewTrap("abort"))
		return 0
	},
	"env._abort": func(ctx *exec.Context, code uint32) uint32 {
		exec.Throw(exec.NewTrap("abort"))
		return 0
	},
	"env.___cxa_allocate_exception": func(ctx *exec.Context, x uint32) uint32 {
		exec.Throw(exec.NewTrap("allocate exception"))
		return 0
	},
	"env.___cxa_throw": func(ctx *exec.Context, x, y, z uint32) uint32 {
		exec.Throw(exec.NewTrap("throw"))
		return 0
	},
	"env.___cxa_pure_virtual": func(ctx *exec.Context) uint32 {
		unimplemented("___cxa_pure_virtual")
		return 0
	},
	"env.___syscall140": func(ctx *exec.Context, x, y uint32) uint32 {
		unimplemented("syscall140")
		return 0
	},
	"env.___syscall146": func(ctx *exec.Context, no, argsPtr uint32) uint32 {
		codec := exec.NewCodec(ctx)
		fd := codec.Uint32(argsPtr)
		if fd != 1 && fd != 2 {
			return errno(-9)
		}
		iov := codec.Uint32(argsPtr + 4)
		iovcnt := codec.Uint32(argsPtr + 8)
		total := uint32(0)
		for i := uint32(0); i < iovcnt; i++ {
			base := codec.Uint32(iov + i*8)
			length := codec.Uint32(iov + i*8 + 4)
			buf := codec.Bytes(base, length)
			total += length
			debug.Write(ctx, buf)
		}
		return total
	},
	"env.___syscall54": func(ctx *exec.Context, x, y uint32) uint32 {
		return 0
	},
	"env.___syscall6": func(ctx *exec.Context, x, y uint32) uint32 {
		unimplemented("syscall6")
		return 0
	},
	"env.___lock": func(ctx *exec.Context, x uint32) uint32 {
		return 0
	},
	"env.___unlock": func(ctx *exec.Context, x uint32) uint32 {
		return 0
	},
	"env._pthread_equal": func(ctx *exec.Context, x, y uint32) uint32 {
		return 0
	},
	"env._llvm_trap": func(ctx *exec.Context) uint32 {
		exec.Throw(exec.NewTrap("llvm trap called"))
		return 0
	},
	"env.___assert_fail": func(ctx *exec.Context, x, y, w, z uint32) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	"env.__table_base":   float64(0),
	"env.tableBase":      float64(0),
	"env.STACKTOP":       float64(stackTop),
	"env.DYNAMICTOP_PTR": float64(mutableGlobalsBase + uint32(unsafe.Offsetof(new(mutableGlobals).DynamicTop))),
	"global.NaN":         math.NaN(),
	"global.Infinity":    math.Inf(0),
})

func errno(n int32) uint32 {
	return *(*uint32)(unsafe.Pointer(&n))
}
