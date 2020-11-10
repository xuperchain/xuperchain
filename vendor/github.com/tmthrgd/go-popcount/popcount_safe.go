// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build appengine

package popcount

import (
	"encoding/binary"
	"math/bits"
)

func countBytesGo(s []byte) (count uint64) {
	for i := 0; i+8 <= len(s); i += 8 {
		x := binary.LittleEndian.Uint64(s[i:])
		count += uint64(bits.OnesCount64(x))
	}

	s = s[len(s)&^7:]

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
