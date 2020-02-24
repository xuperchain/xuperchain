package signature

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xuperchain/xuperchain/core/crypto/common"
	"github.com/xuperchain/xuperchain/core/crypto/config"
	//	"github.com/xuperchain/xuperchain/core/crypto/multisign"
	//	"github.com/xuperchain/xuperchain/core/crypto/schnorr_ring_sign"
	//	"github.com/xuperchain/xuperchain/core/crypto/schnorr_sign"
	"github.com/xuperchain/xuperchain/core/crypto/sign"
	"github.com/xuperchain/xuperchain/core/crypto/utils"

	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/multisign"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/schnorr_ring_sign"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/schnorr_sign"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/sm2"
)

var (
	InvalidInputParamsError        = errors.New("Invalid input params")
	NotExactTheSameCurveInputError = errors.New("The private keys of all the keys are not using the the same curve")
	TooSmallNumOfkeysError         = errors.New("The total num of keys should be greater than one")
	EmptyMessageError              = errors.New("Message to be sign should not be nil")
	InValidSignatureError          = errors.New("XuperSignature is invalid")
)

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
			//			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
			// TODO: 这块逻辑需要重构来提高代码复用
			r, s, err := utils.UnmarshalECDSASignature(signature)
			if err != nil {
				return false, fmt.Errorf("Failed to unmarshal the signature [%s]", err)
			}

			key := new(sm2.PublicKey)
			key.Curve = sm2.P256Sm2() // elliptic.P256()
			key.X = keys[0].X
			key.Y = keys[0].Y

			return sm2.Verify(key, message, r, s), nil
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
		}

		return false, err
	}

	switch xuperSig.SigType {
	// ECDSA签名
	case common.ECDSA:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := sign.VerifyECDSA(keys[0], xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
		}
	// Schnorr签名
	case common.Schnorr:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := schnorr_sign.Verify(keys[0], xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			//			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
			verifyResult, err := schnorr_sign.Verify(keys[0], xuperSig.SigContent, message)
			return verifyResult, err
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
		}
	// Schnorr环签名
	case common.SchnorrRing:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := schnorr_ring_sign.Verify(keys, xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			//			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
			verifyResult, err := schnorr_ring_sign.Verify(keys, xuperSig.SigContent, message)
			return verifyResult, err
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
		}
	// 多重签名
	case common.MultiSig:
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := multisign.VerifyMultiSig(keys, xuperSig.SigContent, message)
			return verifyResult, err
		case config.CurveGm: // 国密
			verifyResult, err := multisign.VerifyMultiSig(keys, xuperSig.SigContent, message)
			return verifyResult, err
			//			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet.", keys[0].Params().Name)
		}
	// 不支持的签名类型
	default:
		err = fmt.Errorf("This XuperSignature type[%v] is not supported in this version.", xuperSig.SigType)
		return false, err
	}

	return false, nil
}

func unmarshalXuperSignature(rawSig []byte) (*common.XuperSignature, error) {
	sig := new(common.XuperSignature)
	_, err := asn1.Unmarshal(rawSig, sig)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmashal xuper signature [%s]", err)
	}

	// validate xuper sig format
	if sig.SigContent == nil {
		return nil, InValidSignatureError
	}

	return sig, nil
}
