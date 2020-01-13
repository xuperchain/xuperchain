package multisign

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/xuperchain/xuperchain/core/crypto/common"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/hdwallet/rand"
)

//type MultiSignature struct {
//	S []byte
//	R []byte
//}

// define errors
var (
	ErrInvalidInputParams        = errors.New("Invalid input params")
	ErrNotExactTheSameCurveInput = errors.New("The private keys of all the keys are not using the the same curve")
	ErrTooSmallNumOfkeys         = errors.New("The total num of keys should be greater than one")
	ErrEmptyMessage              = errors.New("Message to be sign should not be nil")
	ErrNotValidSignature         = errors.New("Signature is invalid")
)

// multisig consts
const (
	MinimumParticipant = 2
)

//BytesCombine concatenates byte array
func BytesCombine(pBytes ...[]byte) []byte {
	var buffer bytes.Buffer

	for i := 0; i < len(pBytes); i++ {
		buffer.Write(pBytes[i])
	}
	return buffer.Bytes()
}

// MultiSign 生成多重签名的算法如下：
// 1. 生成公私钥对(x1, P1), (x2, P2), ..., (xn, Pn)， x代表私钥中的参数大数D，P代表公钥
// 2. 生成临时随机数(k1, k2, ..., kn)
// 3. 计算：R = k1*G + k2*G + ... + kn*G，G代表基点
// 4. 计算公共公钥：C = P1 + P2 + ... + Pn
// 5. 各方计算：si = ki + HASH(C,R,m) * xi
// 6. 生成多重签名：(s1 + s2 + ... + sn, R)
// MultiSign生成对特定消息的多重签名，所有参与签名的私钥必须使用同一条椭圆曲线
func MultiSign(keys []*ecdsa.PrivateKey, message []byte) ([]byte, error) {
	if len(keys) < MinimumParticipant {
		return nil, ErrTooSmallNumOfkeys
	}

	if len(message) == 0 {
		return nil, ErrEmptyMessage
	}

	// 1. 检验传入的私钥参数(x1, P1), (x2, P2), ..., (xn, Pn) 是否合法
	// x代表大数D，P代表公钥
	// 所有参与者需要使用同一条椭圆曲线
	curveCheckResult := checkCurveForPrivateKeys(keys)
	if curveCheckResult == false {
		return nil, ErrNotExactTheSameCurveInput
	}

	// 2. 生成临时随机数的数组(k1, k2, ..., kn)
	num := len(keys)
	arrayOfK, err := getRandomBytesArray(num)
	if err != nil {
		return nil, err
	}

	// 3. 计算：R = k1*G + k2*G + ... + kn*G
	r := getRUsingRandomBytesArray(keys, arrayOfK)

	// 4. 计算公共公钥：C = P1 + P2 + ... + Pn
	c, err := getSharedPublicKeyForPrivateKeys(keys)
	if err != nil {
		return nil, err
	}

	// 5. 各方计算：S = sum(si)
	// si = ki + HASH(C,R,m) * xi
	s := getS(keys, arrayOfK, c, r, message)

	//	return marshalMultiSignature(s, r)

	// 6. 生成多重签名：(sum(S), R)
	multiSig := &common.MultiSignature{
		S: s,
		R: r,
	}

	// 7. 生成超级签名
	// 转换json
	sigContent, err := json.Marshal(multiSig)
	//	sigContent, err := marshalMultiSignature(s, r)
	if err != nil {
		return nil, err
	}

	xuperSig := &common.XuperSignature{
		SigType:    common.MultiSig,
		SigContent: sigContent,
	}

	// log.Printf("xuperSig before marshal: %s", xuperSig)

	//	sig, err := common.MarshalXuperSignature(xuperSig)
	sig, err := json.Marshal(xuperSig)
	if err != nil {
		return nil, err
	}

	return sig, nil

}

