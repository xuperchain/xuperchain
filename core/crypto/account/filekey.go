package account

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/xuperchain/xuperchain/core/crypto/config"
	"github.com/xuperchain/xuperchain/core/crypto/utils"
	walletRand "github.com/xuperchain/xuperchain/core/hdwallet/rand"
)

// 定义助记词的强度类型
const (
	// 不同语言标准不一样，这里用const直接定义值还是好一些
	// 低
	StrengthEasy = 1
	// 中
	StrengthMiddle = 2
	// 高
	StrengthHard = 3
)

// 定义密码算法的类型
const (
	// 不同语言标准不一样，这里用const直接定义值还是好一些
	// 美国Federal Information Processing Standards的椭圆曲线
	EccFIPS = iota
	// 国密椭圆曲线
	EccGM
)

// ECDSAAccount 助记词、私钥的json、公钥的json、钱包地址
type ECDSAAccount struct {
	EntropyByte    []byte
	Mnemonic       string
	JSONPrivateKey string
	JSONPublicKey  string
	Address        string
}

var (
	// Last4BitsMask 1111 - 11个1，当一个大的bigint和它进行“And”比特运算的时候，就会获得大的bigint最右边4位的比特位
	Last4BitsMask = big.NewInt(15)
	// Shift4BitsFactor 10000 - 乘以这个带有4个0的数等于左移4个比特位，除以这个带有4个0的数等于右移4个比特位，
	Shift4BitsFactor = big.NewInt(16)
)

var (
	// ErrStrengthNotSupported : 助记词的强度暂未被支持
	// Strength required for generating Mnemonic has not been supported yet.
	ErrStrengthNotSupported = fmt.Errorf("This strength has not been supported yet")

	// ErrCryptographyNotSupported : 密码学算法暂未被支持
	// Cryptography required for generating Mnemonic has not been supported yet.
	ErrCryptographyNotSupported = fmt.Errorf("This cryptography has not been supported yet")

	// ErrReservedTypeNotSupported : 语言暂未被支持
	// ReservedType required for generating Mnemonic has not been supported yet.
	ErrReservedTypeNotSupported = fmt.Errorf("This ReservedType has not been supported yet")
)

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

/**
 * 生成文件
 */
func writeFileUsingFilename(filename string, content []byte) error {
	//函数向filename指定的文件中写入数据(字节数组)。如果文件不存在将按给出的权限创建文件，否则在写入数据之前清空文件。
	err := ioutil.WriteFile(filename, content, 0666)
	return err
}

/*
 *	生成文件 调用内部方法 外部提供给其它包使用
 */
func WriteToFile(filename string, content []byte) error {
	return writeFileUsingFilename(filename, content)
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

// GenerateAccountByMnemonic 通过助记词来产生/恢复钱包账户
// 这里不应该再需要知道指定曲线了，也不需要指定地址版本了，这个功能应该由助记词中的标记位来判断
func GenerateAccountByMnemonic(mnemonic string, language int) (*ECDSAAccount, error) {
	// 判断密码学算法是否支持
	isOld, cryptography, err := GetCryptoByteFromMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
	}

	if !isOld && cryptography != config.Nist && cryptography != config.NistSN {
		err = fmt.Errorf("Only cryptoGraphy NIST[%d or %d] is supported in this version, this cryptoGraphy is [%v]",
			config.Nist, config.NistSN, cryptography)
		return nil, err
	}
	curve := elliptic.P256()

	// 将助记词转为随机数种子，在此过程中，校验助记词是否合法
	password := "jingbo is handsome!"
	var seed []byte
	if isOld {
		seed, err = walletRand.GenerateOldSeedWithErrorChecking(mnemonic, password, 40, language)
	} else {
		seed, err = walletRand.GenerateSeedWithErrorChecking(mnemonic, password, 40, language)
	}

	if err != nil {
		log.Println("GenerateSeedWithErrorChecking failed, err=", err)
		return nil, err
	}

	// 通过随机数种子来生成椭圆曲线加密的私钥
	privateKey, err := utils.GenerateKeyBySeed(curve, seed)

	if err != nil {
		log.Println("GenerateKeyBySeed failed, err=", err)
		return nil, err
	}
	// 获取私钥的json格式的字符串
	jsonPrivateKey, err := GetEcdsaPrivateKeyJSONFormat(privateKey)
	if err != nil {
		log.Println("GetEcdsaPrivateKeyJSONFormat failed, err=", err)
		return nil, err
	}
	// 通过公钥的json格式的字符串
	jsonPublicKey, err := GetEcdsaPublicKeyJSONFormat(privateKey)
	if err != nil {
		log.Println("GetEcdsaPublicKeyJSONFormat failed, err=", err)
		return nil, err
	}
	// 使用公钥来生成钱包地址
	address, err := GetAddressFromPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Println("GetAddressFromPublicKey failed, err=", err)
		return nil, err
	}
	// 返回的字段：助记词、私钥的json、公钥的json、钱包地址、错误信息
	account := &ECDSAAccount{
		EntropyByte:    seed,
		Mnemonic:       mnemonic,
		JSONPrivateKey: jsonPrivateKey,
		JSONPublicKey:  jsonPublicKey,
		Address:        address,
	}
	//	return &ECDSAAccount{nil, mnemonic, jsonPrivateKey, jsonPublicKey, address}, nil
	return account, nil
}

