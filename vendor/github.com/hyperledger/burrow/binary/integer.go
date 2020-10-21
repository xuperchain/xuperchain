// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	"math"
	"math/big"
)

var big1 = big.NewInt(1)
var Big256 = big.NewInt(256)

// Returns whether a + b would be a uint64 overflow
func IsUint64SumOverflow(a, b uint64) bool {
	return math.MaxUint64-a < b
}

// Converts a possibly negative big int x into a positive big int encoding a twos complement representation of x
// truncated to 32 bytes
func U256(x *big.Int) *big.Int {
	return ToTwosComplement(x, Word256Bits)
}

// Interprets a positive big.Int as a 256-bit two's complement signed integer
func S256(x *big.Int) *big.Int {
	return FromTwosComplement(x, Word256Bits)
}

// Convert a possibly negative big.Int x to a positive big.Int encoded in two's complement
func ToTwosComplement(x *big.Int, n uint) *big.Int {
	// And treats negative arguments a if they were twos complement encoded so we end up with a positive number here
	// with the twos complement bit pattern
	return new(big.Int).And(x, andMask(n))
}

// Interprets a positive big.Int as a n-bit two's complement signed integer
func FromTwosComplement(x *big.Int, n uint) *big.Int {
	signBit := int(n) - 1
	if x.Bit(signBit) == 0 {
		// Sign bit not set => value (v) is positive
		// x = |v| = v
		return x
	} else {
		// Sign bit set => value (v) is negative
		// x = 2^n - |v|
		b := new(big.Int).Lsh(big1, n)
		// v = -|v| = x - 2^n
		return new(big.Int).Sub(x, b)
	}
}

// Treats the positive big int x as if it contains an embedded n bit signed integer in its least significant
// bits and extends that sign
func SignExtend(x *big.Int, n uint) *big.Int {
	signBit := n - 1
	// single bit set at sign bit position
	mask := new(big.Int).Lsh(big1, signBit)
	// all bits below sign bit set to 1 all above (including sign bit) set to 0
	mask.Sub(mask, big1)
	if x.Bit(int(signBit)) == 1 {
		// Number represented is negative - set all bits above sign bit (including sign bit)
		return x.Or(x, mask.Not(mask))
	} else {
		// Number represented is positive - clear all bits above sign bit (including sign bit)
		return x.And(x, mask)
	}
}

func andMask(n uint) *big.Int {
	x := new(big.Int)
	return x.Sub(x.Lsh(big1, n), big1)
}
