package sm2

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	//	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/config"
	"github.com/xuperchain/xuperchain/core/crypto/utils"
	walletRand "github.com/xuperchain/xuperchain/core/hdwallet/rand"
	//	"gmsm/sm3"
)

// 定义助记词的强度类型
const (
	// 不同语言标准不一样，这里用const直接定义值还是好一些
	//	_ = iota
	// 低
	StrengthEasy = 1
	// 中
	StrengthMiddle = 2
	// 高
	StrengthHard = 3
)

// 助记词、私钥的json、公钥的json、钱包地址
//type ECDSAAccount struct {
//	EntropyByte    []byte
//	Mnemonic       string
//	JSONPrivateKey string
//	JsonPublicKey  string
//	Address        string
//}

var (
	// 1111 - 11个1，当一个大的bigint和它进行“And”比特运算的时候，就会获得大的bigint最右边4位的比特位
	Last4BitsMask = big.NewInt(15)
	// 10000 - 乘以这个带有4个0的数等于左移4个比特位，除以这个带有4个0的数等于右移4个比特位，
	Shift4BitsFactor = big.NewInt(16)
)

var (
	// 助记词的强度暂未被支持
	// Strength required for generating Mnemonic not supported yet.
	ErrStrengthNotSupported = fmt.Errorf("This strength has not been supported yet.")

	// 密码学算法暂未被支持
	// Cryptography required for generating Mnemonic has not been supported yet.
	ErrCryptographyNotSupported = fmt.Errorf("This cryptography has not been supported yet.")
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
	//	//判断文件是否存在
	//	if checkFileIsExist(filename) {
	//		//打开文件
	//		f, err = os.OpenFile(filename, os.O_TRUNC, 0666)
	//		log.Printf("File [%v] exist", filename)
	//	} else {
	//		//创建文件
	//		f, err = os.Create(filename)
	//		log.Printf("File [%v] does not exist", filename)
	//	}
	//
	//	if err != nil {
	//		return err
	//	}
	//	var data = []byte(content)
	//函数向filename指定的文件中写入数据(字节数组)。如果文件不存在将按给出的权限创建文件，否则在写入数据之前清空文件。
	err := ioutil.WriteFile(filename, content, 0666)
	return err
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

// 参数字段：版本号、语言、强度
// 返回的字段：助记词、私钥的json、公钥的json、钱包地址、错误信息
func CreateNewAccountWithMnemonic(language int, strength uint8, cryptography uint8) (*account.ECDSAAccount, error) {
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
	case config.Gm: // 国密
		cryptographyBit = []byte{config.Gm}
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

	// 把比特补齐为 1个字节
	tagByte := padByteSlice(tagInt.Bytes(), 1)

	//	newEntropyByteSlice := make([]byte, len(entropyByte)+len(cryptographyBit))
	//	copy(newEntropyByteSlice, entropyByte)
	//	copy(newEntropyByteSlice[len(entropyByte):], cryptographyBit)

	newEntropyByteSlice := make([]byte, len(entropyByte)+len(tagByte))
	copy(newEntropyByteSlice, entropyByte)
	copy(newEntropyByteSlice[len(entropyByte):], tagByte)

	//	log.Printf("newEntropyByteSlice length is %v", len(newEntropyByteSlice))

	// 将随机熵转为指定语言的助记词
	//	mnemonic, err := walletRand.GenerateMnemonic(entropyByte, language)
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

// 通过助记词来恢复钱包账户
// 这里不再需要知道指定曲线了，这个功能应该由助记词中的标记位来判断
//func GenerateAccountByMnemonic(nVersion uint8, mnemonic string, language int, c elliptic.Curve) (*ECDSAAccount, error) {
func RetrieveAccountByMnemonic(mnemonic string, language int) (*account.ECDSAAccount, error) {
	// 判断密码学算法是否支持
	isOldMnemonic := false
	//	entropy, err := walletRand.GetEntropyFromMnemonic(mnemonic, language)
	//	if err != nil {
	//		// 再看看是不是旧版本的助记词
	//		entropy, err = walletRand.GetEntropyFromOldMnemonic(mnemonic, language)
	//		if err != nil {
	//			return nil, err
	//		}
	//		isOldMnemonic = true
	//	}
	//	cryptographyByte := entropy[len(entropy)-1:]
	//	cryptography := uint8(cryptographyByte[0])

	// 判断密码学算法是否支持
	cryptography, err := GetCryptoByteFromMnemonic(mnemonic, language)
	if err != nil {
		// 再看看是不是旧版本的助记词
		_, err = walletRand.GetEntropyFromOldMnemonic(mnemonic, language)
		if err != nil {
			return nil, err
		}
		isOldMnemonic = true
	}

	curve := elliptic.P256()

	//	log.Printf("cryptography is [%v]", cryptography)

	switch cryptography {
	case config.Nist: // NIST
		//		curve = elliptic.P256()
	case config.Gm: // 国密
		curve = P256Sm2()
	default: // 不支持的密码学类型同时还不是旧的助记词
		if isOldMnemonic == false {
			//		log.Printf("Only cryptoGraphy [NIST] and [SM2] is supported in this version.")
			err = fmt.Errorf("Only cryptoGraphy NIST[%d] and SM2[%d] is supported in this version, this cryptoGraphy is [%v].", config.Nist, config.Gm, cryptography)
			return nil, err
		}
		log.Printf("Old Mnemonic detected. CryptoGraphy [NIST] will be used in this case.")
	}

	// 将助记词转为随机数种子，在此过程中，校验助记词是否合法
	password := "jingbo is handsome!"
	seed, err := walletRand.GenerateSeedWithErrorChecking(mnemonic, password, 40, language)
	if err != nil {
		return nil, err
	}

	// 通过随机数种子来生成椭圆曲线加密的私钥
	//	privateKey, err := utils.GenerateKeyBySeed(elliptic.P256(), seed)
	privateKey, err := utils.GenerateKeyBySeed(curve, seed)

	if err != nil {
		return nil, err
	}
	// 获取私钥的json格式的字符串
	jsonPrivateKey, err := account.GetEcdsaPrivateKeyJSONFormat(privateKey)
	if err != nil {
		return nil, err
	}
	// 通过公钥的json格式的字符串
	jsonPublicKey, err := account.GetEcdsaPublicKeyJSONFormat(privateKey)
	if err != nil {
		return nil, err
	}
	// 使用公钥来生成钱包地址
	address, err := account.GetAddressFromPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	// 返回的字段：助记词、私钥的json、公钥的json、钱包地址、错误信息
	return &account.ECDSAAccount{nil, mnemonic, jsonPrivateKey, jsonPublicKey, address}, nil
}

func GetCryptoByteFromMnemonic(mnemonic string, language int) (uint8, error) {
	entropy, err := walletRand.GetEntropyFromMnemonic(mnemonic, language)
	if err != nil {
		return 0, err
	}
	//	cryptographyByte := entropy[len(entropy)-1:]
	//	cryptography := uint8(cryptographyByte[0])

	tagByte := entropy[len(entropy)-1:]
	tagInt := new(big.Int).SetBytes(tagByte)
	//	log.Printf("entropyInt is %b", new(big.Int).SetBytes(entropy))
	//	log.Printf("tagInt is %b", tagInt)

	// 将熵右移4个比特
	tagInt.Div(tagInt, Shift4BitsFactor)
	//	log.Printf("tagInt is %b", tagInt)

	// 综合标记位获取密码学标记位最右边的4个比特
	cryptographyInt := big.NewInt(0)
	cryptographyInt.And(tagInt, Last4BitsMask)

	cryptographyByte := cryptographyInt.Bytes()
	//	log.Printf("cryptographyByte is %v", cryptographyByte)
	if len(cryptographyByte) == 0 {
		err = fmt.Errorf("cryptographyByte %v is not valid.", cryptographyByte)
		return 0, err
	}
	cryptography := uint8(cryptographyByte[0])

	return cryptography, nil
}

// 通过助记词来产生/恢复钱包账户
// TODO: 这里不应该再需要知道指定曲线了，这个功能应该由助记词中的标记位来判断
//func GenerateAccountByMnemonic(nVersion uint8, mnemonic string, language int, c elliptic.Curve) (*ECDSAAccount, error) {
func GenerateAccountByMnemonic(mnemonic string, language int) (*account.ECDSAAccount, error) {
	// 判断密码学算法是否支持
	//	entropy, err := walletRand.GetEntropyFromMnemonic(mnemonic, language)
	//	if err != nil {
	//		return nil, err
	//	}
	//	cryptographyByte := entropy[len(entropy)-1:]
	//	cryptography := uint8(cryptographyByte[0])

	cryptography, err := GetCryptoByteFromMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
	}

	curve := elliptic.P256()

	//	log.Printf("cryptography is [%v]", cryptography)

	switch cryptography {
	case config.Nist: // NIST
		//		curve = elliptic.P256()
	case config.Gm: // 国密
		curve = P256Sm2()
	default: // 不支持的密码学类型
		//		log.Printf("Only cryptoGraphy [NIST] and [SM2] is supported in this version.")
		err = fmt.Errorf("Only cryptoGraphy NIST[%d] and SM2[%d] is supported in this version, this cryptoGraphy is [%v].", config.Nist, config.Gm, cryptography)
		return nil, err
	}

	// 将助记词转为随机数种子，在此过程中，校验助记词是否合法
	password := "jingbo is handsome!"
	seed, err := walletRand.GenerateSeedWithErrorChecking(mnemonic, password, 40, language)
	if err != nil {
		return nil, err
	}

	// 通过随机数种子来生成椭圆曲线加密的私钥
	//	privateKey, err := utils.GenerateKeyBySeed(elliptic.P256(), seed)
	privateKey, err := utils.GenerateKeyBySeed(curve, seed)

	if err != nil {
		return nil, err
	}
	// 获取私钥的json格式的字符串
	jsonPrivateKey, err := account.GetEcdsaPrivateKeyJSONFormat(privateKey)
	if err != nil {
		return nil, err
	}
	// 通过公钥的json格式的字符串
	jsonPublicKey, err := account.GetEcdsaPublicKeyJSONFormat(privateKey)
	if err != nil {
		return nil, err
	}
	// 使用公钥来生成钱包地址
	address, err := account.GetAddressFromPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	// 返回的字段：助记词、私钥的json、公钥的json、钱包地址、错误信息
	return &account.ECDSAAccount{nil, mnemonic, jsonPrivateKey, jsonPublicKey, address}, nil
}

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
	//	log.Printf("Export mnemonic file is successful, the path is %v", path+"mnemonic")
	err = writeFileUsingFilename(path+"private.key", []byte(ecdsaAccount.JSONPrivateKey))
	if err != nil {
		log.Printf("Export private key file failed, the err is %v", err)
		return err
	}
	//	log.Printf("Export private key file is successful, the path is %v", path+"private.key")
	err = writeFileUsingFilename(path+"public.key", []byte(ecdsaAccount.JSONPublicKey))
	if err != nil {
		log.Printf("Export public key file failed, the err is %v", err)
		return err
	}
	//	log.Printf("Export public key file is successful, the path is %v", path+"public.key")
	err = writeFileUsingFilename(path+"address", []byte(ecdsaAccount.Address))
	if err != nil {
		log.Printf("Export address file failed, the err is %v", err)
		return err
	}
	//	log.Printf("Export address key file is successful, the path is %v", path+"address")
	return err
}

