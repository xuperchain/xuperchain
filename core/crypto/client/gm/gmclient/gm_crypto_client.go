/*
Copyright Baidu Inc. All Rights Reserved.
*/

package gmclient

import (
	//	"log"
	"crypto/ecdsa"
	"crypto/rand"
	//"encoding/json"
	"fmt"

	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/client/base"
	//"github.com/xuperchain/xuperchain/core/crypto/common"
	"github.com/xuperchain/xuperchain/core/crypto/config"
	"github.com/xuperchain/xuperchain/core/crypto/ecies"
	"github.com/xuperchain/xuperchain/core/crypto/sign"
	"github.com/xuperchain/xuperchain/core/crypto/utils"
	"github.com/xuperchain/xuperchain/core/hdwallet/key"
	walletRand "github.com/xuperchain/xuperchain/core/hdwallet/rand"

	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/multisign"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/schnorr_ring_sign"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/schnorr_sign"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/signature"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/sm2"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/sm4"
)

// make sure this plugin implemented the interface
var _ base.CryptoClient = (*GmCryptoClient)(nil)

type GmCryptoClient struct {
	base.CryptoClientCommon
}

// 通过随机数种子来生成椭圆曲线加密所需要的公钥和私钥
func (gmcc GmCryptoClient) GenerateKeyBySeed(seed []byte) (*ecdsa.PrivateKey, error) {
	curve := sm2.P256Sm2()
	privateKey, err := utils.GenerateKeyBySeed(curve, seed)
	return privateKey, err
}

//// 获取ECC私钥的json格式的表达
//func (gmcc GmCryptoClient) GetEcdsaPrivateKeyJsonFormat(k *ecdsa.PrivateKey) (string, error) {
//	jsonEcdsaPrivateKeyJsonFormat, err := account.GetEcdsaPrivateKeyJsonFormat(k)
//	return jsonEcdsaPrivateKeyJsonFormat, err
//}
//
//// 获取ECC公钥的json格式的表达
//func (gmcc GmCryptoClient) GetEcdsaPublicKeyJsonFormat(k *ecdsa.PrivateKey) (string, error) {
//	jsonEcdsaPublicKeyJsonFormat, err := account.GetEcdsaPublicKeyJsonFormat(k)
//	return jsonEcdsaPublicKeyJsonFormat, err
//}

// 使用ECC私钥来签名
func (gmcc GmCryptoClient) SignECDSA(k *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	// 判断是否是nist标准的私钥
	switch k.Params().Name {
	case config.CurveNist: // NIST
		signature, err := sign.SignECDSA(k, msg)
		return signature, err
	case config.CurveGm: // 国密
		signature, err := signECDSA(k, msg)
		return signature, err
	default: // 不支持的密码学类型
		return nil, fmt.Errorf("This cryptography has not been supported yet.")
	}

}

// // 使用ECC私钥来签名
// func (gmcc GmCryptoClient) SignV2ECDSA(k *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
// 	// 判断是否是nist标准的私钥
// 	switch k.Params().Name {
// 	case config.CurveNist: // NIST
// 		signature, err := sign.SignV2ECDSA(k, msg)
// 		return signature, err
// 	case config.CurveGm: // 国密
// 		signature, err := signV2ECDSA(k, msg)
// 		return signature, err
// 	default: // 不支持的密码学类型
// 		return nil, fmt.Errorf("This cryptography has not been supported yet.")
// 	}

// }

// 使用ECC私钥来签名
func signECDSA(k *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	if k.D == nil || k.X == nil || k.Y == nil {
		return nil, fmt.Errorf("invalid private key")
	}

	key := new(sm2.PrivateKey)
	//	key := &sm2.PrivateKey{}
	key.PublicKey.Curve = sm2.P256Sm2() // elliptic.P256()
	key.X = k.X
	key.Y = k.Y
	key.D = k.D

	r, s, err := sm2.Sign(key, msg)
	if err != nil {
		return nil, fmt.Errorf("Failed to sign the msg [%s]", err)
	}
	return utils.MarshalECDSASignature(r, s)
}

// // 使用ECC私钥来签名
// func signV2ECDSA(k *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
// 	if k.D == nil || k.X == nil || k.Y == nil {
// 		return nil, fmt.Errorf("invalid private key")
// 	}

// 	key := new(sm2.PrivateKey)
// 	//	key := &sm2.PrivateKey{}
// 	key.PublicKey.Curve = sm2.P256Sm2() // elliptic.P256()
// 	key.X = k.X
// 	key.Y = k.Y
// 	key.D = k.D

