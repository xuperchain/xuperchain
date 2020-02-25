/*
Copyright Baidu Inc. All Rights Reserved.
*/

package schnorr_sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	//	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/xuperchain/xuperchain/core/crypto/common"

	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/sm3"
)

var (
	GenerateSignatureError = errors.New("Failed to generate the schnorr signature, s = 0 happened.")
	EmptyMessageError      = errors.New("The message to be signed should not be empty")
)

// Schnorr signatures use a particular function, defined as:
// H'(m, s, e) = H(m || s * G + e * P)
//
// H is a hash function, for instance SHA256 or SM3.
// s and e are 2 numbers forming the signature itself.
// m is the message to sign.
// P is the public key.
//
// To verify the signature, check that the result of H'(m, s, e) is equal to e.
// Which means that: H(m || s * G + e * P) = e
//
// It's impossible for the others to find such a pair of (s, e) but the signer himself.
// This is because: P = x * G
// So the signer is able to get this equation: H(m || s * G + e * x * G) = e = H(m || (s + e * x) * G)
// It can be considered as:  H(m || k * G) = e, where k = s + e * x
//
// This is the original process:
// 1. Choose a random number k
// 2. Compute e = H(m || k * G)
// 3. Because k = s + e * x, k and x (the key factor of the private key) are already known, we can compute s
// 4. Now we get the SchnorrSignature (e, s)
//
// Note that there is a potential risk for the private key, which also exists in the ECDSA algorithm:
// "The number k must be random enough."
// If not, say the same k has been used twice or the second k can be predicted by the first k,
// the attacker will be able to retrieve the private key (x)
// This is because:
// 1. If the same k has been used twice:
//    k = s0 + e0 * x = s1 + e1 * x
// The attacker knows: x = (s0 - s1) / (e1 - e0)
//
// 2. If the second k1 can be predicted by the first k0:
//    k0 = s0 + e0 * x
//    k1 = s1 + e1 * x
// The attacker knows: x = (k1 - k0 + s0 - s1) / (e1 - e0)
//
// So the final process is:
// 1. Compute k = H(m || x)
//    This makes k unpredictable for anyone who do not know x,
//    therefor it's impossible for the attacker to retrive x by breaking the random number generator of the system,
//    which has happend in the Sony PlayStation 3 firmware attack.
// 2. Compute e = H(m || k * G)
// 3. Because k = s + e * x, k and x (the key factor of the private key) are already known,
//    we can compute s = k - e * x
//    Note that if k < e * x, S may be negative, but we need S to be positive.
//    As when we compute e, e = H(m || s * G + e * P) and N * G = 0, and x < N
//    We can change s = k - e * x + e * N, which will guarantee that s will be positive.
// 4. Now we get the SchnorrSignature (e, s)
func Sign(privateKey *ecdsa.PrivateKey, message []byte) (schnorrSignature []byte, err error) {
	if privateKey == nil {
		return nil, fmt.Errorf("Invalid privateKey. PrivateKey must not be nil.")
	}

	// 1. Compute k = H(m || x)
	k := sm3.Sm3Sum(append(message, privateKey.D.Bytes()...))

	// 2. Compute e = H(m || k * G)
	// 2.1 compute k * G
	curve := privateKey.Curve
	x, y := curve.ScalarBaseMult(k)
	// 2.2 compute H(m || k * G)
	e := sm3.Sm3Sum(append(message, elliptic.Marshal(curve, x, y)...))

	// 3. k = s + e * x, so we can compute s = k - e * x
	intK := new(big.Int).SetBytes(k)
	intE := new(big.Int).SetBytes(e)

	intS, err := ComputeSByKEX(curve, intK, intE, privateKey.D)
	if err != nil {
		return nil, GenerateSignatureError
	}

	//	return marshalSchnorrSignature(intE, intS)

	// generate the schnorr signature：(sum(S), R)
	// 生成Schnorr签名：(sum(S), R)
	schnorrSig := &common.SchnorrSignature{
		E: intE,
		S: intS,
	}
	// convert the signature to json format
	// 将签名格式转换json
	sigContent, err := json.Marshal(schnorrSig)
	if err != nil {
		return nil, err
	}

	// construct the XuperSignature
	// 组装超级签名
	xuperSig := &common.XuperSignature{
		SigType:    common.Schnorr,
		SigContent: sigContent,
	}

	sig, err := json.Marshal(xuperSig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func ComputeSByKEX(curve elliptic.Curve, k, e, x *big.Int) (*big.Int, error) {
	intS := new(big.Int).Sub(k, new(big.Int).Mul(e, x))

	// if intS is a negative big int.
	// we do little trick by exploiting the equation: (s + eN) * G = s * G, as N * G = 0
	if intS.Sign() == -1 {
		intS = intS.Add(intS, new(big.Int).Mul(e, curve.Params().N))

		// length of (s + eN) may be too long, use DivMod(s+eN, N) to reduce the value of (s + eN)
		// Because (s + eN) % N = s - intval(eN / N) * N
		// intS will be positive and 0 <= intS < |curve.Params().N|
		_, intS = intS.DivMod(intS, curve.Params().N, new(big.Int))

		// if intS == 0 happened after DivMod...
		if intS.Sign() != 1 {
			return nil, GenerateSignatureError
		}
	}

	return intS, nil
}

//func marshalSchnorrSignature(e, s *big.Int) ([]byte, error) {
//	return asn1.Marshal(SchnorrSignature{e, s})
//}

//func unmarshalSchnorrSignature(rawSig []byte) (*SchnorrSignature, error) {
//	sig := new(SchnorrSignature)
//	_, err := asn1.Unmarshal(rawSig, sig)
//	if err != nil {
//		return nil, fmt.Errorf("Failed unmashalling schnorr signature [%s]", err)
//	}
//
//	// Validate sig format
//	if sig.E == nil {
//		return nil, fmt.Errorf("Invalid Schnorr signature. E must not be nil.")
//	}
//	if sig.S == nil {
//		return nil, fmt.Errorf("Invalid Schnorr signature. S must not be nil.")
//	}
//
//	if sig.E.Sign() != 1 {
//		return nil, errors.New("Invalid Schnorr signature. E must be larger than zero")
//	}
//	if sig.S.Sign() != 1 {
//		return nil, errors.New("Invalid Schnorr signature. S must be larger than zero")
//	}
//
//	return sig, nil
//}

// In order to verify the signature, only need to check the equation:
// H'(m, s, e) = H(m || s * G + e * P) = e
// i.e. whether e is equal to H(m || s * G + e * P)
func Verify(publicKey *ecdsa.PublicKey, sig []byte, message []byte) (valid bool, err error) {
	//	signature, err := unmarshalSchnorrSignature(sig)
	signature := new(common.SchnorrSignature)
	err = json.Unmarshal(sig, signature)
	if err != nil {
		return false, fmt.Errorf("Failed unmashalling schnorr signature [%s]", err)
	}

	// 1. compute h(m|| s * g + e * p)
	// 1.1 compute s * g
	curve := publicKey.Curve
	x1, y1 := curve.ScalarBaseMult(signature.S.Bytes())

	// 1.2 compute e * p
	x2, y2 := curve.ScalarMult(publicKey.X, publicKey.Y, signature.E.Bytes())

	// 1.3 compute s * g + e * p
	x, y := curve.Add(x1, y1, x2, y2)

	e := sm3.Sm3Sum(append(message, elliptic.Marshal(curve, x, y)...))

	// 2. check the equation
	//	return bytes.Equal(e, signature.E.Bytes()), nil
	intE := new(big.Int).SetBytes(e)

	if intE.Cmp(signature.E) != 0 {
		return false, nil
	}
	return true, nil
}
