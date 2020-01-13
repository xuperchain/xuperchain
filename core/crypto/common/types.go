package common

import (
	"math/big"
)

// 定义签名中所包含的标记符的值，及其所对应的签名算法的类型
const (
	// ECDSA签名算法
	ECDSA = "ECDSA"
	// Schnorr签名算法，EDDSA的前身
	Schnorr = "Schnorr"
	// Schnorr环签名算法
	SchnorrRing = "SchnorrRing"
	// 多重签名算法
	MultiSig = "MultiSig"
)

// --- 签名数据结构相关 start ---

// XuperSignature 统一的签名结构
type XuperSignature struct {
	SigType    string
	SigContent []byte
}

// ECDSASignature ECDSA签名
type ECDSASignature struct {
	R, S *big.Int
}

// SchnorrSignature Schnorr签名，EDDSA的前身
type SchnorrSignature struct {
	E, S *big.Int
}

// --- Schnorr环签名的数据结构定义 start ---

// PublicKeyFactor 公钥元素
type PublicKeyFactor struct {
	X, Y *big.Int
}

// RingSignature Schnorr环签名
type RingSignature struct {
	//	elliptic.Curve
	CurveName string
	Members   []*PublicKeyFactor
	E         *big.Int
	S         []*big.Int
}

// --- Schnorr环签名的数据结构定义 end ---

// MultiSignature 多重签名
type MultiSignature struct {
	S []byte
	R []byte
}

// MultiSigCommon 多重签名中间公共结果，C是公共公钥，R是公共随机数
type MultiSigCommon struct {
	C []byte
	R []byte
}

// --- 签名数据结构相关 end ---
