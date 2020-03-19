package api

import (
	"encoding/json"
	//	"log"

	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/ecies"
	"github.com/xuperchain/xuperchain/core/hdwallet/keychain"

	walletRand "github.com/xuperchain/xuperchain/core/hdwallet/rand"
)

const (
	// HardenedKeyStart is the index at which a hardended key starts.
	// Each extended key has 2^31 normal child keys, and 2^31 hardned child keys.
	// Each of these child keys has an index.
	// The range for normal child keys is [0, 2^31 - 1]
	// The range for hardened child keys is [2^31, 2^32 - 1].
	HardenedKeyStart = 0x80000000 // 2^31 -- 16进制
)

// 通过助记词恢复出分层确定性根密钥
func GenerateMasterKeyByMnemonic(mnemonic string, language int) (string, error) {
	// 判断密码学算法是否支持
	_, cryptography, err := account.GetCryptoByteFromMnemonic(mnemonic, language)
	if err != nil {
		//		log.Printf("GetCryptoByteFromMnemonic failed, Mnemonic might be invalid")
		return "", err
	}

	// 将助记词转为随机数种子，在此过程中，校验助记词是否合法
	password := "jingbo is handsome!"
	seed, err := walletRand.GenerateSeedWithErrorChecking(mnemonic, password, 40, language)
	if err != nil {
		return "", err
	}

	masterKey, err := keychain.NewMaster(seed, cryptography)
	if err != nil {
		return "", err
	}

	jsonMasterKey, err := json.Marshal(masterKey)
	if err != nil {
		return "", err
	}

	return string(jsonMasterKey), nil
}

// 通过分层确定性私钥/公钥（如根私钥）推导出子私钥/公钥
func GenerateChildKey(key string, i uint32) (string, error) {
	var extendedKey *keychain.ExtendedKey
	err := json.Unmarshal([]byte(key), &extendedKey)
	if err != nil {
		return "", err
	}

	childKey, err := extendedKey.Child(i)
	if err != nil {
		return "", err
	}

	jsonChildKey, err := json.Marshal(childKey)
	if err != nil {
		return "", err
	}

	return string(jsonChildKey), nil
}

// 将分层确定性私钥转化为公钥
func ConvertPrvKeyToPubKey(key string) (string, error) {
	var extendedKey *keychain.ExtendedKey
	err := json.Unmarshal([]byte(key), &extendedKey)
	if err != nil {
		return "", err
	}

	publicKey, err := extendedKey.Neuter()
	if err != nil {
		return "", err
	}

	jsonPublicKey, err := json.Marshal(publicKey)
	if err != nil {
		return "", err
	}

	return string(jsonPublicKey), nil
}

// 使用子公钥加密
func Encrypt(publicKey, msg string) (string, error) {
	var extendedKey *keychain.ExtendedKey
	err := json.Unmarshal([]byte(publicKey), &extendedKey)
	if err != nil {
		return "", err
	}

	publicEcdsaKey, err := extendedKey.ECPublicKey()
	if err != nil {
		return "", err
	}

	cypherText, err := ecies.Encrypt(publicEcdsaKey, []byte(msg))
	return string(cypherText), err
}

// 使用子公钥和根私钥解密
func Decrypt(publicKey, masterKey, cypherText string) (string, error) {
	var extendedKey *keychain.ExtendedKey
	err := json.Unmarshal([]byte(publicKey), &extendedKey)
	if err != nil {
		return "", err
	}

	var extendedMasterKey *keychain.ExtendedKey
	err = json.Unmarshal([]byte(masterKey), &extendedMasterKey)
	if err != nil {
		return "", err
	}

	extendedPrivateKey, err := extendedMasterKey.CorrespondingPrivateChild(extendedKey)
	if err != nil {
		return "", err
	}

	privateEcdsaKey, err := extendedPrivateKey.ECPrivateKey()
	if err != nil {
		return "", err
	}

	msg, err := ecies.Decrypt(privateEcdsaKey, []byte(cypherText))
	return string(msg), err
}
