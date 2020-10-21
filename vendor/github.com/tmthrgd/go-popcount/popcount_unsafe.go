// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build !appengine

package popcount

import (
	"encoding/binary"
	"math/bits"
	"runtime"
	"unsafe"
)

const supportsUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "amd64"

func countBytesGo(s []byte) (count uint64) {
	var s64 []uint64

	if len(s) < 8 {
		goto tail
	}

	// Align to 8 byte boundary
	if x := 8 - int(uintptr(unsafe.Pointer(&s[0]))&7); !supportsUnaligned && x != 8 {
		if x >= 4 {
			count += uint64(bits.OnesCount32(binary.LittleEndian.Uint32(s)))
			s = s[4:]
			x -= 4
		}

		if x >= 2 {
			count += uint64(bits.OnesCount16(binary.LittleEndian.Uint16(s)))
			s = s[2:]
			x -= 2
		}

		if x == 1 {
			count += uint64(bits.OnesCount8(s[0]))
			s = s[1:]
		}

		if len(s) < 8 {
			goto tail
		}
	}

	s64 = (*[1 << 27]uint64)(unsafe.Pointer(&s[0]))[:len(s)>>3]
	for _, x := range s64 {
		count += uint64(bits.OnesCount64(x))
	}

	s = s[len(s)&^7:]

tail:
	if len(s) >= 4 {
		count += uint64(bits.OnesCount32(binary.LittleEndian.Uint32(s)))
		s = s[4:]
	}

	if len(s) >= 2 {
		count += uint64(bits.OnesCount16(binary.LittleEndian.Uint16(s)))
		s = s[2:]
	}

	if len(s) == 1 {
		count += uint64(bits.OnesCount8(s[0]))
	}

	return
}
