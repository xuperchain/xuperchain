package main

import (
	"fmt"

	"github.com/xuperchain/xuperchain/core/xvm/exec"
)

var resolver = exec.MapResolver(map[string]interface{}{
	"go.main.puts": func(ctx exec.Context, sp uint32) uint32 {
		codec := exec.NewCodec(ctx)
		str := codec.GoString(sp + 8)
		fmt.Print(str)
		return 0
	},
	"env._print": func(ctx exec.Context, ptr uint32) uint32 {
		codec := exec.NewCodec(ctx)
		str := codec.CString(ptr)
		fmt.Print(str)
		return 0
	},
	"env._println": func(ctx exec.Context, ptr uint32) uint32 {
		codec := exec.NewCodec(ctx)
		str := codec.CString(ptr)
		fmt.Println(str)
		return 0
	},
})
