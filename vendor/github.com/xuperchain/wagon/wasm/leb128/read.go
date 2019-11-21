// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package leb128 provides functions for reading integer values encoded in the
// Little Endian Base 128 (LEB128) format: https://en.wikipedia.org/wiki/LEB128
package leb128

import (
	"errors"
	"io"
)

// readVarUint reads an unsigned integer of size n defined in https://webassembly.github.io/spec/core/binary/values.html#binary-int
// readVarUint panics if n>64.
func readVarUint(r io.Reader, n uint) (uint64, error) {
	if n > 64 {
		panic(errors.New("leb128: n must <= 64"))
	}
	p := make([]byte, 1)
	var res uint64
	var shift uint
	for {
		_, err := io.ReadFull(r, p)
		if err != nil {
			return 0, err
		}
		b := uint64(p[0])
		switch {
		// note: can not use b < 1<<n, when n == 64, 1<<n will overflow to 0
		case b < 1<<7 && b <= 1<<n-1:
			res += (1 << shift) * b
			return res, nil
		case b >= 1<<7 && n > 7:
			res += (1 << shift) * (b - 1<<7)
			shift += 7
			n -= 7
		default:
			return 0, errors.New("leb128: invalid uint")
		}
	}
}

// readVarint reads a signed integer of size n, defined in https://webassembly.github.io/spec/core/binary/values.html#binary-int
// readVarint panics if n>64.
func readVarint(r io.Reader, n uint) (int64, error) {
	if n > 64 {
		panic(errors.New("leb128: n must <= 64"))
	}
	p := make([]byte, 1)
	var res int64
	var shift uint
	for {
		_, err := io.ReadFull(r, p)
		if err != nil {
			return 0, err
		}
		b := int64(p[0])
		switch {
		case b < 1<<6 && uint64(b) < uint64(1<<(n-1)):
			res += (1 << shift) * b
			return res, nil
		case b >= 1<<6 && b < 1<<7 && uint64(b)+1<<(n-1) >= 1<<7:
			res += (1 << shift) * (b - 1<<7)
			return res, nil
		case b >= 1<<7 && n > 7:
			res += (1 << shift) * (b - 1<<7)
			shift += 7
			n -= 7
		default:
			return 0, errors.New("leb128: invalid int")
		}
	}
}

// ReadVarUint32 reads a LEB128 encoded unsigned 32-bit integer from r, and
// returns the integer value, and the error (if any).
func ReadVarUint32(r io.Reader) (uint32, error) {
	n, err := readVarUint(r, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// ReadVarint32 reads a LEB128 encoded signed 32-bit integer from r, and
// returns the integer value, and the error (if any).
func ReadVarint32(r io.Reader) (int32, error) {
	n, err := readVarint(r, 32)
	if err != nil {
		return 0, err
	}

	return int32(n), nil
}

// ReadVarint64 reads a LEB128 encoded signed 64-bit integer from r, and
// returns the integer value, and the error (if any).
func ReadVarint64(r io.Reader) (int64, error) {
	return readVarint(r, 64)
}

// ReadVarUint64 reads a LEB128 encoded unsigned 64-bit integer from r, and
// returns the integer value, and the error (if any).
func ReadVarUint64(r io.Reader) (uint64, error) {
	return readVarUint(r, 64)
}
