// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package utils file implements the Elliptic Curve Digital Signature Algorithm, as
// defined in FIPS 186-3.
//
// This implementation  derives the nonce from an AES-CTR CSPRNG keyed by
// ChopMD(256, SHA2-512(priv.D || entropy || hash)). The CSPRNG key is IRO by
// a result of Coron; the AES-CTR stream is IRO under standard assumptions.
package utils

// References:
//   [NSA]: Suite B implementer's guide to FIPS 186-3,
//     http://www.nsa.gov/ia/_files/ecdsa.pdf
//   [SECG]: SECG, SEC1
//     http://www.secg.org/sec1-v2.pdf

import (
	//	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
)

// A invertible implements fast inverse mod Curve.Params().N
type invertible interface {
	// Inverse returns the inverse of k in GF(P)
	Inverse(k *big.Int) *big.Int
}

// combinedMult implements fast multiplication S1*g + S2*p (g - generator, p - arbitrary point)
type combinedMult interface {
	CombinedMult(bigX, bigY *big.Int, baseScalar, scalar []byte) (x, y *big.Int)
}

const (
	aesIV = "IV for ECDSA CTR"
)

// PublicKey represents an ECDSA public key.
//type PublicKey struct {
//	elliptic.Curve
//	X, Y *big.Int
//}

// PrivateKey represents an ECDSA private key.
//type PrivateKey struct {
//	PublicKey
//	D *big.Int
//}

var one = new(big.Int).SetInt64(1)

// randFieldElement returns a random element of the field underlying the given
// curve using the procedure given in [NSA] A.2.1.
func randFieldElement(c elliptic.Curve, entropy []byte) (k *big.Int, err error) {
	params := c.Params()
	//	b := make([]byte, params.BitSize/8+8)
	//	_, err = io.ReadFull(rand, b)
	//	if err != nil {
	//		return
	//	}

	//	k = new(big.Int).SetBytes(b)
	k = new(big.Int).SetBytes(entropy)
	n := new(big.Int).Sub(params.N, one)
	k.Mod(k, n)
	k.Add(k, one)
	return
}

// GenerateKeyBySeed generates a public and private key pair.
func GenerateKeyBySeed(c elliptic.Curve, seed []byte) (*ecdsa.PrivateKey, error) {
	k, err := randFieldElement(c, seed)
	if err != nil {
		return nil, err
	}

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = k
	priv.PublicKey.X, priv.PublicKey.Y = c.ScalarBaseMult(k.Bytes())
	return priv, nil
}
