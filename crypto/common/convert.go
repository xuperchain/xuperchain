package common

import (
	"encoding/asn1"
	"errors"
	"fmt"
)

// define errors
var (
	ErrInvalidInputParams        = errors.New("Invalid input params")
	ErrNotExactTheSameCurveInput = errors.New("The private keys of all the keys are not using the the same curve")
	ErrTooSmallNumOfkeys         = errors.New("The total num of keys should be greater than one")
	ErrEmptyMessage              = errors.New("Message to be sign should not be nil")
	ErrInvalidSignature          = errors.New("XuperSignature is invalid")
)

// MarshalXuperSignature marshal Xuper signature to []byte
func MarshalXuperSignature(sig *XuperSignature) ([]byte, error) {
	return asn1.Marshal(sig)
}

// unmarshalXuperSignature unmarshal Xuper signature []byte to struct
func unmarshalXuperSignature(rawSig []byte) (*XuperSignature, error) {
	sig := new(XuperSignature)
	_, err := asn1.Unmarshal(rawSig, sig)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmashal xuper signature [%s]", err)
	}

	// validate xuper sig format
	if sig.SigContent == nil {
		return nil, ErrInvalidSignature
	}

	switch sig.SigType {
	// ECDSA签名
	case ECDSA:
	// Schnorr签名
	case Schnorr:
	// Schnorr环签名
	case SchnorrRing:
	// 多重签名
	case MultiSig:
	// 不支持的签名类型
	default:
		err = fmt.Errorf("This XuperSignature type[%v] is not supported in this version", sig.SigType)
		return nil, err
	}

	return sig, nil
}
