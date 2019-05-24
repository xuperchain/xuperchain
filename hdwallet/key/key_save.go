/*
Copyright Baidu Inc. All Rights Reserved.

把私钥加密后存储到客户端本地存储盘上，传入的信息是：私钥、对称加密的key（也就是用户的支付密码）、私钥存储地址
*/

package key

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/xuperchain/xuperunion/crypto/account"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"github.com/xuperchain/xuperunion/hdwallet/config"
	"github.com/xuperchain/xuperunion/hdwallet/error"
)

// AccountInfo describes the data struct for account information
type AccountInfo struct {
	Address           string
	PrivateKey        []byte
	PublicKey         []byte
	Mnemonic          []byte
	EncryptPrivateKey []byte
	EncryptPublicKey  []byte
	EncryptMnemonic   []byte
}

// CreateAndSaveSecretKeyWithMnemonic 通过助记词来恢复并保存私钥
// 这里不应该再需要知道指定曲线了，也不需要知道版本号了，这个功能应该由助记词中的标记位来判断
func CreateAndSaveSecretKeyWithMnemonic(path string, language int, mnemonic string, password string) (*account.ECDSAInfo, error) {
	// 通过助记词来产生钱包账户
	ecdsaAccount, err := account.GenerateAccountByMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
	}

	// 将私钥加密后保存到指定路径
	err = savePrivateKey(path, password, ecdsaAccount)
	if err != nil {
		return nil, err
	}

	// 返回的字段：助记词、随机byte数组、钱包地址
	ecdasaInfo := getECDSAInfoFromECDSAAccount(ecdsaAccount)

	return ecdasaInfo, nil
}

// EncryptAccount 使用支付密码加密账户信息
func EncryptAccount(info *account.ECDSAAccount, password string) (*account.ECDSAAccountToCloud, error) {
	if info.JSONPrivateKey == "" {
		return nil, spverror.ErrParam
	}

	// 将aes对称加密的密钥扩展至32字节
	newPassword := hash.DoubleSha256([]byte(password))

	// 加密私钥
	encryptedPrivateKey, err := aesEncrypt([]byte(info.JSONPrivateKey), newPassword)
	if err != nil {
		return nil, err
	}

	accountToClound := new(account.ECDSAAccountToCloud)
	accountToClound.JSONEncryptedPrivateKey = string(encryptedPrivateKey)
	accountToClound.Password = password
	accountToClound.Address = info.Address

	// 加密助记词
	if info.Mnemonic != "" {
		encryptedMnemonic, err := aesEncrypt([]byte(info.Mnemonic), newPassword)
		if err != nil {
			return nil, err
		}

		accountToClound.EncryptedMnemonic = string(encryptedMnemonic)
	}

	return accountToClound, nil
}

// SaveAccountFile 保存账户信息到文件,只需要保存address 和 privateKey
func SaveAccountFile(path string, address string, encryptPrivateKey []byte) error {
	//如果path不是以/结尾的，自动拼上
	if strings.LastIndex(path, "/") != len([]rune(path))-1 {
		path = path + "/"
	}
	err := writeFileUsingFilename(path+"address", []byte(address))
	if err != nil {
		return err
	}

	err = writeFileUsingFilename(path+"private.key", encryptPrivateKey)
	if err != nil {
		return err
	}
	return nil
}

// CreateAndSaveSecretKey 生成并保存私钥
func CreateAndSaveSecretKey(path string, language int, strength uint8, password string, cryptography uint8) (*account.ECDSAInfo, error) {
	//函数向filename指定的文件中写入数据(字节数组)。如果文件不存在将按给出的权限创建文件，否则在写入数据之前清空文件。
	ecdsaAccount, err := account.CreateNewAccountWithMnemonic(language, strength, cryptography)
	if err != nil {
		return nil, err
	}

	// 将私钥加密后保存到指定路径
	err = savePrivateKey(path, password, ecdsaAccount)
	if err != nil {
		return nil, err
	}

	// 返回的字段：助记词、随机byte数组、钱包地址
	ecdasaInfo := getECDSAInfoFromECDSAAccount(ecdsaAccount)

	return ecdasaInfo, err
}

