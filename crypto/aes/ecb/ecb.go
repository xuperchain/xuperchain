// Copyright 2016 Andre Burgaud. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Electronic Code Book (ECB) mode.

// Implemented for legacy purpose only. ECB should be avoided
// as a mode of operation. Favor other modes available
// in the Go crypto/cipher package (i.e. CBC, GCM, CFB, OFB or CTR).

// See NIST SP 800-38A, pp 9

// The source code in this file is a modified copy of
// https://golang.org/src/crypto/cipher/cbc.go
// and released under the following
// Go Authors copyright and license:

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found at https://golang.org/LICENSE

// Package ecb implements block cipher mode of encryption ECB (Electronic Code
// Book) functions. This is implemented for legacy purposes only and should not
// be used for any new encryption needs. Use CBC (Cipher Block Chaining) instead.
package ecb

import (
	"crypto/cipher"
)

type ecb struct {
	b         cipher.Block
	blockSize int
	tmp       []byte
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
		tmp:       make([]byte, b.BlockSize()),
	}
}

type ecbEncrypter ecb

// NewECBEncrypter returns a BlockMode which encrypts in elecronic codebook (ECB)
// mode, using the given Block (Cipher).
func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {

	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}

	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}

	for len(src) > 0 {
		x.b.Encrypt(dst[:x.blockSize], src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

// NewECBDecrypter returns a BlockMode which decrypts in electronic codebook (ECB)
// mode, using the given Block.
func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	if len(src) == 0 {
		return
	}

	for len(src) > 0 {
		x.b.Decrypt(dst[:x.blockSize], src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}

}
