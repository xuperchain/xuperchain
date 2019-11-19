package emscripten

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	"github.com/xuperchain/xuperunion/xvm/debug"
	"github.com/xuperchain/xuperunion/xvm/exec"
)

const (
	mutableGlobalsKey = "mutableGlobals"
	stackAllocFunc    = "stackAlloc"

	// mutableGlobalsBase is the base pointer of mutableGlobals
	// static data begin at 1024, the first 1024 bytes is not used.
	mutableGlobalsBase = 1024 - 100

	// total stack size, must sync with xc
	stackSize = 256 << 10 // 256KB
)

func unimplemented(symbol string) {
	exec.Throw(exec.NewTrap(fmt.Sprintf("%s not implemented", symbol)))
}

type mutableGlobals struct {
	HeapBase uint32
}

func getMutableGlobals(ctx exec.Context) *mutableGlobals {
	return ctx.GetUserData(mutableGlobalsKey).(*mutableGlobals)
}

func memoryStackBase(ctx exec.Context) (uint32, error) {
	base, err := ctx.Exec(stackAllocFunc, []int64{0})
	if err != nil {
		if _, ok := err.(*exec.ErrFuncNotFound); !ok {
			return 0, err
		}
		// stackAllocFunc not found, fallback to StaticTop method.
		base := ctx.StaticTop()
		// align 4K boundry
		if base%4096 != 0 {
			base += 4096 - base%4096
		}
		return base, nil
	}
	ctx.ResetGasUsed()
	return uint32(base), nil
}

// Init initialize global variables
func Init(ctx exec.Context) error {
	mem := ctx.Memory()
	if mem == nil {
		return errors.New("no memory")
	}
	if mutableGlobalsBase >= len(mem) {
		return errors.New("bad memory size")
	}

	stackBase, err := memoryStackBase(ctx)
	if err != nil {
		return err
	}
	mg := (*mutableGlobals)(unsafe.Pointer(&mem[mutableGlobalsBase]))
	mg.HeapBase = stackBase + stackSize
	ctx.SetUserData(mutableGlobalsKey, mg)
	return nil
}

// NewResolver return exec.Resolver which resolves symbols needed by emscripten environment
func NewResolver() exec.Resolver {
	return resolver
}

var resolver = exec.MapResolver(map[string]interface{}{
	"env.___setErrNo": func(ctx exec.Context, addr uint32) uint32 {
		return 0
	},
	"env.abortOnCannotGrowMemory": func(ctx exec.Context, code uint32) uint32 {
		exec.Throw(exec.NewTrap("cannot grow memory"))
		return 0
	},
	"env.abortStackOverflow": func(ctx exec.Context, code uint32) uint32 {
		exec.Throw(exec.NewTrap("stack overflow"))
		return 0
	},
	"env.getTotalMemory": func(ctx exec.Context) uint32 {
		mem := ctx.Memory()
		if mem != nil {
			return uint32(len(mem))
		}
		return 0
	},
	"env.enlargeMemory": func(ctx exec.Context) uint32 {
		return 0
	},
	"env._emscripten_memcpy_big": func(ctx exec.Context, dest, src, len uint32) uint32 {
		codec := exec.NewCodec(ctx)
		destbuf := codec.Bytes(dest, len)
		srcbuf := codec.Bytes(src, len)
		copy(destbuf, srcbuf)
		return len
	},
	"env._emscripten_get_heap_size": func(ctx exec.Context) uint32 {
		mem := ctx.Memory()
		if mem != nil {
			return uint32(len(mem))
		}
		return 0
	},
	"env._emscripten_resize_heap": func(ctx exec.Context, size uint32) uint32 {
		unimplemented("emscripten_resize_heap")
		return 0
	},
	"env.abort": func(ctx exec.Context, code uint32) uint32 {
		exec.Throw(exec.NewTrap("abort"))
		return 0
	},
	"env._abort": func(ctx exec.Context) uint32 {
		exec.Throw(exec.NewTrap("abort"))
		return 0
	},
	"env.___cxa_allocate_exception": func(ctx exec.Context, x uint32) uint32 {
		exec.Throw(exec.NewTrap("allocate exception"))
		return 0
	},
	"env.___cxa_throw": func(ctx exec.Context, x, y, z uint32) uint32 {
		exec.Throw(exec.NewTrap("throw"))
		return 0
	},
	"env.___cxa_pure_virtual": func(ctx exec.Context) uint32 {
		unimplemented("___cxa_pure_virtual")
		return 0
	},
	"env.___syscall140": func(ctx exec.Context, x, y uint32) uint32 {
		unimplemented("syscall140")
		return 0
	},
	"env.___syscall146": func(ctx exec.Context, no, argsPtr uint32) uint32 {
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
	"env.___syscall54": func(ctx exec.Context, x, y uint32) uint32 {
		return 0
	},
	"env.___syscall6": func(ctx exec.Context, x, y uint32) uint32 {
		unimplemented("syscall6")
		return 0
	},
	"env.___lock": func(ctx exec.Context, x uint32) uint32 {
		return 0
	},
	"env.___unlock": func(ctx exec.Context, x uint32) uint32 {
		return 0
	},
	"env._pthread_equal": func(ctx exec.Context, x, y uint32) uint32 {
		return 0
	},
	"env._llvm_trap": func(ctx exec.Context) uint32 {
		exec.Throw(exec.NewTrap("llvm trap called"))
		return 0
	},
	"env.___assert_fail": func(ctx exec.Context, x, y, w, z uint32) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	// TODO: zq @icex need to implement soon, from _llvm_stackrestore to ___cxa_uncaught_exception
	"env._llvm_stackrestore": func(ctx exec.Context, x uint32) uint32 {
		return 0
	},
	"env._llvm_stacksave": func(ctx exec.Context) uint32 {
		return 0
	},

	"env._getenv": func(ctx exec.Context, x uint32) uint32 {
		return 0
	},

	"env._strftime_l": func(ctx exec.Context, x, y, w, z, f uint32) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	"env._pthread_cond_wait": func(ctx exec.Context, x, y uint32) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	"env.___syscall91": func(ctx exec.Context, x, y uint32) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	"env.___syscall145": func(ctx exec.Context, x, y uint32) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	"env.___map_file": func(ctx exec.Context, x, y uint32) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	"env.___cxa_uncaught_exception": func(ctx exec.Context) uint32 {
		exec.Throw(exec.NewTrap("assert_fail"))
		return 0
	},

	"env.__table_base":   int64(0),
	"env.tableBase":      int64(0),
	"env.DYNAMICTOP_PTR": int64(mutableGlobalsBase + uint32(unsafe.Offsetof(new(mutableGlobals).HeapBase))),
	"global.NaN":         int64(math.Float64bits(math.NaN())),
	"global.Infinity":    int64(math.Float64bits(math.Inf(0))),
})

func errno(n int32) uint32 {
	return *(*uint32)(unsafe.Pointer(&n))
}