// CreateNewAccountWithMnemonic 参数字段：版本号、语言、强度
// 返回的字段：助记词、私钥的json、公钥的json、钱包地址、错误信息
func CreateNewAccountWithMnemonic(language int, strength uint8, cryptography uint8) (*ECDSAAccount, error) {
	var entropyBitLength = 0
	// 根据强度来判断随机数长度
	// 预留出8个bit用来指定当使用助记词时来恢复私钥时所需要的密码学算法组合
	switch strength {
	case StrengthEasy: // 弱 12个助记词
		entropyBitLength = 120
		//		entropyBitLength = 128
	case StrengthMiddle: // 中 18个助记词
		entropyBitLength = 184
		//		entropyBitLength = 192
	case StrengthHard: // 高 24个助记词
		entropyBitLength = 248
		//		entropyBitLength = 256
	default: // 不支持的语言类型
		entropyBitLength = 0
	}

	// 判断强度是否合法
	if entropyBitLength == 0 {
		return nil, ErrStrengthNotSupported
	}

	// 产生随机熵
	entropyByte, err := walletRand.GenerateEntropy(entropyBitLength)
	if err != nil {
		return nil, err
	}

	// 校验密码学算法是否得到支持
	var cryptographyBit = make([]byte, 1)

	switch cryptography {
	case config.Nist: // NIST
		cryptographyBit = []byte{config.Nist}
	case config.NistSN: // NIST+Schnorr
		cryptographyBit = []byte{config.Nist}
	case config.Gm: // 国密
		//		cryptographyBit = []byte{config.Gm}
		//      log.Printf("Only cryptoGraphy [NIST] is supported in this version.")
		return nil, ErrCryptographyNotSupported
	default: // 不支持的密码学类型
		return nil, ErrCryptographyNotSupported
	}

	// TODO: 把语言相关的标记位也加进去
	// 把带有密码学标记位的byte数组转化为一个bigint，方便后续做比特位运算（主要是移位操作）
	cryptographyInt := new(big.Int).SetBytes(cryptographyBit)
	// 创建综合标记位
	tagInt := big.NewInt(0)
	// 综合标记位获取密码学标记位最右边的4个比特
	tagInt.And(cryptographyInt, Last4BitsMask)

	// 将综合标记位左移4个比特
	tagInt.Mul(tagInt, Shift4BitsFactor)

	// 定义预留标记位
	var reservedBit = make([]byte, 1)
	reservedBit = []byte{0}

	//	switch reservedType {
	//	case config.ReservedType2: // 英文
	//		reservedBit = []byte{config.ReservedType1}
	//	case config.ReservedType2: // 中文
	//		reservedBit = []byte{config.ReservedType2}
	//	default: // 不支持的预留标记位类型
	//		return nil, ErrReservedTypeNotSupported
	//	}

	reservedInt := new(big.Int).SetBytes(reservedBit)

	// 综合标记位获取预留标记位最右边的4个比特
	reservedInt.And(reservedInt, Last4BitsMask)

	// 合并密码学标记位和预留标记位
	tagInt.Or(tagInt, reservedInt)

	//	log.Printf("tagInt now is %b", tagInt)

	// 把比特补齐为 1个字节
	tagByte := padByteSlice(tagInt.Bytes(), 1)

	newEntropyByteSlice := make([]byte, len(entropyByte)+len(tagByte))
	copy(newEntropyByteSlice, entropyByte)
	copy(newEntropyByteSlice[len(entropyByte):], tagByte)

	// 将随机熵转为指定语言的助记词
	mnemonic, err := walletRand.GenerateMnemonic(newEntropyByteSlice, language)
	if err != nil {
		return nil, err
	}
	// 通过助记词来产生钱包账户
	ecdsaAccount, err := GenerateAccountByMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
	}
	// 返回的字段：助记词、私钥的json、公钥的json、钱包地址、错误信息
	return ecdsaAccount, nil
}

