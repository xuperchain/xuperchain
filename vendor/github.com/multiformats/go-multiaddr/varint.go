package multiaddr

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

// VarintSize returns the size (in bytes) of `num` encoded as a varint.
func VarintSize(num int) int {
	bits := bits.Len(uint(num))
	q, r := bits/7, bits%7
	size := q
	if r > 0 || size == 0 {
		size++
	}
	return size
}

// CodeToVarint converts an integer to a varint-encoded []byte
func CodeToVarint(num int) []byte {
	buf := make([]byte, VarintSize(num))
	n := binary.PutUvarint(buf, uint64(num))
	return buf[:n]
}

// VarintToCode converts a varint-encoded []byte to an integer protocol code
func VarintToCode(buf []byte) int {
	num, _, err := ReadVarintCode(buf)
	if err != nil {
		panic(err)
	}
	return num
}

// ReadVarintCode reads a varint code from the beginning of buf.
// returns the code, and the number of bytes read.
func ReadVarintCode(buf []byte) (int, int, error) {
	num, n := binary.Uvarint(buf)
	if n < 0 {
		return 0, 0, fmt.Errorf("varints larger than uint64 not yet supported")
	}
	return int(num), n, nil
}
