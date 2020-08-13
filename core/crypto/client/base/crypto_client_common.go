package base

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/xuperchain/xuperchain/core/crypto/config"

	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/multisign"
	"github.com/xuperchain/xuperchain/core/hdwallet/rand"
)

// CryptoClientCommon : common implementation for CryptoClient
type CryptoClientCommon struct {
}

// GenerateEntropy 产生随机熵
func (*CryptoClientCommon) GenerateEntropy(bitSize int) ([]byte, error) {
	entropyByte, err := rand.GenerateEntropy(bitSize)
	return entropyByte, err
}

// GenerateMnemonic 将随机熵转为助记词
func (*CryptoClientCommon) GenerateMnemonic(entropy []byte, language int) (string, error) {
	mnemonic, err := rand.GenerateMnemonic(entropy, language)
	return mnemonic, err
}

// GenerateSeedWithErrorChecking 将助记词转为指定长度的随机数种子，在此过程中，校验助记词是否合法
func (*CryptoClientCommon) GenerateSeedWithErrorChecking(mnemonic string, password string, keyLen int, language int) ([]byte, error) {
	seed, err := rand.GenerateSeedWithErrorChecking(mnemonic, password, keyLen, language)
	return seed, err
}

// GetEcdsaPrivateKeyJSONFormat 获取ECC私钥的json格式的表达
func (*CryptoClientCommon) GetEcdsaPrivateKeyJSONFormat(k *ecdsa.PrivateKey) (string, error) {
	jsonEcdsaPrivateKey, err := account.GetEcdsaPrivateKeyJSONFormat(k)
	return jsonEcdsaPrivateKey, err
}

// GetEcdsaPublicKeyJSONFormat 获取ECC公钥的json格式的表达
func (*CryptoClientCommon) GetEcdsaPublicKeyJSONFormat(k *ecdsa.PrivateKey) (string, error) {
	jsonEcdsaPublicKey, err := account.GetEcdsaPublicKeyJSONFormat(k)
	return jsonEcdsaPublicKey, err
}

// CryptoClientCommonMultiSig is the default implementation of Multisig interface
type CryptoClientCommonMultiSig struct{}

// --- 多重签名相关 start (keep those code here, will integrate to xchain later) ---

// GetRandom32Bytes 每个多重签名算法流程的参与节点生成32位长度的随机byte，返回值可以认为是k
func (ms *CryptoClientCommonMultiSig) GetRandom32Bytes() ([]byte, error) {
	return multisign.GetRandom32Bytes()
}

// GetRiUsingRandomBytes 每个多重签名算法流程的参与节点生成Ri = Ki*G
func (ms *CryptoClientCommonMultiSig) GetRiUsingRandomBytes(key *ecdsa.PublicKey, k []byte) []byte {
	return multisign.GetRiUsingRandomBytes(key, k)
}

// GetRUsingAllRi 负责计算多重签名的节点来收集所有节点的Ri，并计算R = k1*G + k2*G + ... + kn*G
func (ms *CryptoClientCommonMultiSig) GetRUsingAllRi(key *ecdsa.PublicKey, arrayOfRi [][]byte) []byte {
	return multisign.GetRUsingAllRi(key, arrayOfRi)
}

// GetSharedPublicKeyForPublicKeys 负责计算多重签名的节点来收集所有节点的公钥Pi，并计算公共公钥：C = P1 + P2 + ... + Pn
func (ms *CryptoClientCommonMultiSig) GetSharedPublicKeyForPublicKeys(keys []*ecdsa.PublicKey) ([]byte, error) {
	return multisign.GetSharedPublicKeyForPublicKeys(keys)
}

// GetSiUsingKCRM 负责计算多重签名的节点将计算出的R和C分别传递给各个参与节点后，由各个参与节点再次计算自己的Si
// 计算 Si = Ki + HASH(C,R,m) * Xi
// X代表大数D，也就是私钥的关键参数
func (ms *CryptoClientCommonMultiSig) GetSiUsingKCRM(key *ecdsa.PrivateKey, k []byte, c []byte, r []byte, message []byte) []byte {
	return multisign.GetSiUsingKCRM(key, k, c, r, message)
}

// GetSUsingAllSi 负责计算多重签名的节点来收集所有节点的Si，并计算出S = sum(si)
func (ms *CryptoClientCommonMultiSig) GetSUsingAllSi(arrayOfSi [][]byte) []byte {
	return multisign.GetSUsingAllSi(arrayOfSi)
}

// GenerateMultiSignSignature 负责计算多重签名的节点，最终生成多重签名的统一签名格式
func (ms *CryptoClientCommonMultiSig) GenerateMultiSignSignature(s []byte, r []byte) ([]byte, error) {
	return multisign.GenerateMultiSignSignature(s, r)
}

// VerifyMultiSig 使用ECC公钥数组来进行多重签名的验证
func (ms *CryptoClientCommonMultiSig) VerifyMultiSig(keys []*ecdsa.PublicKey, signature, message []byte) (bool, error) {
	// 判断是否是nist标准的私钥
	if len(keys) < 2 {
		return false, fmt.Errorf("The total num of keys should be greater than two")
	}

	switch keys[0].Params().Name {
	case config.CurveNist: // NIST
		signature, err := multisign.VerifyMultiSig(keys, signature, message)
		return signature, err
	case config.CurveNistSN: // NIST + schnorr
		signature, err := multisign.VerifyMultiSig(keys, signature, message)
		return signature, err
	case config.CurveGm: // 国密
		return false, fmt.Errorf("This cryptography has not been supported yet")
	default: // 不支持的密码学类型
		return false, fmt.Errorf("This cryptography has not been supported yet")
	}
}

// MultiSign  多重签名的另一种用法，适用于完全中心化的流程
// 使用ECC私钥数组来进行多重签名，生成统一签名格式
func (ms *CryptoClientCommonMultiSig) MultiSign(keys []*ecdsa.PrivateKey, message []byte) ([]byte, error) {
	// 判断是否是nist标准的私钥
	if len(keys) < 2 {
		return nil, fmt.Errorf("The total num of keys should be greater than two")
	}

	switch keys[0].Params().Name {
	case config.CurveNist: // NIST
		signature, err := multisign.MultiSign(keys, message)
		return signature, err
	case config.CurveNistSN: // NIST + schnorr
		signature, err := multisign.MultiSign(keys, message)
		return signature, err
	case config.CurveGm: // 国密
		return nil, fmt.Errorf("This cryptography has not been supported yet")
	default: // 不支持的密码学类型
		return nil, fmt.Errorf("This cryptography has not been supported yet")
	}
}

// --- 多重签名相关 end ---
