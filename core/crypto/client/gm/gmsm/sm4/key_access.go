/*
Copyright Baidu Inc. All Rights Reserved.

把客户端本地存储盘上的加密后存储的私钥解析出来，传入的信息是：对称加密的key（也就是用户的支付密码）、私钥存储地址
*/

package sm4

import (
	//	"crypto/aes"
	//	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/client/gm/gmsm/sm2"
	"github.com/xuperchain/xuperchain/core/crypto/ecies"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/hdwallet/config"
	spv_error "github.com/xuperchain/xuperchain/core/hdwallet/error"
)

func GetBinaryEcdsaPrivateKeyFromFile(path string, password string) ([]byte, error) {
	filename := path + "private.key"
	content, err := readFileUsingFilename(filename)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}

	//	// 将aes对称加密的密钥扩展至32字节
	//	newPassword := hash.DoubleSha256([]byte(password))

	// 国密SM4只支持16字节的key和分组
	salt := "jingbo is handsome."
	newPassword := pbkdf2.Key([]byte(password), []byte(salt), 256, 16, sha512.New)

	originalContent, err := aesDecrypt(content, newPassword)
	if err != nil {
		log.Printf("Decrypt private key file failed, the err is %v", err)
		return nil, err
	}

	return originalContent, nil
}

// GetBinaryEcdsaPrivateKeyFromString通过二进制字符串获取真实私钥
func GetBinaryEcdsaPrivateKeyFromString(encryptPrivateKey string, password string) ([]byte, error) {
	// 将aes对称加密的密钥扩展至32字节
	//	newPassword := hash.DoubleSha256([]byte(password))

	// 国密SM4只支持16字节的key和分组
	salt := "jingbo is handsome."
	newPassword := pbkdf2.Key([]byte(password), []byte(salt), 256, 16, sha512.New)

	originalContent, err := aesDecrypt([]byte(encryptPrivateKey), newPassword)
	if err != nil {
		log.Printf("Decrypt private key file failed, the err is %v", err)
		return nil, err
	}

	return originalContent, nil
}

func GetEcdsaPrivateKeyFromFile(path string, password string) (*ecdsa.PrivateKey, error) {
	originalContent, err := GetBinaryEcdsaPrivateKeyFromFile(path, password)
	if err != nil {
		log.Printf("GetBinaryEcdsaPrivateKeyFromFile failed, the err is %v", err)
		return nil, err
	}

	return sm2.GetEcdsaPrivateKeyFromJson(originalContent)
}

func aesDecrypt(crypted, key []byte) ([]byte, error) {
	//	block, err := aes.NewCipher(key)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	blockSize := block.BlockSize()
	//	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	//	origData := make([]byte, len(crypted))
	//
	//	blockMode.CryptBlocks(origData, crypted)

	// 使用国密的SM4分组对称加密算法
	block, err := NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()

	origData := make([]byte, len(crypted))
	origPart := make([]byte, blockSize)
	for i := 0; i < len(crypted); i = i + blockSize {
		block.Decrypt(origPart, crypted[i*blockSize:(i+1)*blockSize])
		copy(origData[i*blockSize:], origPart)
		origPart = make([]byte, blockSize)
	}

	return pkcs5UnPadding(origData)
}

func pkcs5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	unpadding := int(origData[length-1])

	if length-unpadding <= 0 {
		// 密码错误时 可能会造成为负数
		return nil, spv_error.ErrPwWrong
	}

	return origData[:(length - unpadding)], nil
}

/**
 * 读取文件
 */
func readFileUsingFilename(filename string) ([]byte, error) {
	//从filename指定的文件中读取数据并返回文件的内容。
	//成功的调用返回的err为nil而非EOF。
	//因为本函数定义为读取整个文件，它不会将读取返回的EOF视为应报告的错误。
	content, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		log.Printf("File [%v] does not exist", filename)
	}
	return content, err
}

