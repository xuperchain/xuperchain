// Copyright 2016 Tom Thorogood. All rights reserved.
// Use of this source code is governed by a
// Modified BSD License license that can be found in
// the LICENSE file.

// +build amd64,!gccgo,!appengine

// Package memset is an efficient memset implementation for Golang.
package memset

import "golang.org/x/sys/cpu"

var useAVX = cpu.X86.HasAVX

// Memset sets each byte in data to value.
func Memset(data []byte, value byte) {
	if len(data) == 0 {
		return
	}

	memsetAsm(&data[0], uint64(len(data)), value)
}

// This function is implemented in memset_amd64.s
//go:noescape
func memsetAsm(dst *byte, len uint64, value byte)
