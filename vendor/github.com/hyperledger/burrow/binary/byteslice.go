// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package binary

func Fingerprint(slice []byte) []byte {
	fingerprint := make([]byte, 6)
	copy(fingerprint, slice)
	return fingerprint
}

func IsZeros(slice []byte) bool {
	for _, byt := range slice {
		if byt != byte(0) {
			return false
		}
	}
	return true
}

func RightPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}
	padded := make([]byte, l)
	copy(padded[0:len(slice)], slice)
	return padded
}

func LeftPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}
	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)
	return padded
}
