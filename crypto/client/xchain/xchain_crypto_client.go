// Package main is the plugin for xuperchain default crypto client
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/xuperchain/xuperunion/crypto/common"

	"github.com/xuperchain/xuperunion/crypto/account"
	"github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/crypto/config"
	"github.com/xuperchain/xuperunion/crypto/ecies"
	"github.com/xuperchain/xuperunion/crypto/sign"
	"github.com/xuperchain/xuperunion/crypto/utils"
	"github.com/xuperchain/xuperunion/hdwallet/key"
	walletRand "github.com/xuperchain/xuperunion/hdwallet/rand"
)

// make sure this plugin implemented the interface
var _ base.CryptoClient = (*XchainCryptoClient)(nil)

// XchainCryptoClient is the implementation for xchain default crypto
type XchainCryptoClient struct {
	base.CryptoClientCommon
	base.CryptoClientCommonMultiSig
}

// GetInstance returns the an instance of XchainCryptoClient
func GetInstance() interface{} {
	return &XchainCryptoClient{}
}

// GenerateKeyBySeed 通过随机数种子来生成椭圆曲线加密所需要的公钥和私钥
func (xcc XchainCryptoClient) GenerateKeyBySeed(seed []byte) (*ecdsa.PrivateKey, error) {
	curve := elliptic.P256()
	privateKey, err := utils.GenerateKeyBySeed(curve, seed)
	return privateKey, err
}

// SignECDSA 使用ECC私钥来签名
func (xcc XchainCryptoClient) SignECDSA(k *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	signature, err := sign.SignECDSA(k, msg)
	return signature, err
}

// VerifyECDSA 使用ECC公钥来验证签名
func (xcc XchainCryptoClient) VerifyECDSA(k *ecdsa.PublicKey, signature, msg []byte) (bool, error) {
	result, err := sign.VerifyECDSA(k, signature, msg)
	return result, err
}

// ExportNewAccount 创建新账户(不使用助记词，不推荐使用)
func (xcc XchainCryptoClient) ExportNewAccount(path string) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	return account.ExportNewAccount(path, privateKey)
}

// CreateNewAccountWithMnemonic 创建含有助记词的新的账户
// 返回的字段：（助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (xcc XchainCryptoClient) CreateNewAccountWithMnemonic(language int, strength uint8) (*account.ECDSAAccount, error) {
	ecdsaAccount, err := account.CreateNewAccountWithMnemonic(language, strength, config.Nist)
	return ecdsaAccount, err
}

// CreateNewAccountAndSaveSecretKey 创建新的账户，并用支付密码加密私钥后存在本地，
// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (xcc XchainCryptoClient) CreateNewAccountAndSaveSecretKey(path string, language int, strength uint8, password string) (*account.ECDSAInfo, error) {
	ecdasaInfo, err := key.CreateAndSaveSecretKey(path, walletRand.SimplifiedChinese, account.StrengthHard, password, config.Nist)
	return ecdasaInfo, err
}

// ExportNewAccountWithMnemonic 创建新的账户，并导出相关文件（含助记词）到本地。
// 生成如下几个文件：1.助记词，2.私钥，3.公钥，4.钱包地址
func (xcc XchainCryptoClient) ExportNewAccountWithMnemonic(path string, language int, strength uint8) error {
	err := account.ExportNewAccountWithMnemonic(path, language, strength, config.Nist)
	return err
}

// RetrieveAccountByMnemonic 从助记词恢复钱包账户
// TODO: 后续可以从助记词中识别出语言类型
func (xcc XchainCryptoClient) RetrieveAccountByMnemonic(mnemonic string, language int) (*account.ECDSAAccount, error) {
	ecdsaAccount, err := account.GenerateAccountByMnemonic(mnemonic, language)
	return ecdsaAccount, err
}

// RetrieveAccountByMnemonicAndSavePrivKey 从助记词恢复钱包账户，并用支付密码加密私钥后存在本地，
// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (xcc XchainCryptoClient) RetrieveAccountByMnemonicAndSavePrivKey(path string, language int, mnemonic string, password string) (*account.ECDSAInfo, error) {
	ecdsaAccount, err := key.CreateAndSaveSecretKeyWithMnemonic(path, language, mnemonic, password)
	return ecdsaAccount, err
}

// EncryptAccount 使用支付密码加密账户信息并返回加密后的数据（后续用来回传至云端）
func (xcc XchainCryptoClient) EncryptAccount(info *account.ECDSAAccount, password string) (*account.ECDSAAccountToCloud, error) {
	ecdsaAccountToCloud, err := key.EncryptAccount(info, password)
	return ecdsaAccountToCloud, err
}

