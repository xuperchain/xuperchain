package exec

import (
	"sync"
	"testing"
)

func TestNewContext(t *testing.T) {
	withCode(t, "testdata/add.wat", nil, func(code *Code) {
		ctx, err := NewContext(code, DefaultContextConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Release()
		ret, err := ctx.Exec("_add", []int64{1, 2})
		if err != nil {
			t.Fatal(err)
		}
		if ret != 3 {
			t.Errorf("expect 3 got %d", ret)
		}
	})
}

func TestResolveFunc(t *testing.T) {
	r := MapResolver(map[string]interface{}{
		"env._print": func(ctx *Context, addr uint32) uint32 {
			c := NewCodec(ctx)
			if c.CString(addr) != "hello world" {
				panic("not equal")
			}
			return 0
		},
	})
	withCode(t, "testdata/extern_func.wat", r, func(code *Code) {
		ctx, err := NewContext(code, DefaultContextConfig())
		if err != nil {
			t.Fatal(err)
		}
		_, err = ctx.Exec("_run", nil)
		if err != nil {
			t.Fatal(err)
		}
		ctx.Release()
	})
}

func TestGasUsed(t *testing.T) {
	r := MapResolver(map[string]interface{}{
		"env._print": func(ctx *Context, addr uint32) uint32 {
			return 0
		},
		"env.___setErrNo": func(ctx *Context, addr uint32) uint32 {
			return 0
		},
		"env.abortOnCannotGrowMemory": func(ctx *Context, code uint32) uint32 {
			return 0
		},
		"env.getTotalMemory": func(ctx *Context) uint32 {
			return 0
		},
		"env.enlargeMemory": func(ctx *Context) uint32 {
			return 0
		},
		"env.STACKTOP":       float64(4 << 10),
		"env.DYNAMICTOP_PTR": float64(4<<10 + 4),
	})
	withCode(t, "testdata/malloc.wat", r, func(code *Code) {
		ctx, err := NewContext(code, DefaultContextConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer ctx.Release()
		for i := 0; i < 10; i++ {
			_, err = ctx.Exec("_run", nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("gas %d", ctx.GasUsed())
		}
	})
}

func BenchmarkExecParallel(b *testing.B) {
	r := MapResolver(map[string]interface{}{
		"env._print": func(ctx *Context, addr uint32) uint32 {
			return 0
		},
	})
	withCode(b, "testdata/extern_func.wat", r, func(code *Code) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx, err := NewContext(code, DefaultContextConfig())
				if err != nil {
					b.Fatal(err)
				}
				_, err = ctx.Exec("_run", nil)
				if err != nil {
					b.Fatal(err)
				}
				ctx.Release()
			}
		})
	})
}

func BenchmarkExecSerial(b *testing.B) {
	r := MapResolver(map[string]interface{}{
		"env._print": func(ctx *Context, addr uint32) uint32 {
			return 0
		},
	})
	withCode(b, "testdata/extern_func.wat", r, func(code *Code) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx, err := NewContext(code, DefaultContextConfig())
			if err != nil {
				b.Fatal(err)
			}
			_, err = ctx.Exec("_run", nil)
			if err != nil {
				b.Fatal(err)
			}
			ctx.Release()
		}
	})
}

func BenchmarkExecWorker(b *testing.B) {
	r := MapResolver(map[string]interface{}{
		"env._print": func(ctx *Context, addr uint32) uint32 {
			return 0
		},
	})

	withCode(b, "testdata/extern_func.wat", r, func(code *Code) {
		wg := new(sync.WaitGroup)
		worker := func(ch chan int) {
			for range ch {
				ctx, err := NewContext(code, DefaultContextConfig())
				if err != nil {
					b.Fatal(err)
				}
				ctx.Exec("_run", nil)
				ctx.Release()
			}
			wg.Done()
		}
		b.ResetTimer()
		ch := make(chan int)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go worker(ch)
		}
		for i := 0; i < b.N; i++ {
			ch <- i
		}
		close(ch)
		wg.Wait()
	})
}