// 	r, s, err := sm2.Sign(key, msg)
// 	if err != nil {
// 		return nil, fmt.Errorf("Failed to sign the msg [%s]", err)
// 	}

// 	// 生成ECDSA签名：(sum(S), R)
// 	ecdsaSig := &common.ECDSASignature{
// 		R: r,
// 		S: s,
// 	}

// 	// 生成超级签名
// 	// 转换json
// 	sigContent, err := json.Marshal(ecdsaSig)
// 	if err != nil {
// 		return nil, err
// 	}

// 	xuperSig := &common.XuperSignature{
// 		SigType:    common.ECDSA,
// 		SigContent: sigContent,
// 	}

// 	//	log.Printf("xuperSig before marshal: %s", xuperSig)

// 	sig, err := json.Marshal(xuperSig)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return sig, nil
// }

// 使用ECC公钥来验证签名
func (gmcc GmCryptoClient) VerifyECDSA(k *ecdsa.PublicKey, signature, msg []byte) (bool, error) {
	// 判断是否是nist标准的私钥
	switch k.Params().Name {
	case config.CurveNist: // NIST
		result, err := sign.VerifyECDSA(k, signature, msg)
		return result, err
	case config.CurveGm: // 国密
		// TODO: 这块逻辑需要重构来提高代码复用
		result, err := verifyECDSA(k, signature, msg)
		return result, err
	default: // 不支持的密码学类型
		return false, fmt.Errorf("This cryptography has not been supported yet.")
	}
}

// // 使用ECC公钥来验证签名
// func (gmcc GmCryptoClient) VerifyV2ECDSA(k *ecdsa.PublicKey, signature, msg []byte) (bool, error) {
// 	// 判断是否是nist标准的私钥
// 	switch k.Params().Name {
// 	case config.CurveNist: // NIST
// 		result, err := sign.VerifyV2ECDSA(k, signature, msg)
// 		return result, err
// 	case config.CurveGm: // 国密
// 		// TODO: 这块逻辑需要重构来提高代码复用
// 		result, err := verifyV2ECDSA(k, signature, msg)
// 		return result, err
// 	default: // 不支持的密码学类型
// 		return false, fmt.Errorf("This cryptography has not been supported yet.")
// 	}
// }

// 使用ECC公钥来验证签名
func verifyECDSA(k *ecdsa.PublicKey, signature, msg []byte) (bool, error) {
	r, s, err := utils.UnmarshalECDSASignature(signature)
	if err != nil {
		return false, fmt.Errorf("Failed to unmarshal the signature [%s]", err)
	}

	key := new(sm2.PublicKey)
	key.Curve = sm2.P256Sm2() // elliptic.P256()
	key.X = k.X
	key.Y = k.Y

	return sm2.Verify(key, msg, r, s), nil
}

// // 使用ECC公钥来验证签名
// func verifyV2ECDSA(k *ecdsa.PublicKey, sig, msg []byte) (bool, error) {
// 	signature := new(common.ECDSASignature)
// 	err := json.Unmarshal(sig, signature)
// 	if err != nil {
// 		return false, fmt.Errorf("Failed to unmarshal the ecdsa signature [%s]", err)
// 	}

// 	key := new(sm2.PublicKey)
// 	key.Curve = sm2.P256Sm2() // elliptic.P256()
// 	key.X = k.X
// 	key.Y = k.Y

// 	return sm2.Verify(key, msg, signature.R, signature.S), nil
// }

// ExportNewAccount 创建新账户(不使用助记词，不推荐使用)
func (gmcc GmCryptoClient) ExportNewAccount(path string) error {
	lowLevelPrivateKey, err := ecdsa.GenerateKey(sm2.P256Sm2(), rand.Reader)
	if err != nil {
		return err
	}
	return account.ExportNewAccount(path, lowLevelPrivateKey)
}

//// 使用公钥来生成钱包地址
//func (gmcc GmCryptoClient) GetAddressFromPublicKey(nVersion uint8, pub *ecdsa.PublicKey) string {
//	address := account.GetAddressFromPublicKey(nVersion, pub)
//	return address
//}

// 创建含有助记词的新的账户，返回的字段：（助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (gmcc GmCryptoClient) CreateNewAccountWithMnemonic(language int, strength uint8) (*account.ECDSAAccount, error) {
	cryptography := uint8(config.Gm)
	//	ecdsaAccount, err := account.CreateNewAccountWithMnemonic(nVersion, language, strength, cryptography)
	ecdsaAccount, err := sm2.CreateNewAccountWithMnemonic(language, strength, cryptography)
	return ecdsaAccount, err
}

