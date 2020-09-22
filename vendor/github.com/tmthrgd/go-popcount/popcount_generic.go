// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build !amd64 gccgo appengine

package popcount

const usePOPCNT = false

// CountBytes function counts number of non-zero bits in slice of 8bit unsigned integers.
func CountBytes(s []byte) uint64 {
	return countBytesGo(s)
}

// CountSlice64 function counts number of non-zero bits in slice of 64bit unsigned integers.
func CountSlice64(s []uint64) uint64 {
	return countSlice64Go(s)
}
