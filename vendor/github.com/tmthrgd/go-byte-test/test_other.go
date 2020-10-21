// Copyright 2017 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build !amd64 gccgo appengine

// Package bytetest is an efficient byte test implementation for Golang.
package bytetest

// Test returns true iff each byte in data is equal to value.
func Test(data []byte, value byte) bool {
	for _, v := range data {
		if v != value {
			return false
		}
	}

	return true
}
