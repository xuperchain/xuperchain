// Package sign is the ECDSA sign and verify implementation
package sign

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"

	"github.com/xuperchain/xuperunion/crypto/utils"
)

// SignECDSA sign message using private key
func SignECDSA(k *ecdsa.PrivateKey, msg []byte) (signature []byte, err error) {
	if k.D == nil || k.X == nil || k.Y == nil {
		return nil, fmt.Errorf("Invalid private key")
	}
	r, s, err := ecdsa.Sign(rand.Reader, k, msg)
	if err != nil {
		return nil, err
	}

	return utils.MarshalECDSASignature(r, s)
}

// VerifyECDSA verify message's signature using public key
func VerifyECDSA(k *ecdsa.PublicKey, signature, msg []byte) (valid bool, err error) {
	r, s, err := utils.UnmarshalECDSASignature(signature)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal the ecdsa signature [%s]", err)
	}

	return ecdsa.Verify(k, msg, r, s), nil
}
