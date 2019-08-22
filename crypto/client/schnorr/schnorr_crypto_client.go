// Package main is the plugin for xuperchain crypto client with schnorr sign
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/xuperchain/xuperunion/crypto/account"
	"github.com/xuperchain/xuperunion/crypto/client/base"
	schnorr_ring_sign "github.com/xuperchain/xuperunion/crypto/client/schnorr/ringsign"
	schnorr_sign "github.com/xuperchain/xuperunion/crypto/client/schnorr/sign"
	"github.com/xuperchain/xuperunion/crypto/client/schnorr/verify"
	"github.com/xuperchain/xuperunion/crypto/config"
	"github.com/xuperchain/xuperunion/crypto/ecies"
	"github.com/xuperchain/xuperunion/crypto/utils"
	"github.com/xuperchain/xuperunion/hdwallet/key"
	walletRand "github.com/xuperchain/xuperunion/hdwallet/rand"
)

// make sure this plugin implemented the interface
var _ base.CryptoClient = (*SchnorrCryptoClient)(nil)

// SchnorrCryptoClient is the implementation for xchain default crypto
type SchnorrCryptoClient struct {
	base.CryptoClientCommon
	base.CryptoClientCommonMultiSig
}

// GetInstance returns the an instance of SchnorrCryptoClient
func GetInstance() interface{} {
	return &SchnorrCryptoClient{}
}

// GenerateKeyBySeed 通过随机数种子来生成椭圆曲线加密所需要的公钥和私钥
func (xcc SchnorrCryptoClient) GenerateKeyBySeed(seed []byte) (*ecdsa.PrivateKey, error) {
	curve := elliptic.P256()
	curve.Params().Name = "P-256-SN"
	privateKey, err := utils.GenerateKeyBySeed(curve, seed)
	return privateKey, err
}

// SignECDSA 使用ECC私钥来签名
func (xcc SchnorrCryptoClient) SignECDSA(k *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	signature, err := schnorr_sign.Sign(k, msg)
	return signature, err
}

// VerifyECDSA 使用ECC公钥来验证签名
func (xcc SchnorrCryptoClient) VerifyECDSA(k *ecdsa.PublicKey, signature, msg []byte) (bool, error) {
	result, err := schnorr_sign.Verify(k, signature, msg)
	return result, err
}

// ExportNewAccount 创建新账户(不使用助记词，不推荐使用)
func (xcc SchnorrCryptoClient) ExportNewAccount(path string) error {
	curve := elliptic.P256()
	curve.Params().Name = "P-256-SN"
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}
	return account.ExportNewAccount(path, privateKey)
}

// CreateNewAccountWithMnemonic 创建含有助记词的新的账户
// 返回的字段：（助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (xcc SchnorrCryptoClient) CreateNewAccountWithMnemonic(language int, strength uint8) (*account.ECDSAAccount, error) {
	ecdsaAccount, err := account.CreateNewAccountWithMnemonic(language, strength, config.NistSN)
	return ecdsaAccount, err
}

// CreateNewAccountAndSaveSecretKey 创建新的账户，并用支付密码加密私钥后存在本地，
// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (xcc SchnorrCryptoClient) CreateNewAccountAndSaveSecretKey(path string, language int, strength uint8, password string) (*account.ECDSAInfo, error) {
	ecdasaInfo, err := key.CreateAndSaveSecretKey(path, walletRand.SimplifiedChinese, account.StrengthHard, password, config.NistSN)
	return ecdasaInfo, err
}

// ExportNewAccountWithMnemonic 创建新的账户，并导出相关文件（含助记词）到本地。
// 生成如下几个文件：1.助记词，2.私钥，3.公钥，4.钱包地址
func (xcc SchnorrCryptoClient) ExportNewAccountWithMnemonic(path string, language int, strength uint8) error {
	err := account.ExportNewAccountWithMnemonic(path, language, strength, config.NistSN)
	return err
}

// RetrieveAccountByMnemonic 从助记词恢复钱包账户
// TODO: 后续可以从助记词中识别出语言类型
func (xcc SchnorrCryptoClient) RetrieveAccountByMnemonic(mnemonic string, language int) (*account.ECDSAAccount, error) {
	ecdsaAccount, err := account.GenerateAccountByMnemonic(mnemonic, language)
	return ecdsaAccount, err
}

// RetrieveAccountByMnemonicAndSavePrivKey 从助记词恢复钱包账户，并用支付密码加密私钥后存在本地，
// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (xcc SchnorrCryptoClient) RetrieveAccountByMnemonicAndSavePrivKey(path string, language int, mnemonic string, password string) (*account.ECDSAInfo, error) {
	ecdsaAccount, err := key.CreateAndSaveSecretKeyWithMnemonic(path, language, mnemonic, password)
	return ecdsaAccount, err
}