func GetEcdsaPublicKeyFromJson(jsonContent []byte) (*ecdsa.PublicKey, error) {
	publicKey := new(account.ECDSAPublicKey)
	err := json.Unmarshal(jsonContent, publicKey)
	if err != nil {
		return nil, err //json有问题
	}
	if publicKey.Curvname != "P-256" {
		log.Printf("curve [%v] is not supported yet.", publicKey.Curvname)
		err = fmt.Errorf("curve [%v] is not supported yet.", publicKey.Curvname)
		return nil, err
	}
	lowLevelPublicKey := &ecdsa.PublicKey{}
	lowLevelPublicKey.Curve = elliptic.P256()
	lowLevelPublicKey.X = publicKey.X
	lowLevelPublicKey.Y = publicKey.Y
	return lowLevelPublicKey, nil
}

// GetAccountFromLocal 读取本地文件获取账户信息
func GetAccountFromLocal(path string) (*account.ECDSAAccountToCloud, error) {
	account := new(account.ECDSAAccountToCloud)
	privateKeyFile := path + "private.key"
	privateKey, err := readFileUsingFilename(privateKeyFile)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}

	addressFile := path + "address"
	address, err := readFileUsingFilename(addressFile)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}
	account.JSONEncryptedPrivateKey = string(privateKey)
	account.Address = string(address)
	return account, nil
}

// 获取云端有支付密码的账户
//func GetAccountFromServer(bduss string) (*pb.ECDSAAccountFromCloud, error) {
func GetAccountFromServer(bduss string) (*account.ECDSAAccountToCloud, error) {
	// 请求服务器
	client := &http.Client{
		Timeout: config.HTTPTimeOut,
	}

	var r http.Request
	r.ParseForm()

	// 使用公钥加密bduss
	encodeBduss, err := EciesEncryptByJsonPublicKey(config.APIPublicKey, bduss)
	if err != nil {
		return nil, err
	}
	encodeBdussByBase64 := base64.StdEncoding.EncodeToString([]byte(encodeBduss))
	r.Form.Add("bduss", string(encodeBdussByBase64))

	bodystr := strings.TrimSpace(r.Form.Encode())
	// 使用配置文件中查询passport账户在云端是否已经绑定区块链账户的url接口:config.QueryAccountUrl
	req, err := http.NewRequest("POST", config.QueryEncryptedAccountURL, strings.NewReader(bodystr))
	if err != nil {
		return nil, spv_error.ErrRequestFailed
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, spv_error.ErrRequestFailed
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, spv_error.ErrRequestFailed
	}

	var respBody map[string]interface{}
	json.Unmarshal(body, &respBody)
	code, ok := respBody["code"].(float64)
	if !ok {
		return nil, spv_error.ErrRequestFailed
	}
	log.Printf("Response body is: %v", respBody)

	// 获取请求
	if code == 3010221 {
		// 支付密码不存在
		return nil, spv_error.ErrPwNotExist
	} else if code == 3010217 {
		return nil, spv_error.ErrNotLogin
	} else if code == 3010218 {
		// 用户账户不存在
		return nil, spv_error.ErrAccountNotExist
	} else if code == 0 {
		data, ok := respBody["data"].(map[string]interface{})
		if ok {
			//			ret := new(pb.ECDSAAccountFromCloud)
			ret := new(account.ECDSAAccountToCloud)
			//			fmt.Println(respBody["data"])
			ret.Address = data["address"].(string)
			decodeBytes, err := base64.StdEncoding.DecodeString(data["enpt_private_key"].(string))
			if err != nil {
				log.Fatalln(err)
			}
			ret.JSONEncryptedPrivateKey = string(decodeBytes)

			decodeBytes, err = base64.StdEncoding.DecodeString(data["enpt_mnemonic"].(string))
			if err != nil {
				log.Fatalln(err)
			}
			ret.EncryptedMnemonic = string(decodeBytes)

			return ret, nil
		}
	}
	return nil, fmt.Errorf(respBody["msg"].(string))
}