// GenCommonPublicKey using public key to generate common result of multisig
// return (R, C), Array[K], Error
// R is common random key, C is common public key, Array[K] is random K array for each public key
func GenCommonPublicKey(keys []*ecdsa.PublicKey, message []byte) (*common.MultiSigCommon, [][]byte, error) {
	if len(keys) < MinimumParticipant {
		return nil, nil, ErrTooSmallNumOfkeys
	}

	if len(message) == 0 {
		return nil, nil, ErrEmptyMessage
	}

	// 1. 检验传入的公钥参数是否合法
	// 所有参与者需要使用同一条椭圆曲线
	curveCheckResult := checkCurveForPublicKeys(keys)
	if curveCheckResult == false {
		return nil, nil, ErrNotExactTheSameCurveInput
	}

	// 2. 生成临时随机数的数组(k1, k2, ..., kn)
	num := len(keys)
	arrayOfK, err := getRandomBytesArray(num)
	if err != nil {
		return nil, nil, err
	}

	// 3. 计算：R = k1*G + k2*G + ... + kn*G
	r := getRUsingRandomK(keys[0].Curve, arrayOfK)

	// 4. 计算公共公钥：C = P1 + P2 + ... + Pn
	c, err := GetSharedPublicKeyForPublicKeys(keys)
	if err != nil {
		return nil, nil, err
	}
	res := &common.MultiSigCommon{
		R: r,
		C: c,
	}
	return res, arrayOfK, nil
}

// GetPartialSign return partial multisig value of a private key
func GetPartialSign(key *ecdsa.PrivateKey, k []byte, msc *common.MultiSigCommon, msg []byte) ([]byte, error) {
	// 计算HASH(P,R,m)
	hashBytes := hash.UsingSha256(BytesCombine(msc.C, msc.R, msg))

	// 计算HASH(P,R,m) * xi
	tempRHS := new(big.Int).Mul(new(big.Int).SetBytes(hashBytes), key.D)

	// 计算ki + HASH(P,R,m) * xi
	res := new(big.Int).Add(new(big.Int).SetBytes(k), tempRHS)

	return res.Bytes(), nil
}

