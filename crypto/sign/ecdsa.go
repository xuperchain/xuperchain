// Package sign is the ECDSA sign and verify implementation
package sign

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperunion/crypto/common"
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

	//	return utils.MarshalECDSASignature(r, s)

	// 生成ECDSA签名：(sum(S), R)
	ecdsaSig := &common.ECDSASignature{
		R: r,
		S: s,
	}

	// 生成超级签名
	// 转换json
	sigContent, err := json.Marshal(ecdsaSig)
	if err != nil {
		return nil, err
	}

	xuperSig := &common.XuperSignature{
		SigType:    common.ECDSA,
		SigContent: sigContent,
	}

	//log.Printf("xuperSig before marshal: %s", xuperSig)

	sig, err := json.Marshal(xuperSig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func VerifyECDSA(k *ecdsa.PublicKey, sig, msg []byte) (valid bool, err error) {
	//	r, s, err := utils.UnmarshalECDSASignature(signature)
	signature := new(common.ECDSASignature)
	err = json.Unmarshal(sig, signature)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal the ecdsa signature [%s]", err)
	}

	return ecdsa.Verify(k, msg, signature.R, signature.S), nil
}

// VerifyECDSA verify message's signature using public key
func VerifyOldECDSA(k *ecdsa.PublicKey, signature, msg []byte) (valid bool, err error) {
	r, s, err := utils.UnmarshalECDSASignature(signature)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal the ecdsa signature [%s]", err)
	}

	return ecdsa.Verify(k, msg, r, s), nil
}