// 创建新的账户，并用支付密码加密私钥后存在本地，
// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (gmcc GmCryptoClient) CreateNewAccountAndSaveSecretKey(path string, language int, strength uint8, password string) (*account.ECDSAInfo, error) {
	cryptography := uint8(config.Gm)
	ecdasaInfo, err := key.CreateAndSaveSecretKey(path, walletRand.SimplifiedChinese, account.StrengthHard, password, cryptography)
	return ecdasaInfo, err
}

// 创建新的账户，并导出相关文件（含助记词）到本地。生成如下几个文件：1.助记词，2.私钥，3.公钥，4.钱包地址
func (gmcc GmCryptoClient) ExportNewAccountWithMnemonic(path string, language int, strength uint8) error {
	//	curve := sm2.P256Sm2()
	cryptography := uint8(config.Gm)
	//	err := account.ExportNewAccountWithMnemonic(path, nVersion, language, strength, cryptography)
	err := sm2.ExportNewAccountWithMnemonic(path, language, strength, cryptography)
	return err
}

// 从助记词恢复钱包账户
func (gmcc GmCryptoClient) RetrieveAccountByMnemonic(mnemonic string, language int) (*account.ECDSAAccount, error) {
	//	ecdsaAccount, err := sm2.GenerateAccountByMnemonic(mnemonic, language)
	ecdsaAccount, err := sm2.RetrieveAccountByMnemonic(mnemonic, language)
	return ecdsaAccount, err
}

// 从助记词恢复钱包账户，并用支付密码加密私钥后存在本地，
// 返回的字段：（随机熵（供其他钱包软件推导出私钥）、助记词、私钥的json、公钥的json、钱包地址） as ECDSAAccount，以及可能的错误信息
func (gmcc GmCryptoClient) RetrieveAccountByMnemonicAndSavePrivKey(path string, language int, mnemonic string, password string) (*account.ECDSAInfo, error) {
	//	curve := sm2.P256Sm2()
	ecdsaAccount, err := key.CreateAndSaveSecretKeyWithMnemonic(path, language, mnemonic, password)
	return ecdsaAccount, err
}

// 使用支付密码加密账户信息并返回加密后的数据（后续用来回传至云端）
func (gmcc GmCryptoClient) EncryptAccount(info *account.ECDSAAccount, password string) (*account.ECDSAAccountToCloud, error) {
	ecdsaAccountToCloud, err := sm4.EncryptAccount(info, password)
	return ecdsaAccountToCloud, err
}

// 从导出的私钥文件读取私钥的byte格式
func (gmcc GmCryptoClient) GetBinaryEcdsaPrivateKeyFromFile(path string, password string) ([]byte, error) {
	binaryEcdsaPrivateKey, err := sm4.GetBinaryEcdsaPrivateKeyFromFile(path, password)
	return binaryEcdsaPrivateKey, err
}

// 从导出的私钥文件读取私钥
func (gmcc GmCryptoClient) GetEcdsaPrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	ecdsaPrivateKey, err := sm2.GetEcdsaPrivateKeyFromFile(filename)
	return ecdsaPrivateKey, err
}

// 使用支付密码从导出的私钥文件读取私钥
func (gmcc GmCryptoClient) GetEcdsaPrivateKeyFromFileByPassword(path string, password string) (*ecdsa.PrivateKey, error) {
	ecdsaPrivateKey, err := sm4.GetEcdsaPrivateKeyFromFile(path, password)
	return ecdsaPrivateKey, err
}

// 从二进制加密字符串获取真实私钥的byte格式
func (gmcc GmCryptoClient) GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey string, password string) ([]byte, error) {
	binaryEcdsaPrivateKey, err := sm4.GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey, password)
	return binaryEcdsaPrivateKey, err
}

// 从导出的公钥文件读取公钥
func (gmcc GmCryptoClient) GetEcdsaPublicKeyFromFile(filename string) (*ecdsa.PublicKey, error) {
	ecdsaPublicKey, err := sm2.GetEcdsaPublicKeyFromFile(filename)
	return ecdsaPublicKey, err
}

// 使用ECIES加密
// TODO: 后面根据公钥中的标记为来判断需要使用哪种解密算法
func (gmcc GmCryptoClient) Encrypt(k *ecdsa.PublicKey, msg []byte) (cypherText []byte, err error) {
	// 判断是否是nist标准的私钥
	switch k.Params().Name {
	case config.CurveNist: // NIST
		cypherText, err := ecies.Encrypt(k, msg)
		return cypherText, err
	case config.CurveGm: // 国密
		cypherText, err := encrypt(k, msg)
		return cypherText, err
	default: // 不支持的密码学类型
		return nil, fmt.Errorf("This cryptography has not been supported yet.")
	}
}

