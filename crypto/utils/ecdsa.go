package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
)

// ECDSASignature is the structure for marshall signature
type ECDSASignature struct {
	R, S *big.Int
}

// MarshalECDSASignature use DER-encoded ASN.1 octet standard to represent the signature
//与比特币算法一样，基于DER-encoded ASN.1 octet标准，来表达使用椭圆曲线签名算法返回的结果
func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ECDSASignature{r, s})
}

// MarshalPublicKey 将公钥序列化成byte数组
func MarshalPublicKey(publicKey *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
}

// UnmarshalECDSASignature 从封装的签名中拿出原始签名
func UnmarshalECDSASignature(rawSig []byte) (*big.Int, *big.Int, error) {
	sig := new(ECDSASignature)
	_, err := asn1.Unmarshal(rawSig, sig)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmashal the signature [%v] to R & S, and the error is [%s]", rawSig, err)
	}

	if sig.R == nil {
		return nil, nil, errors.New("invalid signature, R is nil")
	}
	if sig.S == nil {
		return nil, nil, errors.New("invalid signature, S is nil")
	}

	if sig.R.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, R must be larger than zero")
	}
	if sig.S.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, S must be larger than zero")
	}

	return sig.R, sig.S, nil
}