// 把slice的长度补齐到指定的长度
func padByteSlice(slice []byte, length int) []byte {
	newSlice := make([]byte, length-len(slice))
	return append(newSlice, slice...)
}

// ExportNewAccountWithMnemonic create new account with mnemonic and export to local file
func ExportNewAccountWithMnemonic(path string, language int, strength uint8, cryptography uint8) error {
	// 先获得返回值
	ecdsaAccount, err := CreateNewAccountWithMnemonic(language, strength, cryptography)

	if err != nil {
		return err
	}
	// 把返回值保持到文件
	//如果path不是以/结尾的，自动拼上
	if strings.LastIndex(path, "/") != len([]rune(path))-1 {
		path = path + "/"
	}
	err = writeFileUsingFilename(path+"mnemonic", []byte(ecdsaAccount.Mnemonic))
	if err != nil {
		log.Printf("Export mnemonic file failed, the err is %v", err)
		return err
	}
	err = writeFileUsingFilename(path+"private.key", []byte(ecdsaAccount.JSONPrivateKey))
	if err != nil {
		log.Printf("Export private key file failed, the err is %v", err)
		return err
	}
	err = writeFileUsingFilename(path+"public.key", []byte(ecdsaAccount.JSONPublicKey))
	if err != nil {
		log.Printf("Export public key file failed, the err is %v", err)
		return err
	}
	err = writeFileUsingFilename(path+"address", []byte(ecdsaAccount.Address))
	if err != nil {
		log.Printf("Export address file failed, the err is %v", err)
		return err
	}
	//	log.Printf("Export address key file is successful, the path is %v", path+"address")
	return err
}

// GetAccInfo get account info from local data path
// TODO： 后续要废弃，已过时的方法
func GetAccInfo() ([]byte, []byte, []byte, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}
	jsonPrivateKey, err := GetEcdsaPrivateKeyJSONFormat(privateKey)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}
	jsonPublicKey, err := GetEcdsaPublicKeyJSONFormat(privateKey)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}
	address, err := GetAddressFromPublicKey(&privateKey.PublicKey)
	if err != nil {
		return []byte{}, []byte{}, []byte{}, err
	}
	return []byte(address), []byte(jsonPrivateKey), []byte(jsonPublicKey), nil
}

// ExportNewAccount creates new account and export to local file
func ExportNewAccount(path string, privateKey *ecdsa.PrivateKey) error {
	jsonPrivateKey, err := GetEcdsaPrivateKeyJSONFormat(privateKey)
	if err != nil {
		return err
	}
	jsonPublicKey, err := GetEcdsaPublicKeyJSONFormat(privateKey)
	if err != nil {
		return err
	}
	address, err := GetAddressFromPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	//如果path不是以/结尾的，自动拼上
	if strings.LastIndex(path, "/") != len([]rune(path))-1 {
		path = path + "/"
	}
	err = writeFileUsingFilename(path+"private.key", []byte(jsonPrivateKey))
	if err != nil {
		log.Printf("Export private key file failed, the err is %v", err)
		return err
	}
	err = writeFileUsingFilename(path+"public.key", []byte(jsonPublicKey))
	if err != nil {
		log.Printf("Export public key file failed, the err is %v", err)
		return err
	}
	err = writeFileUsingFilename(path+"address", []byte(address))
	if err != nil {
		log.Printf("Export address file failed, the err is %v", err)
		return err
	}
	return err
}

