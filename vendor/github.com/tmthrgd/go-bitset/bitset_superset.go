// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// Copyright 2014 Will Fitzgerald. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitset

import "github.com/tmthrgd/go-bitset/internal/bitwise"

func (b Bitset) IsSuperSet(b1 Bitset) bool {
	return bitwise.AndEq(b, b1)
}

func (b Bitset) IsStrictSuperSet(b1 Bitset) bool {
	return b.IsSuperSet(b1) && b.Count() > b1.Count()
}
