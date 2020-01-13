# golang-asm

A mirror of the assembler from the Go compiler, with import paths re-written for the assembler to be functional as a standalone library.

License as per the Go project.

# Status

Works, but expect to dig into the assembler godoc's to work out what to set different parameters of `obj.Prog` to get it to generate specific instructions.

# Example

Demonstrates assembly of a NOP & an ADD instruction on x86-64.

```go

package main

import (
	"fmt"

	asm "github.com/twitchyliquid64/golang-asm"
	"github.com/twitchyliquid64/golang-asm/obj"
	"github.com/twitchyliquid64/golang-asm/obj/x86"
)

func noop(builder *asm.Builder) *obj.Prog {
	prog := builder.NewProg()
	prog.As = x86.ANOPL
	prog.From.Type = obj.TYPE_REG
	prog.From.Reg = x86.REG_AX
	return prog
}

func addImmediateByte(builder *asm.Builder, in int32) *obj.Prog {
	prog := builder.NewProg()
	prog.As = x86.AADDB
	prog.To.Type = obj.TYPE_REG
	prog.To.Reg = x86.REG_AL
	prog.From.Type = obj.TYPE_CONST
	prog.From.Offset = int64(in)
	return prog
}

func movImmediateByte(builder *asm.Builder, reg int16, in int32) *obj.Prog {
	prog := builder.NewProg()
	prog.As = x86.AMOVB
	prog.To.Type = obj.TYPE_REG
	prog.To.Reg = reg
	prog.From.Type = obj.TYPE_CONST
	prog.From.Offset = int64(in)
	return prog
}

func main() {
	b, _ := asm.NewBuilder("amd64", 64)
	b.AddInstruction(noop(b))
	b.AddInstruction(movImmediateByte(b, x86.REG_AL, 16))
	b.AddInstruction(addImmediateByte(b, 16))
	fmt.Printf("Bin: %x\n", b.Assemble())
}

```

# Working out the parameters of `obj.Prog`

This took me some time to work out, so I'll write a bit here.

## Use these references

 * `obj.Prog` - [godoc](https://godoc.org/github.com/golang/go/src/cmd/internal/obj#Prog)
  * Some instructions (like NOP, JMP) are abstract per-platform & can be found [here](https://godoc.org/github.com/golang/go/src/cmd/internal/obj#As)

 * (for amd64) `x86 pkg-constants` - [registers & instructions](https://godoc.org/github.com/golang/go/src/cmd/internal/obj/x86#pkg-constant)

## Instruction constants have a naming scheme

Instructions are defined as constants in the package for the relavant architecture, and have an 'A' prefix and a size suffix.

For example, the MOV instruction for 64 bits of data is `AMOVQ` (well, at least in amd64).

## Search the go source for usage of a given instruction

For example, if I wanted to work out how to emit the MOV instruction for 64bits, I would search the go source on github for `AMOVQ` or `x86.AMOVQ`. Normally, you see find a few examples where the compiler backend fills in a `obj.Prog` structure, and you follow it's lead.