func GetEcdsaPrivateKeyFromJson(jsonContent []byte) (*ecdsa.PrivateKey, error) {
	privateKey := new(account.ECDSAPrivateKey)
	err := json.Unmarshal(jsonContent, privateKey)
	if err != nil {
		return nil, err
	}
	//	if privateKey.Curvname != "SM2-P-256" {
	//		log.Printf("curve [%v] is not supported yet.", privateKey.Curvname)
	//		err = fmt.Errorf("curve [%v] is not supported yet.", privateKey.Curvname)
	//		return nil, err
	//	}
	var curve = elliptic.P256()

	switch privateKey.Curvname {
	case "P-256": // NIST
		//		curve = elliptic.P256()
	case "SM2-P-256": // 国密
		curve = P256Sm2()
	default: // 不支持的密码学类型
		log.Printf("Only cryptoGraphy [NIST] and [SM2] is supported in this version.")
		err = fmt.Errorf("curve [%v] is not supported yet.", privateKey.Curvname)
		return nil, err
	}

	newPrivateKey := &ecdsa.PrivateKey{}
	//	newPrivateKey.PublicKey.Curve = P256Sm2() // elliptic.P256()
	newPrivateKey.PublicKey.Curve = curve
	newPrivateKey.X = privateKey.X
	newPrivateKey.Y = privateKey.Y
	newPrivateKey.D = privateKey.D
	return newPrivateKey, nil
}
func GetEcdsaPrivateKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	content, err := readFileUsingFilename(filename)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}
	return GetEcdsaPrivateKeyFromJson(content)
}
func GetEcdsaPublicKeyFromJson(jsonContent []byte) (*ecdsa.PublicKey, error) {
	publicKey := new(account.ECDSAPublicKey)
	err := json.Unmarshal(jsonContent, publicKey)
	if err != nil {
		return nil, err //json有问题
	}
	//	if publicKey.Curvname != "SM2-P-256" {
	//		log.Printf("curve [%v] is not supported yet.", publicKey.Curvname)
	//		err = fmt.Errorf("curve [%v] is not supported yet.", publicKey.Curvname)
	//		return nil, err
	//	}

	var curve = elliptic.P256()

	switch publicKey.Curvname {
	case "P-256": // NIST
		//		curve = elliptic.P256()
	case "SM2-P-256": // 国密
		curve = P256Sm2()
	default: // 不支持的密码学类型
		log.Printf("Only cryptoGraphy [NIST] and [SM2] is supported in this version.")
		err = fmt.Errorf("curve [%v] is not supported yet.", publicKey.Curvname)
		return nil, err
	}

	newPublicKey := &ecdsa.PublicKey{}
	//	newPublicKey.Curve = P256Sm2() // elliptic.P256()
	newPublicKey.Curve = curve
	newPublicKey.X = publicKey.X
	newPublicKey.Y = publicKey.Y
	return newPublicKey, nil
}
func GetEcdsaPublicKeyFromFile(filename string) (*ecdsa.PublicKey, error) {
	content, err := readFileUsingFilename(filename)
	if err != nil {
		log.Printf("readFileUsingFilename failed, the err is %v", err)
		return nil, err
	}
	return GetEcdsaPublicKeyFromJson(content)
}