// getECDSAInfoFromECDSAAccount 剔除掉ECDSAAccount需要隐藏的数据，返回的字段：助记词、随机byte数组、钱包地址
func getECDSAInfoFromECDSAAccount(ecdsaAccount *account.ECDSAAccount) *account.ECDSAInfo {
	ecdasaInfo := new(account.ECDSAInfo)
	ecdasaInfo.Mnemonic = ecdsaAccount.Mnemonic
	ecdasaInfo.EntropyByte = ecdsaAccount.EntropyByte
	ecdasaInfo.Address = ecdsaAccount.Address

	return ecdasaInfo
}

// savePrivateKey 将私钥加密后保存到指定路径
func savePrivateKey(path string, password string, ecdsaAccount *account.ECDSAAccount) error {
	//如果path不是以/结尾的，自动拼上
	if strings.LastIndex(path, "/") != len([]rune(path))-1 {
		path = path + "/"
	}

	// 将aes对称加密的密钥扩展至32字节
	newPassword := hash.DoubleSha256([]byte(password))

	// 加密密钥文件
	encryptContent, err := aesEncrypt([]byte(ecdsaAccount.JSONPrivateKey), newPassword)
	if err != nil {
		log.Printf("encrypt private key failed, the err is %v", err)
		return err
	}

	//	log.Printf("Export mnemonic file is successful, the path is %v", path+"mnemonic")
	err = writeFileUsingFilename(path+"private.key", encryptContent)
	if err != nil {
		log.Printf("Export private key file failed, the err is %v", err)
		return err
	}

	return nil
}

func aesEncrypt(origData, key []byte) ([]byte, error) {
	// 密钥长度只能是16、24、32字节，用以选择AES-128、AES-192、AES-256。
	// 非此长度范围会返回KeySizeError
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))

	blockMode.CryptBlocks(crypted, origData)

	return crypted, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// writeFileUsingFilename 保存私钥
func writeFileUsingFilename(filename string, content []byte) error {
	//函数向filename指定的文件中写入数据(字节数组)。如果文件不存在将按给出的权限创建文件，否则在写入数据之前清空文件。
	err := ioutil.WriteFile(filename, content, 0666)
	return err
}

// SaveAccountToServer 将账户信息保存到云端
func SaveAccountToServer(accountInfo *account.ECDSAAccountToCloud, bduss string) error {
	// 为请求进行赋值
	var r http.Request
	r.ParseForm()

	// 使用公钥机密bduss
	encodeBduss, err := EciesEncryptByJSONPublicKey(config.APIPublicKey, bduss)
	if err != nil {
		return err
	}
	encodeBdussByBase64 := base64.StdEncoding.EncodeToString([]byte(encodeBduss))
	r.Form.Add("bduss", string(encodeBdussByBase64))

	// 私钥加密后的二进制是不可见字符,需要base64编码
	encodeString := base64.StdEncoding.EncodeToString([]byte(accountInfo.JSONEncryptedPrivateKey))
	r.Form.Add("enpt_private_key", encodeString)
	r.Form.Add("address", accountInfo.Address)
	encodeString2 := base64.StdEncoding.EncodeToString([]byte(accountInfo.EncryptedMnemonic))
	r.Form.Add("enpt_mnemonic", encodeString2)

	// 对password加密传输
	key := generateRandomKey(8)
	password, err := EncryptByKey(accountInfo.Password, key)
	if err != nil {
		return err
	}
	r.Form.Add("key", key)
	encodePassword := base64.StdEncoding.EncodeToString([]byte(password))
	r.Form.Add("password", encodePassword)
	bodystr := strings.TrimSpace(r.Form.Encode())

	// 请求服务器
	client := &http.Client{
		Timeout: config.HTTPTimeOut,
	}
	req, err := http.NewRequest("POST", config.SaveEncryptedAccountURL, strings.NewReader(bodystr))
	if err != nil {
		return spverror.ErrRequestFailed
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return spverror.ErrRequestFailed
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return spverror.ErrReadResponseFailed
	}

	var respBody map[string]interface{}
	json.Unmarshal(body, &respBody)
	code, ok := respBody["code"].(float64)

	if ok {
		if code == 0 {
			return nil
		} else if code == 3010217 {
			return spverror.ErrNotLogin
		} else if code == 3020008 {
			return spverror.ErrRequestParam
		} else if code == 3010222 {
			return spverror.ErrPwExist
		} else if code == 3020010 {
			return spverror.ErrDbFail
		} else {
			return errors.New(respBody["msg"].(string))
		}
	}
	return spverror.ErrResponseFailed
}

// generateRandomKey 生成随机字符串
func generateRandomKey(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
