// Package hash including some useful hash utilities
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"

	"golang.org/x/crypto/ripemd160"
)

// UsingSha256 get the hash result of data using SHA256
func UsingSha256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	out := h.Sum(nil)

	return out
}

// DoubleSha256 执行2次SHA256，这是为了防止SHA256算法被攻破。
func DoubleSha256(data []byte) []byte {
	return UsingSha256(UsingSha256(data))
}

// UsingRipemd160 这种hash算法可以缩短长度
func UsingRipemd160(data []byte) []byte {
	h := ripemd160.New()
	h.Write(data)
	out := h.Sum(nil)

	return out
}

func HashUsingHmac512(seed, key []byte) []byte {
	hmac512 := hmac.New(sha512.New, key)
	hmac512.Write(seed)
	out := hmac512.Sum(nil)

	return out
}