// EncryptAccount 使用支付密码加密账户信息并返回加密后的数据（后续用来回传至云端）
func (xcc SchnorrCryptoClient) EncryptAccount(info *account.ECDSAAccount, password string) (*account.ECDSAAccountToCloud, error) {
	ecdsaAccountToCloud, err := key.EncryptAccount(info, password)
	return ecdsaAccountToCloud, err
}

// GetBinaryEcdsaPrivateKeyFromFile 从导出的私钥文件读取私钥的byte格式
func (xcc SchnorrCryptoClient) GetBinaryEcdsaPrivateKeyFromFile(path string, password string) ([]byte, error) {
	binaryEcdsaPrivateKey, err := key.GetBinaryEcdsaPrivateKeyFromFile(path, password)
	return binaryEcdsaPrivateKey, err
}

// GetEcdsaPrivateKeyFromFileByPassword 使用支付密码从导出的私钥文件读取私钥
func (xcc SchnorrCryptoClient) GetEcdsaPrivateKeyFromFileByPassword(path string, password string) (*ecdsa.PrivateKey, error) {
	ecdsaPrivateKey, err := key.GetEcdsaPrivateKeyFromFile(path, password)
	return ecdsaPrivateKey, err
}

// GetBinaryEcdsaPrivateKeyFromString 从二进制加密字符串获取真实私钥的byte格式
func (xcc SchnorrCryptoClient) GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey string, password string) ([]byte, error) {
	binaryEcdsaPrivateKey, err := key.GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey, password)
	return binaryEcdsaPrivateKey, err
}

// GetEcdsaPrivateKeyFromFile 从导出的私钥文件读取私钥
func (xcc SchnorrCryptoClient) GetEcdsaPrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	ecdsaPrivateKey, err := account.GetEcdsaPrivateKeyFromFile(filename)
	return ecdsaPrivateKey, err
}

// GetEcdsaPublicKeyFromFile 从导出的公钥文件读取公钥
func (xcc SchnorrCryptoClient) GetEcdsaPublicKeyFromFile(filename string) (*ecdsa.PublicKey, error) {
	ecdsaPublicKey, err := account.GetEcdsaPublicKeyFromFile(filename)
	return ecdsaPublicKey, err
}

// Encrypt 使用ECIES加密
func (xcc SchnorrCryptoClient) Encrypt(publicKey *ecdsa.PublicKey, msg []byte) (cypherText []byte, err error) {
	cypherText, err = ecies.Encrypt(publicKey, msg)
	return cypherText, err
}

// Decrypt 使用ECIES解密
func (xcc SchnorrCryptoClient) Decrypt(privateKey *ecdsa.PrivateKey, cypherText []byte) (msg []byte, err error) {
	msg, err = ecies.Decrypt(privateKey, cypherText)
	return msg, err
}

// GetEcdsaPrivateKeyFromJSON 从导出的私钥文件读取私钥
func (xcc SchnorrCryptoClient) GetEcdsaPrivateKeyFromJSON(jsonBytes []byte) (*ecdsa.PrivateKey, error) {
	return account.GetEcdsaPrivateKeyFromJSON(jsonBytes)
}

// GetEcdsaPublicKeyFromJSON 从导出的公钥文件读取公钥
func (xcc SchnorrCryptoClient) GetEcdsaPublicKeyFromJSON(jsonBytes []byte) (*ecdsa.PublicKey, error) {
	return account.GetEcdsaPublicKeyFromJSON(jsonBytes)
}

// --- 	schnorr 环签名算法相关 start ---

// SignSchnorrRing schnorr环签名算法 生成统一签名
func (xcc SchnorrCryptoClient) SignSchnorrRing(keys []*ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, message []byte) (ringSignature []byte, err error) {
	return schnorr_ring_sign.Sign(keys, privateKey, message)
}

// VerifySchnorrRing schnorr环签名算法 验证签名
func (xcc SchnorrCryptoClient) VerifySchnorrRing(keys []*ecdsa.PublicKey, sig, message []byte) (bool, error) {
	return schnorr_ring_sign.Verify(keys, sig, message)
}

// --- 	schnorr 环签名算法相关 end ---

// XuperVerify 统一验签算法
func (xcc SchnorrCryptoClient) XuperVerify(publicKeys []*ecdsa.PublicKey, sig []byte, message []byte) (valid bool, err error) {
	return verify.XuperSigVerify(publicKeys, sig, message)
}
