/*
Copyright Baidu Inc. All Rights Reserved.
*/

package base

import (
	"crypto/ecdsa"

	"github.com/xuperchain/xuperunion/crypto/account"
)

// CryptoClient is the interface of all Crypto functions
type CryptoClient interface {

	// 通过随机数种子生成ECC私钥
	GenerateKeyBySeed(seed []byte) (*ecdsa.PrivateKey, error)

	// 使用ECC私钥来签名
	SignECDSA(k *ecdsa.PrivateKey, msg []byte) (signature []byte, err error)

	// 使用ECC公钥来验证签名
	VerifyECDSA(k *ecdsa.PublicKey, signature, msg []byte) (valid bool, err error)

	// schnorr签名算法 生成统一签名
	SignSchnorr(k *ecdsa.PrivateKey, msg []byte) (signature []byte, err error)

	// schnorr签名算法 验证签名
	VerifySchnorr(k *ecdsa.PublicKey, signature, msg []byte) (valid bool, err error)

	// 统一验签算法
	VerifyXuperSignature(publicKeys []*ecdsa.PublicKey, sig []byte, message []byte) (valid bool, err error)

	// 创建新的账户，不需要助记词。生成如下几个文件：1.私钥，2.公钥，3.钱包地址 (不建议使用)
	ExportNewAccount(path string) error

	// 创建含有助记词的新的账户，返回的字段：（助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
	CreateNewAccountWithMnemonic(language int, strength uint8) (*account.ECDSAAccount, error)

	// 创建新的账户，并用支付密码加密私钥后存在本地，
	// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
	//CreateAndSaveSecretKey(path string, nVersion uint8, language int, strength uint8, password string) (*account.ECDSAInfo, error)

	// 创建新的账户，并导出相关文件（含助记词）到本地。生成如下几个文件：1.助记词，2.私钥，3.公钥，4.钱包地址
	ExportNewAccountWithMnemonic(path string, language int, strength uint8) error

	// 从助记词恢复钱包账户
	RetrieveAccountByMnemonic(mnemonic string, language int) (*account.ECDSAAccount, error)

	// 从助记词恢复钱包账户，并用支付密码加密私钥后存在本地，
	// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
	RetrieveAccountByMnemonicAndSavePrivKey(path string, language int, mnemonic string, password string) (*account.ECDSAInfo, error)

	// 使用支付密码加密账户信息并返回加密后的数据（后续用来回传至云端）
	EncryptAccount(info *account.ECDSAAccount, password string) (*account.ECDSAAccountToCloud, error)

	// 从导出的私钥文件读取私钥的byte格式
	GetBinaryEcdsaPrivateKeyFromFile(path string, password string) ([]byte, error)

	// 从导出的私钥文件读取私钥
	GetEcdsaPrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error)

	// 使用支付密码从导出的私钥文件读取私钥
	GetEcdsaPrivateKeyFromFileByPassword(path string, password string) (*ecdsa.PrivateKey, error)

	// 从二进制加密字符串获取真实私钥的byte格式
	GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey string, password string) ([]byte, error)

	// 从导出的公钥文件读取公钥
	GetEcdsaPublicKeyFromFile(filename string) (*ecdsa.PublicKey, error)

	// 使用ECIES加密
	Encrypt(publicKey *ecdsa.PublicKey, msg []byte) (cypherText []byte, err error)

	// 使用ECIES解密
	Decrypt(privateKey *ecdsa.PrivateKey, cypherText []byte) (msg []byte, err error)

	// 产生随机熵
	GenerateEntropy(bitSize int) ([]byte, error)

	// 将随机熵转为助记词
	GenerateMnemonic(entropy []byte, language int) (string, error)

	// 将助记词转为指定长度的随机数种子，在此过程中，校验助记词是否合法
	GenerateSeedWithErrorChecking(mnemonic string, password string, keyLen int, language int) ([]byte, error)

	// 从导出的私钥文件读取私钥
	GetEcdsaPrivateKeyFromJSON(jsonBytes []byte) (*ecdsa.PrivateKey, error)

	// 从导出的公钥文件读取公钥
	GetEcdsaPublicKeyFromJSON(jsonBytes []byte) (*ecdsa.PublicKey, error)

	// 获取ECC私钥的json格式的表达
	GetEcdsaPrivateKeyJSONFormat(k *ecdsa.PrivateKey) (string, error)

	// 获取ECC公钥的json格式的表达
	GetEcdsaPublicKeyJSONFormat(k *ecdsa.PrivateKey) (string, error)

	// 通过公钥来计算地址
	GetAddressFromPublicKey(pub *ecdsa.PublicKey) (string, error)

	// 验证钱包地址是否是合法的格式。如果成功，返回true和对应的版本号；如果失败，返回false和默认的版本号0
	CheckAddressFormat(address string) (bool, uint8)

	// 验证钱包地址是否和指定的公钥match。如果成功，返回true和对应的版本号；如果失败，返回false和默认的版本号0
	VerifyAddressUsingPublicKey(address string, pub *ecdsa.PublicKey) (bool, uint8)

	// --- 多重签名相关 start ---

	// 每个多重签名算法流程的参与节点生成32位长度的随机byte，返回值可以认为是k
	GetRandom32Bytes() ([]byte, error)

	// 每个多重签名算法流程的参与节点生成Ri = Ki*G
	GetRiUsingRandomBytes(key *ecdsa.PublicKey, k []byte) []byte

	// 负责计算多重签名的节点来收集所有节点的Ri，并计算R = k1*G + k2*G + ... + kn*G
	GetRUsingAllRi(key *ecdsa.PublicKey, arrayOfRi [][]byte) []byte

	// 负责计算多重签名的节点来收集所有节点的公钥Pi，并计算公共公钥：C = P1 + P2 + ... + Pn
	GetSharedPublicKeyForPublicKeys(keys []*ecdsa.PublicKey) ([]byte, error)

	// 负责计算多重签名的节点将计算出的R和C分别传递给各个参与节点后，由各个参与节点再次计算自己的Si
	// 计算 Si = Ki + HASH(C,R,m) * Xi
	// X代表大数D，也就是私钥的关键参数
	GetSiUsingKCRM(key *ecdsa.PrivateKey, k []byte, c []byte, r []byte, message []byte) []byte

	// 负责计算多重签名的节点来收集所有节点的Si，并计算出S = sum(si)
	GetSUsingAllSi(arrayOfSi [][]byte) []byte

	// 负责计算多重签名的节点，最终生成多重签名的统一签名格式
	//func (xcc XchainCryptoClient) GenerateMultiSignSignature(s []byte, r []byte) (*multisign.MultiSignature, error) {
	GenerateMultiSignSignature(s []byte, r []byte) ([]byte, error)

	// 使用ECC公钥数组来进行多重签名的验证
	//func (xcc XchainCryptoClient) VerifyMultiSig(keys []*ecdsa.PublicKey, signature *multisign.MultiSignature, message []byte) (bool, error) {
	VerifyMultiSig(keys []*ecdsa.PublicKey, signature, message []byte) (bool, error)

	// -- 多重签名的另一种用法，适用于完全中心化的流程
	// 使用ECC私钥数组来进行多重签名，生成统一签名格式
	//func (xcc XchainCryptoClient) MultiSign(keys []*ecdsa.PrivateKey, message []byte) (*multisign.MultiSignature, error) {
	MultiSign(keys []*ecdsa.PrivateKey, message []byte) ([]byte, error)

	// --- 多重签名相关 end ---

	// --- 	schnorr 环签名算法相关 start ---

	//func (xcc XchainCryptoClient) SignSchnorrRing(keys []*ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, message []byte) (*schnorr_ring_sign.RingSignature, error) {
	//	return schnorr_ring_sign.Sign(keys, privateKey, message)
	//}
	//
	//func (xcc XchainCryptoClient) VerifySchnorrRing(sig *schnorr_ring_sign.RingSignature, message []byte) bool {
	//	return schnorr_ring_sign.Verify(sig, message)
	//}

	// schnorr环签名算法 生成统一签名
	SignSchnorrRing(keys []*ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error)

	// schnorr环签名算法 验证签名
	VerifySchnorrRing(keys []*ecdsa.PublicKey, sig, message []byte) (bool, error)

	// --- 	schnorr 环签名算法相关 end ---

}
