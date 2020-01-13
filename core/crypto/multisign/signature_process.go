package multisign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"math/big"

	"github.com/xuperchain/xuperchain/core/crypto/common"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/hdwallet/rand"
)

// GetRandom32Bytes 生成默认随机数Ki
func GetRandom32Bytes() ([]byte, error) {
	randomBytes, err := rand.GenerateSeedWithStrengthAndKeyLen(rand.KeyStrengthHard, rand.KeyLengthInt32)
	if err != nil {
		return nil, err
	}

	return randomBytes, nil
}

// GetRiUsingRandomBytes 计算：Ri = Ki*G
func GetRiUsingRandomBytes(key *ecdsa.PublicKey, k []byte) []byte {
	curve := key.Curve

	// 计算K*G
	x, y := curve.ScalarBaseMult(k)

	// 计算R，converts a point into the uncompressed form specified in section 4.3.6 of ANSI X9.62
	r := elliptic.Marshal(curve, x, y)

	return r
}

// GetRUsingAllRi 计算：R = k1*G + k2*G + ... + kn*G
func GetRUsingAllRi(key *ecdsa.PublicKey, arrayOfRi [][]byte) []byte {
	num := len(arrayOfRi)
	curve := key.Curve
	x, y := big.NewInt(0), big.NewInt(0)
	for i := 0; i < num; i++ {
		// Unmarshal converts a point, serialized by Marshal, into an x, y pair.
		// It is an error if the point is not in uncompressed form or is not on the curve.
		// On error, x = nil.
		x1, y1 := elliptic.Unmarshal(curve, arrayOfRi[i])

		// 计算k1*G + k2*G + ...
		x, y = curve.Add(x, y, x1, y1)
	}
	// 计算R，converts a point into the uncompressed form specified in section 4.3.6 of ANSI X9.62
	r := elliptic.Marshal(curve, x, y)

	return r
}

// GetSiUsingKCRM 计算 si = ki + HASH(C,R,m) * xi
// x代表大数D，也就是私钥的关键参数
func GetSiUsingKCRM(key *ecdsa.PrivateKey, k []byte, c []byte, r []byte, message []byte) []byte {
	// 计算HASH(P,R,m)，这里的hash算法选择NIST算法
	hashBytes := hash.UsingSha256(BytesCombine(c, r, message))

	// 计算HASH(P,R,m) * xi
	tmpResult := new(big.Int).Mul(new(big.Int).SetBytes(hashBytes), key.D)

	// 计算ki + HASH(P,R,m) * xi
	s := new(big.Int).Add(new(big.Int).SetBytes(k), tmpResult)

	return s.Bytes()
}

// GetSUsingAllSi 计算：S = sum(si)
func GetSUsingAllSi(arrayOfSi [][]byte) []byte {
	num := len(arrayOfSi)
	s := big.NewInt(0)
	for i := 0; i < num; i++ {
		// 计算s1 + s2 + ... + sn
		s = s.Add(s, new(big.Int).SetBytes(arrayOfSi[i]))
	}

	return s.Bytes()
}

// GenerateMultiSignSignature 生成多重签名的流程如下：
//1. 各方分别生成自己的随机数Ki(K1, K2, ..., Kn) --- func getRandomBytes() ([]byte, error)
//2. 各方计算自己的 Ri = Ki*G，G代表基点 --- func getRiUsingRandomBytes(key *ecdsa.PublicKey, k []byte) []byte
//3. 发起者收集Ri，计算：R = sum(Ri) --- func getRUsingAllRi(key *ecdsa.PublicKey, arrayOfRi [][]byte) []byte
//4. 发起者收集公钥Pi，计算公共公钥：C = P1 + P2 + ... + Pn --- func getSharedPublicKeyForPrivateKeys(keys []*ecdsa.PrivateKey) ([]byte, error)
//5. 各方计算自己的Si：si = Ki + HASH(C,R,m) * xi，x代表私钥中的参数大数D
// --- func getSiUsingKCRM(key *ecdsa.PrivateKey, k []byte, c []byte, r []byte, message []byte) []byte
//6. 发起者收集Si，生成多重签名：(s1 + s2 + ... + sn, R)
// --- func getSUsingAllSi(arrayOfSi [][]byte) []byte
// --- func GenerateMultiSignSignature(s []byte, r []byte) (*MultiSignature, error)
// GenerateMultiSignSignature生成对特定消息的多重签名，所有参与签名的私钥必须使用同一条椭圆曲线
//func GenerateMultiSignSignature(s []byte, r []byte) (*MultiSignature, error) {
func GenerateMultiSignSignature(s []byte, r []byte) ([]byte, error) {
	//	return marshalMultiSignature(s, r)

	// 生成多重签名：(sum(S), R)
	multiSig := &common.MultiSignature{
		S: s,
		R: r,
	}
	// 生成超级签名
	// 转换json
	sigContent, err := json.Marshal(multiSig)
	if err != nil {
		return nil, err
	}

	xuperSig := &common.XuperSignature{
		SigType:    common.MultiSig,
		SigContent: sigContent,
	}

	//	log.Printf("xuperSig before marshal: %s", xuperSig)

	sig, err := json.Marshal(xuperSig)
	if err != nil {
		return nil, err
	}

	return sig, nil

}
