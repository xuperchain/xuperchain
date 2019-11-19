package wasi

import "github.com/xuperchain/xuperunion/xvm/exec"

var resolver = exec.MapResolver(map[string]interface{}{
	"wasi_unstable.fd_prestat_get": func(ctx exec.Context, x, y uint32) uint32 {
		return 8
	},
	"wasi_unstable.fd_fdstat_get": func(ctx exec.Context, x, y uint32) uint32 {
		return 8
	},
	"wasi_unstable.fd_prestat_dir_name": func(ctx exec.Context, x, y, z uint32) uint32 {
		return 8
	},
	"wasi_unstable.fd_close": func(ctx exec.Context, x uint32) uint32 {
		return 8
	},
	"wasi_unstable.fd_seek": func(ctx exec.Context, x, y, z, w uint32) uint32 {
		return 8
	},
	"wasi_unstable.fd_write": func(ctx exec.Context, x, y, z, w uint32) uint32 {
		return 8
	},
	"wasi_unstable.environ_sizes_get": func(ctx exec.Context, x, y uint32) uint32 {
		return 0
	},
	"wasi_unstable.environ_get": func(ctx exec.Context, x, y uint32) uint32 {
		return 0
	},
	"wasi_unstable.args_sizes_get": func(ctx exec.Context, x, y uint32) uint32 {
		return 0
	},
	"wasi_unstable.args_get": func(ctx exec.Context, x, y uint32) uint32 {
		return 0
	},
	"wasi_unstable.proc_exit": func(ctx exec.Context, x uint32) uint32 {
		exec.Throw(exec.NewTrap("exit"))
		return 0
	},
})

// NewResolver return exec.Resolver which resolves symbols needed by wasi environment
func NewResolver() exec.Resolver {
	return resolver
}