// 使用ECIES加密
func encrypt(k *ecdsa.PublicKey, msg []byte) (cypherText []byte, err error) {
	key := new(sm2.PublicKey)
	//	key := &sm2.PrivateKey{}
	key.Curve = sm2.P256Sm2() // elliptic.P256()
	key.X = k.X
	key.Y = k.Y

	cypherText, err = sm2.Encrypt(key, msg)
	return cypherText, err
}

// 使用ECIES解密
func (gmcc GmCryptoClient) Decrypt(k *ecdsa.PrivateKey, cypherText []byte) (msg []byte, err error) {
	// 判断是否是nist标准的私钥
	switch k.Params().Name {
	case config.CurveNist: // NIST
		msg, err := ecies.Decrypt(k, cypherText)
		return msg, err
	case config.CurveGm: // 国密
		msg, err := decrypt(k, cypherText)
		return msg, err
	default: // 不支持的密码学类型
		return nil, fmt.Errorf("This cryptography has not been supported yet.")
	}
}

// 使用ECIES解密
func decrypt(k *ecdsa.PrivateKey, cypherText []byte) (msg []byte, err error) {
	key := new(sm2.PrivateKey)
	//	key := &sm2.PrivateKey{}
	key.PublicKey.Curve = sm2.P256Sm2() // elliptic.P256()
	key.X = k.X
	key.Y = k.Y
	key.D = k.D

	msg, err = sm2.Decrypt(key, cypherText)
	return msg, err
}

// 从导出的私钥文件读取私钥
func (gmcc GmCryptoClient) GetEcdsaPrivateKeyFromJSON(jsonBytes []byte) (*ecdsa.PrivateKey, error) {
	return sm2.GetEcdsaPrivateKeyFromJson(jsonBytes)
}

// 从导出的公钥文件读取公钥
func (gmcc GmCryptoClient) GetEcdsaPublicKeyFromJSON(jsonBytes []byte) (*ecdsa.PublicKey, error) {
	return sm2.GetEcdsaPublicKeyFromJson(jsonBytes)
}

// 使用对称加密算法加密
func EncryptByKey(info string, key string) (string, error) {
	//TODO
	return "", nil
}

// 使用对称加密算法解密
func DecryptByKey(cipherInfo string, key string) (string, error) {
	//TODO
	return "", nil
}

// 从云端获取已经加密的账户
func GetEncryptedAccountFromCloud(bduss string) (*account.ECDSAAccountToCloud, error) {
	//TODO
	return nil, nil
}

// 从云端获取未加密的账户
func GetAccountFromCloud(bduss string) (*account.ECDSAAccount, error) {
	//TODO
	return nil, nil
}

// 将经支付密码加密的账户保存到云端
func SaveEncryptedAccountToCloud(account *account.ECDSAAccountToCloud, bduss string) error {
	//TODO
	return nil
}

// 将经过支付密码加密的账户保存到文件中
func SaveEncryptedAccountToFile(account *account.ECDSAAccountToCloud, path string) error {
	//TODO
	return nil
}

// --- 多重签名相关 start ---

// 每个多重签名算法流程的参与节点生成32位长度的随机byte，返回值可以认为是k
func (gmcc GmCryptoClient) GetRandom32Bytes() ([]byte, error) {
	return multisign.GetRandom32Bytes()
}

// 每个多重签名算法流程的参与节点生成Ri = Ki*G
func (gmcc GmCryptoClient) GetRiUsingRandomBytes(key *ecdsa.PublicKey, k []byte) []byte {
	return multisign.GetRiUsingRandomBytes(key, k)
}

// 负责计算多重签名的节点来收集所有节点的Ri，并计算R = k1*G + k2*G + ... + kn*G
func (gmcc GmCryptoClient) GetRUsingAllRi(key *ecdsa.PublicKey, arrayOfRi [][]byte) []byte {
	return multisign.GetRUsingAllRi(key, arrayOfRi)
}

// 负责计算多重签名的节点来收集所有节点的公钥Pi，并计算公共公钥：C = P1 + P2 + ... + Pn
func (gmcc GmCryptoClient) GetSharedPublicKeyForPublicKeys(keys []*ecdsa.PublicKey) ([]byte, error) {
	return multisign.GetSharedPublicKeyForPublicKeys(keys)
}

