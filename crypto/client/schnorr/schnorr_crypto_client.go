// Package main is the plugin for xuperchain crypto client with schnorr sign
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/xuperchain/xuperunion/crypto/account"
	"github.com/xuperchain/xuperunion/crypto/client/base"
	schnorr_ring_sign "github.com/xuperchain/xuperunion/crypto/client/schnorr/ringsign"
	schnorr_sign "github.com/xuperchain/xuperunion/crypto/client/schnorr/sign"
	"github.com/xuperchain/xuperunion/crypto/client/schnorr/verify"
	"github.com/xuperchain/xuperunion/crypto/config"
	"github.com/xuperchain/xuperunion/crypto/ecies"
	"github.com/xuperchain/xuperunion/crypto/multisign"
	"github.com/xuperchain/xuperunion/crypto/utils"
	"github.com/xuperchain/xuperunion/hdwallet/key"
	walletRand "github.com/xuperchain/xuperunion/hdwallet/rand"
)

// SchnorrCryptoClient is the implementation for xchain default crypto
type SchnorrCryptoClient struct {
	base.CryptoClientCommon
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

// --- 多重签名相关 start (keep those code here, will integrate to xchain later) ---

// GetRandom32Bytes 每个多重签名算法流程的参与节点生成32位长度的随机byte，返回值可以认为是k
func (xcc SchnorrCryptoClient) GetRandom32Bytes() ([]byte, error) {
	return multisign.GetRandom32Bytes()
}

// GetRiUsingRandomBytes 每个多重签名算法流程的参与节点生成Ri = Ki*G
func (xcc SchnorrCryptoClient) GetRiUsingRandomBytes(key *ecdsa.PublicKey, k []byte) []byte {
	return multisign.GetRiUsingRandomBytes(key, k)
}

// GetRUsingAllRi 负责计算多重签名的节点来收集所有节点的Ri，并计算R = k1*G + k2*G + ... + kn*G
func (xcc SchnorrCryptoClient) GetRUsingAllRi(key *ecdsa.PublicKey, arrayOfRi [][]byte) []byte {
	return multisign.GetRUsingAllRi(key, arrayOfRi)
}

// GetSharedPublicKeyForPublicKeys 负责计算多重签名的节点来收集所有节点的公钥Pi，并计算公共公钥：C = P1 + P2 + ... + Pn
func (xcc SchnorrCryptoClient) GetSharedPublicKeyForPublicKeys(keys []*ecdsa.PublicKey) ([]byte, error) {
	return multisign.GetSharedPublicKeyForPublicKeys(keys)
}

// GetSiUsingKCRM 负责计算多重签名的节点将计算出的R和C分别传递给各个参与节点后，由各个参与节点再次计算自己的Si
// 计算 Si = Ki + HASH(C,R,m) * Xi
// X代表大数D，也就是私钥的关键参数
func (xcc SchnorrCryptoClient) GetSiUsingKCRM(key *ecdsa.PrivateKey, k []byte, c []byte, r []byte, message []byte) []byte {
	return multisign.GetSiUsingKCRM(key, k, c, r, message)
}

// GetSUsingAllSi 负责计算多重签名的节点来收集所有节点的Si，并计算出S = sum(si)
func (xcc SchnorrCryptoClient) GetSUsingAllSi(arrayOfSi [][]byte) []byte {
	return multisign.GetSUsingAllSi(arrayOfSi)
}

// GenerateMultiSignSignature 负责计算多重签名的节点，最终生成多重签名的统一签名格式
func (xcc SchnorrCryptoClient) GenerateMultiSignSignature(s []byte, r []byte) ([]byte, error) {
	return multisign.GenerateMultiSignSignature(s, r)
}

// VerifyMultiSig 使用ECC公钥数组来进行多重签名的验证
func (xcc SchnorrCryptoClient) VerifyMultiSig(keys []*ecdsa.PublicKey, signature, message []byte) (bool, error) {
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
func (xcc SchnorrCryptoClient) MultiSign(keys []*ecdsa.PrivateKey, message []byte) ([]byte, error) {
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

// VerifyXuperSignature 统一验签算法
func (xcc SchnorrCryptoClient) VerifyXuperSignature(publicKeys []*ecdsa.PublicKey, sig []byte, message []byte) (valid bool, err error) {
	return verify.XuperSigVerify(publicKeys, sig, message)
}
