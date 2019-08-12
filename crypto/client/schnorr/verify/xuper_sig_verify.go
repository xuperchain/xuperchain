package verify

import (
	//	"bytes"
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"

	schnorr_ring_sign "github.com/xuperchain/xuperunion/crypto/client/schnorr/ringsign"
	schnorr_sign "github.com/xuperchain/xuperunion/crypto/client/schnorr/sign"
	"github.com/xuperchain/xuperunion/crypto/common"
	"github.com/xuperchain/xuperunion/crypto/config"
	"github.com/xuperchain/xuperunion/crypto/multisign"
	"github.com/xuperchain/xuperunion/crypto/sign"
)

// define errors
var (
	ErrInvalidInputParams        = errors.New("Invalid input params")
	ErrNotExactTheSameCurveInput = errors.New("The private keys of all the keys are not using the the same curve")
	ErrTooSmallNumOfkeys         = errors.New("The total num of keys should be greater than one")
	ErrEmptyMessage              = errors.New("Message to be sign should not be nil")
	ErrInvalidSignature          = errors.New("XuperSignature is invalid")
)

// XuperSigVerify support to verify multiple kinds of signatures
func XuperSigVerify(keys []*ecdsa.PublicKey, signature, message []byte) (bool, error) {
	//	xuperSig, err := unmarshalXuperSignature(signature)
	xuperSig := new(common.XuperSignature)
	err := json.Unmarshal(signature, xuperSig)

	// 说明不是统一超级签名的格式
	if err != nil {
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := sign.VerifyECDSA(keys[0], signature, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		}
	}

	switch xuperSig.SigType {
	// ECDSA签名
	case common.ECDSA:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := sign.XuperVerify(keys[0], xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		}
	// Schnorr签名
	case common.Schnorr:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := schnorr_sign.Verify(keys[0], xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		}
	// Schnorr环签名
	case common.SchnorrRing:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := schnorr_ring_sign.Verify(keys, xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		}
	// 多重签名
	case common.MultiSig:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := multisign.VerifyMultiSig(keys, xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		}
	// 不支持的签名类型
	default:
		err = fmt.Errorf("This XuperSignature type[%v] is not supported in this version", xuperSig.SigType)
		return false, err
	}
}

func unmarshalXuperSignature(rawSig []byte) (*common.XuperSignature, error) {
	sig := new(common.XuperSignature)
	_, err := asn1.Unmarshal(rawSig, sig)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmashal xuper signature [%s]", err)
	}

	// validate xuper sig format
	if sig.SigContent == nil {
		return nil, ErrInvalidSignature
	}

	return sig, nil
}
