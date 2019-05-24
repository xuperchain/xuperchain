package base

import (
	"crypto/ecdsa"

	"github.com/xuperchain/xuperunion/crypto/account"
	"github.com/xuperchain/xuperunion/hdwallet/rand"
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

// GetAddressFromPublicKey 通过公钥来计算地址
func (*CryptoClientCommon) GetAddressFromPublicKey(pub *ecdsa.PublicKey) (string, error) {
	address, err := account.GetAddressFromPublicKey(pub)
	return address, err
}

// CheckAddressFormat 验证钱包地址是否是合法的格式。
// 如果成功，返回true和对应的版本号；如果失败，返回false和默认的版本号0
func (*CryptoClientCommon) CheckAddressFormat(address string) (bool, uint8) {
	isValid, nVersion := account.CheckAddressFormat(address)
	return isValid, nVersion
}

// VerifyAddressUsingPublicKey 验证钱包地址是否和指定的公钥match。
// 如果成功，返回true和对应的版本号；如果失败，返回false和默认的版本号0
func (*CryptoClientCommon) VerifyAddressUsingPublicKey(address string, pub *ecdsa.PublicKey) (bool, uint8) {
	isValid, nVersion := account.VerifyAddressUsingPublicKey(address, pub)
	return isValid, nVersion
}
