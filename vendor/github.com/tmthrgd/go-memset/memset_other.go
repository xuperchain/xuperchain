// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64 gccgo appengine

// Package memset is an efficient memset implementation for Golang.
package memset

// Memset sets each byte in data to value.
func Memset(data []byte, value byte) {
	if value == 0 {
		for i := range data {
			data[i] = 0
		}
	} else if len(data) != 0 {
		data[0] = value

		for i := 1; i < len(data); i *= 2 {
			copy(data[i:], data[:i])
		}
	}
}