// GetEcdsaPrivateKeyFromJSON get ECDSA private key from json encoded data
func GetEcdsaPrivateKeyFromJSON(jsonContent []byte) (*ecdsa.PrivateKey, error) {
	privateKey := new(ECDSAPrivateKey)
	err := json.Unmarshal(jsonContent, privateKey)
	if err != nil {
		return nil, err
	}
	if privateKey.Curvname != "P-256" && privateKey.Curvname != "P-256-SN" {
		log.Printf("curve [%v] is not supported yet.\n", privateKey.Curvname)
		err = fmt.Errorf("curve [%v] is not supported yet", privateKey.Curvname)
		return nil, err
	}
	newPrivateKey := &ecdsa.PrivateKey{}
	newPrivateKey.PublicKey.Curve = elliptic.P256()
	newPrivateKey.PublicKey.Curve.Params().Name = privateKey.Curvname
	newPrivateKey.X = privateKey.X
	newPrivateKey.Y = privateKey.Y
	newPrivateKey.D = privateKey.D
	return newPrivateKey, nil
}

// GetEcdsaPrivateKeyFromFile get ECDSA private key from local file
func GetEcdsaPrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	content, err := readFileUsingFilename(filename)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}
	return GetEcdsaPrivateKeyFromJSON(content)
}

// GetEcdsaPublicKeyFromJSON get ECDSA public key from json encoded data
func GetEcdsaPublicKeyFromJSON(jsonContent []byte) (*ecdsa.PublicKey, error) {
	publicKey := new(ECDSAPublicKey)
	err := json.Unmarshal(jsonContent, publicKey)
	if err != nil {
		return nil, err //json有问题
	}
	if publicKey.Curvname != "P-256" && publicKey.Curvname != "P-256-SN" {
		log.Printf("curve [%v] is not supported yet.\n", publicKey.Curvname)
		err = fmt.Errorf("curve [%v] is not supported yet", publicKey.Curvname)
		return nil, err
	}
	newPublicKey := &ecdsa.PublicKey{}
	newPublicKey.Curve = elliptic.P256()
	newPublicKey.Curve.Params().Name = publicKey.Curvname
	newPublicKey.X = publicKey.X
	newPublicKey.Y = publicKey.Y
	return newPublicKey, nil
}

// GetEcdsaPublicKeyFromFile get ECDSA public key from local file
func GetEcdsaPublicKeyFromFile(filename string) (*ecdsa.PublicKey, error) {
	content, err := readFileUsingFilename(filename)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}
	return GetEcdsaPublicKeyFromJSON(content)
}

// GetAccInfoFromFile get account info from local file
func GetAccInfoFromFile(filename string) ([]byte, []byte, []byte, error) {
	addr, err := ioutil.ReadFile(filename + "/address")
	if err != nil {
		log.Printf("GetAccInfoFromFile error load address error = %v", err)
		return nil, nil, nil, err
	}
	pubkey, err := ioutil.ReadFile(filename + "/public.key")
	if err != nil {
		log.Printf("GetAccInfoFromFile error load pubkey error = %v", err)
		return nil, nil, nil, err
	}
	prikey, err := ioutil.ReadFile(filename + "/private.key")
	if err != nil {
		log.Printf("GetAccInfoFromFile error load prikey error = %v", err)
		return nil, nil, nil, err
	}
	return addr, pubkey, prikey, err
}

// GetCryptoByteFromMnemonic get crypto byte from mnemonic
// return values:
//    bool: is the mnemonic a old version, true means it's a old version of mnemonic
//    uint8: the cryptographic indicator
//    error: the error message
func GetCryptoByteFromMnemonic(mnemonic string, language int) (bool, uint8, error) {
	entropy, err := walletRand.GetEntropyFromMnemonic(mnemonic, language)
	if err != nil {
		// 再看看是不是旧版本的助记词
		entropy, err = walletRand.GetEntropyFromOldMnemonic(mnemonic, language)
		if err != nil {
			return false, 0, err
		}
		// old mnemonic
		return true, 0, nil
	}

	tagByte := entropy[len(entropy)-1:]
	tagInt := new(big.Int).SetBytes(tagByte)
	// log.Printf("tagInt is %b\n", tagInt)

	// 将熵右移4个比特
	tagInt.Div(tagInt, Shift4BitsFactor)

	// 综合标记位获取密码学标记位最右边的4个比特
	cryptographyInt := big.NewInt(0)
	cryptographyInt.And(tagInt, Last4BitsMask)

	cryptographyByte := cryptographyInt.Bytes()
	if len(cryptographyByte) == 0 {
		err = fmt.Errorf("cryptographyByte %v is not valid", cryptographyByte)
		return false, 0, err
	}
	cryptography := uint8(cryptographyByte[0])

	return false, cryptography, nil
}
