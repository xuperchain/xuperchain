package gowasm

import (
	"time"

	"github.com/xuperchain/xuperunion/xvm/debug"
	"github.com/xuperchain/xuperunion/xvm/exec"
	"github.com/xuperchain/xuperunion/xvm/runtime/go/js"
	"github.com/xuperchain/xuperunion/xvm/runtime/go/js/fs"
)

const (
	goRuntimeKey = "goruntime"
)

// Runtime implements the runtime needed to run wasm code compiled by go toolchain
type Runtime struct {
	exitcode int32
	exited   bool
	global   *js.Global
	jsvm     *js.VM
	ctx      *exec.Context

	timeOrigin time.Time
}

// RegisterRuntime 用于向exec.Context里面注册一个初始化好的js Runtime
func RegisterRuntime(ctx *exec.Context) *Runtime {
	rt := &Runtime{
		global:     js.NewGlobal(),
		timeOrigin: time.Now(),
		ctx:        ctx,
	}
	ctx.SetUserData(goRuntimeKey, rt)

	jsmem := js.NewMemory(func() []byte {
		return rt.ctx.Memory()
	})
	rt.jsvm = js.NewVM(&js.VMConfig{
		Memory: jsmem,
		Global: rt.global,
	})
	rt.global.Register("Fs", fs.NewFS())
	return rt
}

func (rt *Runtime) wasmExit(code int32) {
	rt.exitcode = code
	rt.exited = true
}

func (rt *Runtime) wasmWrite(fd int64, p int64, n int32) {
	codec := exec.NewCodec(rt.ctx)
	if fd == 1 || fd == 2 {
		debug.Write(rt.ctx, codec.Bytes(uint32(p), uint32(n)))
	}
}

func (rt *Runtime) nanotime() int64 {
	return 1
}

func (rt *Runtime) walltime() (int64, int32) {
	return 0, 0
}

// Exited will be true if runtime.wasmExit has been called
func (rt *Runtime) Exited() bool {
	return rt.exited
}

// ExitCode return the exit code of go application
func (rt *Runtime) ExitCode() int32 {
	return rt.exitcode
}

// WaitTimer waiting for timeout of timers set by go runtime in wasm
func (rt *Runtime) WaitTimer() {
}

func (rt *Runtime) scheduleCallback(delay int64) int32 {
	return 0
}

func (rt *Runtime) clearScheduledCallback(id int32) {
}

func (rt *Runtime) getRandomData(r []byte) {
	for i := range r {
		r[i] = 0
	}
}

func (rt *Runtime) debug(v int64) {
}

func (rt *Runtime) syscallJsValueGet(ref js.Ref, name string) (ret js.Ref) {
	defer rt.jsvm.CatchException(&ret)
	ret = rt.jsvm.Property(ref, name)
	return ret
}

func (rt *Runtime) syscallJsValueSet(ref js.Ref, name string, value js.Ref) {

}

func (rt *Runtime) syscallJsValueNew(ref js.Ref, args []js.Ref) (ret js.Ref, ok bool) {
	defer rt.jsvm.CatchException(&ret)
	return rt.jsvm.New(ref, args), true
}

func (rt *Runtime) syscallJsValueCall(ref js.Ref, method string, args []js.Ref) (ret js.Ref, ok bool) {
	defer rt.jsvm.CatchException(&ret)
	return rt.jsvm.Call(ref, method, args), true
}

func (rt *Runtime) syscallJsValueInvoke(ref js.Ref, args []js.Ref) (ret js.Ref, ok bool) {
	defer rt.jsvm.CatchException(&ret)
	return rt.jsvm.Invoke(ref, args), true
}

func (rt *Runtime) syscallJsValuePrepareString(ref js.Ref) (js.Ref, int64) {
	v := rt.jsvm.Value(ref)
	if v == nil {
		return rt.jsvm.Exception(js.ExceptionRefNotFound(ref)), 0
	}
	str := v.String()
	return rt.jsvm.Store(str), int64(len(str))
}

func (rt *Runtime) syscallJsValueLoadString(ref js.Ref, b []byte) {
	v := rt.jsvm.Value(ref)
	if v == nil {
		return
	}
	str := v.String()
	copy(b, str)
}

func (rt *Runtime) syscallJsStringVal(value string) js.Ref {
	return rt.jsvm.Store(value)
}

func (rt *Runtime) syscallJsValueIndex(ref js.Ref, idx int64) js.Ref {
	return 0
}

func (rt *Runtime) syscallJsValueSetIndex(ref js.Ref, idx int64, x js.Ref) {
}

func (rt *Runtime) syscallJsValueLength(ref js.Ref) int64 {
	return 0
}

func (rt *Runtime) syscallJsValueInstanceOf(v, t js.Ref) {
}
