package ringsign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	mathRand "math/rand"

	schnorr_sign "github.com/xuperchain/xuperunion/crypto/client/schnorr/sign"
	"github.com/xuperchain/xuperunion/crypto/common"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"github.com/xuperchain/xuperunion/hdwallet/rand"
)

// define errors
var (
	ErrGenerateRingSignature     = errors.New("failed to generate ring signature")
	ErrTooSmallNumOfkeys         = errors.New("The total num of keys should be greater than one")
	ErrCurveParamNil             = errors.New("curve input param is nil")
	ErrNotExactTheSameCurveInput = errors.New("the curve is not same as curve of members")
	ErrInvalidInputParams        = errors.New("invalid input")
	ErrKeyParamNotMatch          = errors.New("key param not match")
)

// ring sign consts
const (
	MinimumParticipant = 2
)

//type RingSignature struct {
//	members []*ecdsa.PublicKey
//	e       []byte
//	s       [][]byte
//}

//type PublicKeyFactor struct {
//	X, Y *big.Int
//}
//
//type RingSignature struct {
//	elliptic.Curve
//	Members []*PublicKeyFactor
//	E       *big.Int
//	S       []*big.Int
//}

// Sign : Schnorr Ring Signature use a particular function, the same as Schnorr signatures, defined as:
// H'(m, s(i), e(i)) = H(m || s(i)*G + e(i)*P(i))
//
// To verify the ring signature, check that the result of H'(m, s(i), e(i)) is equal to e((i+1) % R).
// Which means that: H(m || s(i)*G + e(i)*P(i)) = e((i+1) % R)
//
// P is the public key. P(i) = x(i)*G  ---- x is the scalar factor of the private key
// R is the total number of P(0), P(1), ..., P(R-1) which are all the public keys participate in generating the ring signature.
// H is a hash function, for instance SHA256 or SM3.
// N is the order of the curve.
// r is the index of the actual signer located in the public keys of the ring.
//
// This is the process:
//
// 1. Signer choose a random int index r within [0:lenOfRing-1] to hide his public key
// 2. Signer choose a random number k within [1:N-1]
// 3. Signer use k to compute the next index e: e((r+1)%R) = H(m || k*G), %R in case of (r+1) == R
// 4. Then repeat the procedure to compute every e within the ring until reach index r
//       for i := (r+1)%R; i != r; i++%R:
//	        Choose a random number s((r+1)%R), i.e. s(i) within [1:N-1]
//	        Then compute e((r+2)%R), i.e. e((i+1) % R) = H(m || s(i)*G + e(i)*P(i))
//
// 5. Now we get e((r+1)%R), s((r+1)%R), ..., e(r), except s(r), which means the ring has a gap.
//
//    In order to close the ring, or say fulfill the gap, someone must have a private key which
//    corresponding public key exists within the public key set.
//
//    Finally we use e(r), k and x(r) to compute corresponding s(r):
//       Compute s(r) = k - e(r)*x(r)
//
// 5. Now we get everything, the ring is closed.
//    The Output signature: (P(0), ..., P(1), e(0), s(0), ..., s(r))
//
// It is impossible for us to know who signed the signature, as everyone can use his private key
// to fulfill the gap and close the ring.
//func Sign(keys []*ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, message []byte) (*RingSignature, error) {
func Sign(keys []*ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error) {
	// params check
	err := checkRingSignParams(keys, privateKey, message)
	if err != nil {
		return nil, err
	}

	// 1. hide the signer's private key within all the public keys
	// choose a random integer signer index r: 0 <= r < lenOfRing
	lenOfRing := len(keys) + 1
	seed, err := rand.GenerateSeedWithStrengthAndKeyLen(rand.KeyStrengthHard, rand.KeyLengthInt64)
	if err != nil {
		return nil, err
	}

	r := mathRand.New(mathRand.NewSource(int64(binary.BigEndian.Uint64(seed))))
	signerIndex := r.Intn(lenOfRing)
	// log.Printf("signerIndex: %d", signerIndex)
	temp := append([]*ecdsa.PublicKey{}, keys[signerIndex:]...)
	keys = append(keys[:signerIndex], &privateKey.PublicKey)
	keys = append(keys, temp...)

	// 1. Signer(index r) choose a random number k within [1:N-1]
	k, err := rand.GenerateSeedWithStrengthAndKeyLen(rand.KeyStrengthHard, rand.KeyLengthInt32)
	if err != nil {
		return nil, err
	}

	// 2. Signer use k to compute the next index e: e((r+1)%R) = H(m || k*G), %R in case of (r+1) == R
	//	lenOfRing := len(keys)
	allOfE := make([]*big.Int, lenOfRing)
	allOfS := make([]*big.Int, lenOfRing)
	curve := privateKey.Curve

	x, y := curve.ScalarBaseMult(k)
	allOfE[(signerIndex+1)%lenOfRing] = new(big.Int).SetBytes(hash.UsingSha256(append(message, elliptic.Marshal(curve, x, y)...)))

	// 3. Then repeat the procedure to compute every e within the ring until reach index r, r = signerIndex
	// for i:=(r+1)%R; i!=r; i++%R
	for i := (signerIndex + 1) % lenOfRing; i != signerIndex; i = (i + 1) % lenOfRing {
		// Choose a random number s((r+1)%R), i.e. s(i) within [1:N-1]
		s, err := rand.GenerateSeedWithStrengthAndKeyLen(rand.KeyStrengthHard, rand.KeyLengthInt32)
		if err != nil {
			return nil, err
		}
		allOfS[i] = new(big.Int).SetBytes(s)

		// Then compute e((r+2)%R), i.e. e((i+1) % R) = H(m || s(i)*G + e(i)*P(i))
		x1, y1 := curve.ScalarBaseMult(s)
		x2, y2 := curve.ScalarMult(keys[i].X, keys[i].Y, allOfE[i].Bytes())

		x, y := curve.Add(x1, y1, x2, y2)
		allOfE[(i+1)%lenOfRing] = new(big.Int).SetBytes(hash.UsingSha256(append(message, elliptic.Marshal(curve, x, y)...)))
	}

	// 4. Now we get e((r+1)%R), s((r+1)%R), ..., e(r), except s(r), which means the ring has a gap.
	// Finally we use e(r), k and x(r) to compute corresponding s(r):
	// Compute s(r) = k - e(r)*x(r), i.e. compute s(signerIndex) = k - e(signerIndex)*x(signerIndex)
	selfE := allOfE[signerIndex]
	selfK := new(big.Int).SetBytes(k)

	selfS, err := schnorr_sign.ComputeSByKEX(curve, selfK, selfE, privateKey.D)
	if err != nil {
		return nil, ErrGenerateRingSignature
	}

	allOfS[signerIndex] = selfS

	// --- start for test
	// e0 = H(m || s1 * G + e1 * P1)
	//	x1, y1 := curve.ScalarBaseMult(allOfS[signerIndex].Bytes())
	//	x2, y2 := curve.ScalarMult(keys[signerIndex].X, keys[signerIndex].Y, allOfE[signerIndex].Bytes())
	//	x, y = curve.Add(x1, y1, x2, y2)
	//	newE := new(big.Int).SetBytes(hash.HashUsingSha256(append(message, elliptic.Marshal(curve, x, y)...)))
	//	log.Printf("E[%d]: %d", (signerIndex+1)%lenOfRing, newE)
	// --- end for test

	sigRing := &common.RingSignature{}
	//	sigRing.Curve = curve
	sigRing.CurveName = curve.Params().Name
	keyFactors := make([]*common.PublicKeyFactor, len(keys))
	for index, key := range keys {
		//		keyFactor := &PublicKeyFactor{}
		keyFactor := new(common.PublicKeyFactor)
		keyFactor.X = key.X
		keyFactor.Y = key.Y
		keyFactors[index] = keyFactor
	}
	sigRing.Members = keyFactors
	sigRing.E = allOfE[0]
	sigRing.S = allOfS

	// 转换json
	sigContent, err := json.Marshal(sigRing)
	//	sigContent, err := asn1.Marshal(sigRing)
	//	sigContent, err := asn1.Marshal(common.RingSignature{sigRing.Curve, sigRing.Members, sigRing.E, sigRing.S})
	if err != nil {
		return nil, err
	}

	// 组装超级签名
	xuperSig := &common.XuperSignature{
		SigType:    common.SchnorrRing,
		SigContent: sigContent,
	}

	sig, err := json.Marshal(xuperSig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func checkRingSignParams(keys []*ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, message []byte) error {
	if privateKey == nil {
		return fmt.Errorf("Invalid privateKey. PrivateKey must not be nil")
	}

	// members of ring should be no less than two
	if len(keys) < MinimumParticipant {
		return ErrTooSmallNumOfkeys
	}

	// all the public keys need to use the same curve
	// 所有参与者需要使用同一条椭圆曲线
	curveCheckResult := checkCurveForPublicKeys(keys)
	if curveCheckResult == false {
		return ErrNotExactTheSameCurveInput
	}

	curve := keys[0].Curve

	if curve != privateKey.Curve {
		return ErrNotExactTheSameCurveInput
	}

	return nil
}

// check whether all the public keys are using the same curve
// 检查是否所有的环签名验证参与者使用的都是同一条椭圆曲线
func checkCurveForPublicKeys(keys []*ecdsa.PublicKey) bool {
	curve := keys[0].Curve

	for _, key := range keys {
		if curve != key.Curve {
			return false
		}
	}

	return true
}

// 判断传入的公钥数组是否精准的匹配了环签名中的公钥内容
func checkPublicKeysMatchSignature(keys []*ecdsa.PublicKey, signature *common.RingSignature) bool {
	if len(keys) != len(signature.Members) {
		log.Printf("key length and ring key length does not match")
		return false
	}

	publicKeyMap := make(map[string]string)
	for _, key := range keys {
		publicKeyMap[key.X.String()] = key.Y.String()
		//		log.Printf("publicKeys.X[%d], publicKeys.Y[%d]", key.X, key.Y)
	}

	for _, key := range signature.Members {
		publicKeyMapY, isExists := publicKeyMap[key.X.String()]
		//		log.Printf("publicKeys.Y[%s], isExists[%v]", publicKeyMapY, isExists)
		if !isExists || publicKeyMapY != key.Y.String() {
			log.Printf("X[%d], publicKeys.Y[%s], signature.Members.Y[%d]", key.X, publicKeyMapY, key.Y)
			return false
		}
	}

	return true
}

// Verify check the ring signature
func Verify(keys []*ecdsa.PublicKey, signature, message []byte) (bool, error) {
	if len(keys) < MinimumParticipant {
		return false, ErrTooSmallNumOfkeys
	}

	// Sanity check begins
	// Empty message
	if len(message) == 0 {
		return false, nil
	}

	// nil signature
	if signature == nil {
		return false, nil
	}

	sig := new(common.RingSignature)
	err := json.Unmarshal(signature, sig)
	//	_, err := asn1.Unmarshal(signature, sig)

	if err != nil {
		return false, fmt.Errorf("Failed unmashalling schnorr ring signature [%s]", err)
	}

	// log.Printf("ring sig: %v", sig)

	// len check
	if len(sig.S) != len(sig.Members) {
		return false, nil
	}

	// e sanity check
	if sig.E == nil {
		return false, nil
	}

	//	curve := sig.Curve

	// 参与者和验签公钥需要使用同一条椭圆曲线
	if keys[0].Curve.Params().Name != sig.CurveName {
		return false, ErrNotExactTheSameCurveInput
	}

	curve := keys[0].Curve

	// 所有参与者需要使用同一条椭圆曲线
	curveCheckResult := checkCurveForPublicKeys(keys)
	if curveCheckResult == false {
		return false, ErrNotExactTheSameCurveInput
	}

	// 判断传入的公钥数组是否精准的匹配了环签名中的公钥内容
	keyMatchCheckResult := checkPublicKeysMatchSignature(keys, sig)
	if keyMatchCheckResult == false {
		return false, ErrKeyParamNotMatch
	}

	e := sig.E

	lenOfRing := len(sig.Members)
	for i := 0; i < lenOfRing; i++ {
		// compute h(m|| s * g + e * p)
		// s*g
		x1, y1 := curve.ScalarBaseMult(sig.S[i].Bytes())
		// e*p
		x2, y2 := curve.ScalarMult(sig.Members[i].X, sig.Members[i].Y, e.Bytes())

		x, y := curve.Add(x1, y1, x2, y2)
		e = new(big.Int).SetBytes(hash.UsingSha256(append(message, elliptic.Marshal(curve, x, y)...)))
	}

	// log.Printf("ring sig.E: %d", sig.E)
	// log.Printf("ring sig E should be: %d", e)

	if e.Cmp(sig.E) != 0 {
		return false, nil
	}

	return true, nil
}
