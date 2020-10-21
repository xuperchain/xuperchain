// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

// Package bytetest is an efficient byte test implementation for Golang.
package bytetest

// Test returns true iff each byte in data is equal to value.
func Test(data []byte, value byte) bool {
	if len(data) == 0 {
		return true
	}

	return testAsm(&data[0], uint64(len(data)), value)
}

// This function is implemented in test_amd64.s
//go:noescape
func testAsm(src *byte, len uint64, value byte) (ret bool)
