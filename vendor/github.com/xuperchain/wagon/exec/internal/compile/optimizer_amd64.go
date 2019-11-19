// Copyright 2019 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package compile

import (
	"math"

	asm "github.com/twitchyliquid64/golang-asm"
	"github.com/twitchyliquid64/golang-asm/obj"
	"github.com/twitchyliquid64/golang-asm/obj/x86"
)

func peepholeOptimizeAMD64(builder *asm.Builder) error {
	inst := builder.Root()
	for ; inst.Link != nil; inst = inst.Link {
		// Replace mov <reg>, 0 with xor.
		if inst.As == x86.AMOVQ && inst.To.Type == obj.TYPE_REG &&
			inst.From.Type == obj.TYPE_CONST && inst.From.Offset == 0 {
			inst.As = x86.AXORL
			inst.From = inst.To
		}

		// If we are loading a constant to a register and its less than 32bit,
		// use the 32bit version (its shorter).
		if n := inst.From.Offset; inst.As == x86.AMOVQ && inst.From.Type == obj.TYPE_CONST && (n > 0 && n < math.MaxInt32) {
			inst.As = x86.AMOVL
		}

		// If its an add by an immediate 1, convert to increment.
		if inst.As == x86.AADDQ && inst.From.Type == obj.TYPE_CONST && inst.From.Offset == 1 {
			inst.As = x86.AINCQ
			inst.From = obj.Addr{}
		}
	}

	return nil
}