// 获取云端没有支付密码的账户
func GetOriginalAccountFromServer(bduss string) (*account.ECDSAAccount, error) {
	// 请求服务器
	client := &http.Client{
		Timeout: config.HTTPTimeOut,
	}

	var r http.Request
	r.ParseForm()

	// 使用公钥加密bduss
	encodeBduss, err := EciesEncryptByJsonPublicKey(config.APIPublicKey, bduss)
	if err != nil {
		return nil, err
	}
	encodeBdussByBase64 := base64.StdEncoding.EncodeToString([]byte(encodeBduss))
	r.Form.Add("bduss", string(encodeBdussByBase64))

	bodystr := strings.TrimSpace(r.Form.Encode())
	req, err := http.NewRequest("POST", config.QueryPlainAccountURL, strings.NewReader(bodystr))
	if err != nil {
		return nil, spv_error.ErrRequestFailed
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return nil, spv_error.ErrRequestFailed
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, spv_error.ErrRequestFailed
	}

	var respBody map[string]interface{}
	json.Unmarshal(body, &respBody)
	code, ok := respBody["code"].(float64)
	if !ok {
		return nil, spv_error.ErrRequestFailed
	}
	// 获取请求
	if code == 3010222 {
		// 支付密码已存在
		return nil, spv_error.ErrPwExist
	} else if code == 3010217 {
		return nil, spv_error.ErrNotLogin
	} else if code == 3010218 {
		// 用户账户不存在
		return nil, spv_error.ErrAccountNotExist
	} else if code == 0 {
		data, ok := respBody["data"].(map[string]interface{})
		if ok {
			ret := new(account.ECDSAAccount)
			// 采用对称加密方式传递私钥和助记词,解密
			privateKey := data["private_key"].(string)
			decodePrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
			if err != nil {
				return nil, err
			}
			mnemonicKey := data["mnemonic"].(string)
			decodeMnemonicKey, err := base64.StdEncoding.DecodeString(mnemonicKey)
			if err != nil {
				return nil, err
			}

			key := data["key"].(string)
			ret.Address = data["address"].(string)
			ret.JSONPrivateKey, err = DecryptByKey(string(decodePrivateKey), key)
			if err != nil {
				return nil, err
			}
			ret.Mnemonic, err = DecryptByKey(string(decodeMnemonicKey), key)
			if err != nil {
				return nil, err
			}

			return ret, nil
		}
	}
	return nil, fmt.Errorf(respBody["msg"].(string))
}

// EncryptByKey 加密
func EncryptByKey(info string, key string) (string, error) {
	// 将aes对称加密的密钥扩展至32字节
	newPassword := hash.DoubleSha256([]byte(key))

	// 加密info
	cipherInfo, err := aesEncrypt([]byte(info), newPassword)
	if err != nil {
		return "", err
	}
	return string(cipherInfo), err
}

// DecryptByKey 解密
func DecryptByKey(cipherInfo string, key string) (string, error) {
	// 将aes对称加密的密钥扩展至32字节
	newPassword := hash.DoubleSha256([]byte(key))

	// 解密cipherInfo
	info, err := aesDecrypt([]byte(cipherInfo), newPassword)
	if err != nil {
		return "", err
	}
	return string(info), nil
}

// GetPublicKeyByPrivateKey通过私钥获取公钥
func GetPublicKeyByPrivateKey(binaryPrivateKey string) (string, error) {
	privatekey, err := sm2.GetEcdsaPrivateKeyFromJson([]byte(binaryPrivateKey))
	if err != nil {
		return "", err
	}

	// 补充公钥
	jsonPublicKey, err := account.GetEcdsaPublicKeyJSONFormat(privatekey)
	if err != nil {
		return "", err
	}
	return jsonPublicKey, nil
}

// EciesEncryptByJsonPublicKey 使用字符串公钥进行ecies加密
func EciesEncryptByJsonPublicKey(publicKey string, msg string) (string, error) {
	apiPublicKey, err := GetEcdsaPublicKeyFromJson([]byte(publicKey))
	if err != nil {
		return "", errors.New("api public key is wrong")
	}
	cipherInfo, err := ecies.Encrypt(apiPublicKey, []byte(msg))
	if err != nil {
		return "", spv_error.ErrParam
	}
	return string(cipherInfo), nil
}

// EciesDecryptByJsonPublicKey 使用字符串私钥进行ecies解密
func EciesDecryptByJSONPrivateKey(privateKey string, cipherInfo string) (string, error) {
	apiPrivateKey, err := sm2.GetEcdsaPrivateKeyFromJson([]byte(privateKey))
	if err != nil {
		return "", errors.New("api public key is wrong")
	}
	msg, err := ecies.Decrypt(apiPrivateKey, []byte(cipherInfo))
	if err != nil {
		return "", spv_error.ErrParam
	}
	return string(msg), nil
}
