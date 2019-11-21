// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wasm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"unicode/utf8"

	"github.com/xuperchain/wagon/wasm/leb128"
)

// to avoid memory attack
const maxInitialCap = 10 * 1024

func getInitialCap(count uint32) uint32 {
	if count > maxInitialCap {
		return maxInitialCap
	}
	return count
}

func readBytes(r io.Reader, n uint32) ([]byte, error) {
	if n == 0 {
		return nil, nil
	}
	limited := io.LimitReader(r, int64(n))
	buf := &bytes.Buffer{}
	num, _ := buf.ReadFrom(limited)
	if num == int64(n) {
		return buf.Bytes(), nil
	}
	return nil, io.ErrUnexpectedEOF
}

func writeByte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	return err
}

func ReadByte(r io.Reader) (byte, error) {
	p := make([]byte, 1)
	_, err := r.Read(p)
	return p[0], err
}

func readBytesUint(r io.Reader) ([]byte, error) {
	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return nil, err
	}
	return readBytes(r, n)
}

func readUTF8String(r io.Reader, n uint32) (string, error) {
	bytes, err := readBytes(r, n)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(bytes) {
		return "", errors.New("wasm: invalid utf-8 string")
	}
	return string(bytes), nil
}

func readUTF8StringUint(r io.Reader) (string, error) {
	n, err := leb128.ReadVarUint32(r)
	if err != nil {
		return "", err
	}
	return readUTF8String(r, n)
}

func readU32(r io.Reader) (uint32, error) {
	var buf [4]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf[:]), nil
}

func readU64(r io.Reader) (uint64, error) {
	var buf [8]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}