// MergeMultiSig merge partial signatures into multisig result
func MergeMultiSig(partialS [][]byte, r []byte) ([]byte, error) {
	s := big.NewInt(0)
	// 计算s1 + s2 + ... + sn
	for _, ps := range partialS {
		psi := big.NewInt(0).SetBytes(ps)
		s = s.Add(s, psi)
	}

	// 生成多重签名：(sum(S), R)
	multiSig := &common.MultiSignature{
		S: s.Bytes(),
		R: r,
	}

	// 生成超级签名
	// 转换json
	sigContent, err := json.Marshal(multiSig)
	//	sigContent, err := marshalMultiSignature(s, r)
	if err != nil {
		return nil, err
	}

	xuperSig := &common.XuperSignature{
		SigType:    common.MultiSig,
		SigContent: sigContent,
	}

	// log.Printf("xuperSig before marshal: %s", xuperSig)

	//	sig, err := common.MarshalXuperSignature(xuperSig)
	sig, err := json.Marshal(xuperSig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

//func marshalMultiSignature(s, r []byte) ([]byte, error) {
//	return asn1.Marshal(MultiSignature{s, r})
//}

//func unmarshalMultiSignature(rawSig []byte) (*MultiSignature, error) {
//	sig := new(MultiSignature)
//	_, err := asn1.Unmarshal(rawSig, sig)
//	if err != nil {
//		return nil, fmt.Errorf("Failed unmashalling multi signature [%s]", err)
//	}
//
//	// Validate sig format
//	if sig.S == nil {
//		return nil, fmt.Errorf("Invalid multi signature. S must not be nil.")
//	}
//	if sig.R == nil {
//		return nil, fmt.Errorf("Invalid multi signature. R must not be nil.")
//	}
//
//	return sig, nil
//}

// 计算：S = sum(si)
// si = ki + HASH(C,R,m) * xi
// x代表大数D，也就是私钥的关键参数
func getS(keys []*ecdsa.PrivateKey, arrayOfK [][]byte, c []byte, r []byte, message []byte) []byte {
	num := len(arrayOfK)
	s := big.NewInt(0)
	for i := 0; i < num; i++ {
		// 计算HASH(P,R,m)，这里的hash算法选择国密SM3算法
		hashBytes := hash.UsingSha256(BytesCombine(c, r, message))

		// 计算HASH(P,R,m) * xi
		tempRHS := new(big.Int).Mul(new(big.Int).SetBytes(hashBytes), keys[i].D)

		// 计算ki + HASH(P,R,m) * xi
		res := new(big.Int).Add(new(big.Int).SetBytes(arrayOfK[i]), tempRHS)
		// 6.1 计算s1 + s2 + ... + sn
		s = s.Add(s, res)
	}

	return s.Bytes()
}

// 计算：R = k1*G + k2*G + ... + kn*G
func getRUsingRandomBytesArray(keys []*ecdsa.PrivateKey, arrayOfK [][]byte) []byte {
	num := len(keys)
	curve := keys[0].Curve
	x, y := big.NewInt(0), big.NewInt(0)
	for i := 0; i < num; i++ {
		// 计算K*G
		x1, y1 := curve.ScalarBaseMult(arrayOfK[i])

		// 计算k1*G + k2*G + ...
		x, y = curve.Add(x, y, x1, y1)
	}
	// 计算R，converts a point into the uncompressed form specified in section 4.3.6 of ANSI X9.62
	r := elliptic.Marshal(curve, x, y)

	return r
}

// 计算：R = k1*G + k2*G + ... + kn*G
func getRUsingRandomK(curve elliptic.Curve, arrayOfK [][]byte) []byte {
	num := len(arrayOfK)
	x, y := big.NewInt(0), big.NewInt(0)
	for i := 0; i < num; i++ {
		// 计算K*G
		x1, y1 := curve.ScalarBaseMult(arrayOfK[i])

		// 计算k1*G + k2*G + ...
		x, y = curve.Add(x, y, x1, y1)
	}
	// 计算R，converts a point into the uncompressed form specified in section 4.3.6 of ANSI X9.62
	r := elliptic.Marshal(curve, x, y)

	return r
}

// 生成临时随机数(k1, k2, ..., kn)
func getRandomBytesArray(num int) ([][]byte, error) {
	randomBytesArray := make([][]byte, num)
	for i := 0; i < num; i++ {
		randomBytes, err := rand.GenerateSeedWithStrengthAndKeyLen(rand.KeyStrengthHard, rand.KeyLengthInt32)
		if err != nil {
			return nil, err
		}
		randomBytesArray[i] = randomBytes
	}

	return randomBytesArray, nil
}

// 计算公共公钥：C = P1 + P2 + ... + Pn
func getSharedPublicKeyForPrivateKeys(keys []*ecdsa.PrivateKey) ([]byte, error) {
	num := len(keys)
	curve := keys[0].Curve
	x, y := big.NewInt(0), big.NewInt(0)
	for i := 0; i < num; i++ {
		if keys[i] == nil {
			return nil, ErrInvalidInputParams
		}
		// 计算P1 + P2 + ...
		x, y = curve.Add(keys[i].PublicKey.X, keys[i].PublicKey.Y, x, y)
	}

	// 计算C，converts a point into the uncompressed form specified in section 4.3.6 of ANSI X9.62
	c := elliptic.Marshal(curve, x, y)

	return c, nil
}

// GetSharedPublicKeyForPublicKeys 计算公共公钥：C = P1 + P2 + ... + Pn
func GetSharedPublicKeyForPublicKeys(keys []*ecdsa.PublicKey) ([]byte, error) {
	// 所有参与者需要使用同一条椭圆曲线
	curveCheckResult := checkCurveForPublicKeys(keys)
	if curveCheckResult == false {
		return nil, ErrNotExactTheSameCurveInput
	}

	num := len(keys)
	curve := keys[0].Curve
	x, y := big.NewInt(0), big.NewInt(0)
	for i := 0; i < num; i++ {
		if keys[i] == nil {
			return nil, ErrInvalidInputParams
		}

		x, y = curve.Add(keys[i].X, keys[i].Y, x, y)
	}

	// 计算C，converts a point into the uncompressed form specified in section 4.3.6 of ANSI X9.62
	c := elliptic.Marshal(curve, x, y)

	return c, nil
}

// 检查是否所有的多重签名生成参与者使用的都是同一条椭圆曲线
func checkCurveForPrivateKeys(keys []*ecdsa.PrivateKey) bool {
	curve := keys[0].Curve
	//	curveName := curve.Params().Name
	//	for _, key := range keys {
	//		if curveName != key.Params().Name {
	//			return false
	//		}
	//	}

	for _, key := range keys {
		if curve != key.Curve {
			return false
		}
	}

	return true
}

// 检查是否所有的多重签名验证参与者使用的都是同一条椭圆曲线
func checkCurveForPublicKeys(keys []*ecdsa.PublicKey) bool {
	curve := keys[0].Curve

	for _, key := range keys {
		if curve != key.Curve {
			return false
		}
	}

	return true
}

// VerifyMultiSig 验签算法如下：
// 1. 计算：e = sm3(C,R,m)
// 2. 计算：Rv = sG - eC
// 3. 如果Rv == R则返回true，否则返回false
func VerifyMultiSig(keys []*ecdsa.PublicKey, signature []byte, message []byte) (bool, error) {
	if len(keys) < MinimumParticipant {
		return false, ErrTooSmallNumOfkeys
	}

	//	sig, err := unmarshalMultiSignature(signature)
	sig := new(common.MultiSignature)
	err := json.Unmarshal(signature, sig)
	if err != nil {
		return false, fmt.Errorf("Failed unmashalling multi signature [%s]", err)
	}

	// sig nil check and sig format check
	if sig == nil || len(sig.R) == 0 || len(sig.S) == 0 {
		return false, ErrNotValidSignature
	}

	// empty message
	if len(message) == 0 {
		return false, ErrEmptyMessage
	}

	// 所有参与者需要使用同一条椭圆曲线
	curveCheckResult := checkCurveForPublicKeys(keys)
	if curveCheckResult == false {
		return false, ErrNotExactTheSameCurveInput
	}

	curve := keys[0].Curve

	// 计算公共公钥：C = P1 + P2 + ... + Pn
	c, err := GetSharedPublicKeyForPublicKeys(keys)
	if err != nil {
		return false, err
	}

	// 计算sG
	lhsX, lhsY := curve.ScalarBaseMult(sig.S)

	// 计算e = HASH(P,R,m)，这里的hash算法选择NIST算法
	hashBytes := hash.UsingSha256(BytesCombine(c, sig.R, message))
	// 计算eC,也就是HASH(P,R,m) * C
	x, y := elliptic.Unmarshal(curve, c)
	rhsX, rhsY := curve.ScalarMult(x, y, hashBytes)

	// 计算-eC，如果 eC = (x,y)，则 -eC = (x, -y mod P)
	negativeOne := big.NewInt(-1)
	rhsY = new(big.Int).Mod(new(big.Int).Mul(negativeOne, rhsY), curve.Params().P)

	// 计算Rv = sG - eC
	resX, resY := curve.Add(lhsX, lhsY, rhsX, rhsY)

	// 原始签名中的R
	rX, rY := elliptic.Unmarshal(curve, sig.R)

	// 对比签名是否一致
	if resX.Cmp(rX) == 0 && resY.Cmp(rY) == 0 {
		return true, nil
	}

	return false, nil
}