// GetBinaryEcdsaPrivateKeyFromFile 从导出的私钥文件读取私钥的byte格式
func (xcc XchainCryptoClient) GetBinaryEcdsaPrivateKeyFromFile(path string, password string) ([]byte, error) {
	binaryEcdsaPrivateKey, err := key.GetBinaryEcdsaPrivateKeyFromFile(path, password)
	return binaryEcdsaPrivateKey, err
}

// GetEcdsaPrivateKeyFromFileByPassword 使用支付密码从导出的私钥文件读取私钥
func (xcc XchainCryptoClient) GetEcdsaPrivateKeyFromFileByPassword(path string, password string) (*ecdsa.PrivateKey, error) {
	ecdsaPrivateKey, err := key.GetEcdsaPrivateKeyFromFile(path, password)
	return ecdsaPrivateKey, err
}

// GetBinaryEcdsaPrivateKeyFromString 从二进制加密字符串获取真实私钥的byte格式
func (xcc XchainCryptoClient) GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey string, password string) ([]byte, error) {
	binaryEcdsaPrivateKey, err := key.GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey, password)
	return binaryEcdsaPrivateKey, err
}

// GetEcdsaPrivateKeyFromFile 从导出的私钥文件读取私钥
func (xcc XchainCryptoClient) GetEcdsaPrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	ecdsaPrivateKey, err := account.GetEcdsaPrivateKeyFromFile(filename)
	return ecdsaPrivateKey, err
}

// GetEcdsaPublicKeyFromFile 从导出的公钥文件读取公钥
func (xcc XchainCryptoClient) GetEcdsaPublicKeyFromFile(filename string) (*ecdsa.PublicKey, error) {
	ecdsaPublicKey, err := account.GetEcdsaPublicKeyFromFile(filename)
	return ecdsaPublicKey, err
}

// Encrypt 使用ECIES加密
func (xcc XchainCryptoClient) Encrypt(publicKey *ecdsa.PublicKey, msg []byte) (cypherText []byte, err error) {
	cypherText, err = ecies.Encrypt(publicKey, msg)
	return cypherText, err
}

// Decrypt 使用ECIES解密
func (xcc XchainCryptoClient) Decrypt(privateKey *ecdsa.PrivateKey, cypherText []byte) (msg []byte, err error) {
	msg, err = ecies.Decrypt(privateKey, cypherText)
	return msg, err
}

// GetEcdsaPrivateKeyFromJSON 从导出的私钥文件读取私钥
func (xcc XchainCryptoClient) GetEcdsaPrivateKeyFromJSON(jsonBytes []byte) (*ecdsa.PrivateKey, error) {
	return account.GetEcdsaPrivateKeyFromJSON(jsonBytes)
}

// GetEcdsaPublicKeyFromJSON 从导出的公钥文件读取公钥
func (xcc XchainCryptoClient) GetEcdsaPublicKeyFromJSON(jsonBytes []byte) (*ecdsa.PublicKey, error) {
	return account.GetEcdsaPublicKeyFromJSON(jsonBytes)
}

// XuperVerify 统一验签算法, common signature verification function
func (xcc XchainCryptoClient) XuperVerify(keys []*ecdsa.PublicKey, sig []byte, message []byte) (bool, error) {
	if len(keys) < 1 {
		return false, fmt.Errorf("no public key found")
	}
	curveName := keys[0].Params().Name
	xuperSig := new(common.XuperSignature)
	err := json.Unmarshal(sig, xuperSig)
	if err != nil {
		return false, err
	}

	// 说明不是统一超级签名的格式
	if err != nil {
		switch keys[0].Params().Name {
		case config.CurveNist: // NIST
			verifyResult, err := sign.VerifyECDSA(keys[0], sig, message)
			return verifyResult, err
		default: // 不支持的密码学类型
			return false, fmt.Errorf("This cryptography[%v] has not been supported yet", keys[0].Params().Name)
		}
	}

	switch xuperSig.SigType {
	// ECDSA签名
	case common.ECDSA:
		if curveName == config.CurveNist {
			verifyResult, err := sign.VerifyECDSA(keys[0], xuperSig.SigContent, message)
			return verifyResult, err
		}
		return false, fmt.Errorf("This cryptography[%v] has not been supported yet", curveName)

	// 多重签名
	case common.MultiSig:
		if curveName == config.CurveNist {
			verifyResult, err := xcc.VerifyMultiSig(keys, xuperSig.SigContent, message)
			return verifyResult, err
		}
		return false, fmt.Errorf("This cryptography[%v] has not been supported yet", curveName)

	// 不支持的签名类型
	default:
		err = fmt.Errorf("This XuperSignature type[%v] is not supported in this version", xuperSig.SigType)
		return false, err
	}
}