// 负责计算多重签名的节点将计算出的R和C分别传递给各个参与节点后，由各个参与节点再次计算自己的Si
// 计算 Si = Ki + HASH(C,R,m) * Xi
// X代表大数D，也就是私钥的关键参数
func (gmcc GmCryptoClient) GetSiUsingKCRM(key *ecdsa.PrivateKey, k []byte, c []byte, r []byte, message []byte) []byte {
	return multisign.GetSiUsingKCRM(key, k, c, r, message)
}

// 负责计算多重签名的节点来收集所有节点的Si，并计算出S = sum(si)
func (gmcc GmCryptoClient) GetSUsingAllSi(arrayOfSi [][]byte) []byte {
	return multisign.GetSUsingAllSi(arrayOfSi)
}

// 负责计算多重签名的节点，最终生成多重签名
//func (gmcc GmCryptoClient) GenerateMultiSignSignature(s []byte, r []byte) (*multisign.MultiSignature, error) {
func (gmcc GmCryptoClient) GenerateMultiSignSignature(s []byte, r []byte) ([]byte, error) {
	return multisign.GenerateMultiSignSignature(s, r)
}

// 使用ECC公钥数组来进行多重签名的验证
//func (gmcc GmCryptoClient) VerifyMultiSig(keys []*ecdsa.PublicKey, signature *multisign.MultiSignature, message []byte) (bool, error) {
func (gmcc GmCryptoClient) VerifyMultiSig(keys []*ecdsa.PublicKey, signature []byte, message []byte) (bool, error) {
	// 判断是否是nist标准的私钥
	if len(keys) < 2 {
		return false, fmt.Errorf("The total num of keys should be greater than two.")
	}

	switch keys[0].Params().Name {
	case config.CurveNist: // NIST
		signature, err := multisign.VerifyMultiSig(keys, signature, message)
		return signature, err
	case config.CurveGm: // 国密
		signature, err := multisign.VerifyMultiSig(keys, signature, message)
		return signature, err
	default: // 不支持的密码学类型
		return false, fmt.Errorf("This cryptography has not been supported yet.")
	}
}

// -- 多重签名的另一种用法，适用于完全中心化的流程

// 使用ECC私钥数组来进行多重签名
//func (gmcc GmCryptoClient) MultiSign(keys []*ecdsa.PrivateKey, message []byte) (*multisign.MultiSignature, error) {
func (gmcc GmCryptoClient) MultiSign(keys []*ecdsa.PrivateKey, message []byte) ([]byte, error) {
	// 判断是否是nist标准的私钥
	if len(keys) < 2 {
		return nil, fmt.Errorf("The total num of keys should be greater than two.")
	}

	switch keys[0].Params().Name {
	case config.CurveNist: // NIST
		signature, err := multisign.MultiSign(keys, message)
		return signature, err
	case config.CurveGm: // 国密
		signature, err := multisign.MultiSign(keys, message)
		return signature, err
	default: // 不支持的密码学类型
		return nil, fmt.Errorf("This cryptography has not been supported yet.")
	}
}

// --- 多重签名相关 end ---

// --- 	schnorr签名算法相关 start ---schnorr_sign

// schnorr签名算法 生成签名
func (gmcc GmCryptoClient) SignSchnorr(privateKey *ecdsa.PrivateKey, message []byte) (schnorrSignature []byte, err error) {
	return schnorr_sign.Sign(privateKey, message)
}

// schnorr签名算法 验证签名
func (gmcc GmCryptoClient) VerifySchnorr(publicKey *ecdsa.PublicKey, sig []byte, message []byte) (valid bool, err error) {
	return schnorr_sign.Verify(publicKey, sig, message)
}

// --- 	schnorr签名算法相关 end ---

// --- 	schnorr 环签名算法相关 start ---

// 将签名者的私钥的公钥隐藏在参数1的公钥列表里
func (gmcc GmCryptoClient) SignSchnorrRing(keys []*ecdsa.PublicKey, privateKey *ecdsa.PrivateKey, message []byte) (ringSignature []byte, err error) {
	return schnorr_ring_sign.Sign(keys, privateKey, message)
}

func (gmcc GmCryptoClient) VerifySchnorrRing(keys []*ecdsa.PublicKey, sig, message []byte) (bool, error) {
	return schnorr_ring_sign.Verify(keys, sig, message)
}

// --- 	schnorr 环签名算法相关 end ---

// --- 统一验签算法
func (gmcc GmCryptoClient) XuperVerify(publicKeys []*ecdsa.PublicKey, sig []byte, message []byte) (valid bool, err error) {
	return signature.XuperSigVerify(publicKeys, sig, message)
}
