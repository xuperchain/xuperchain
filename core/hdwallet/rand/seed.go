/*
Copyright Baidu Inc. All Rights Reserved.
*/

package rand

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	//	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/xuperchain/xuperchain/core/crypto/hash"
	"github.com/xuperchain/xuperchain/core/hdwallet/wordlist"

	"golang.org/x/crypto/pbkdf2"
)

// 定义助记词的语言类型
const (
	// 不同语言标准不一样，这里用const直接定义值还是好一些
	// 简体中文
	SimplifiedChinese = 1

	// 英文
	English = 2
)

// BigInt相关的比特位运算常量
var (
	// 11111111111 - 11个1，当一个大的bigint和它进行“And”比特运算的时候，就会获得大的bigint最右边11位的比特位
	Last11BitsMask = big.NewInt(2047)

	// 100000000000 - 除以这个带有11个0的数等于右移11个比特位
	RightShift11BitsDivider = big.NewInt(2048)

	// 1
	BigOne = big.NewInt(1)

	// 10
	BigTwo = big.NewInt(2)
)

// GenerateEntropy 底层调用跟操作系统相关的函数（读取系统熵）来产生一些伪随机数，
// 对外建议管这个返回值叫做“熵”
func GenerateEntropy(bitSize int) ([]byte, error) {
	//	err := validateEntropyBitSize(bitSize)
	err := validateRawEntropyBitSize(bitSize)
	if err != nil {
		return nil, err
	}

	entropy := make([]byte, bitSize/8)
	_, err = rand.Read(entropy)
	return entropy, err
}

//  检查试图获取的Entropy的比特大小是否符合规范要求：
//  在128-256之间，并且是32的倍数
//  为什么这么设计，详见比特币改进计划第39号提案的数学模型
//
//  checksum length (CS)
//  entropy length (ENT)
//  mnemonic sentence (MS)
//
//	CS = ENT / 32
//	MS = (ENT + CS) / 11
//
//	|  ENT  | CS | ENT+CS |  MS  |
//	+-------+----+--------+------+
//	|  128  |  4 |   132  |  12  |
//	|  160  |  5 |   165  |  15  |
//	|  192  |  6 |   198  |  18  |
//	|  224  |  7 |   231  |  21  |
//	|  256  |  8 |   264  |  24  |
func validateEntropyBitSize(bitSize int) error {
	if (bitSize%32) != 0 || bitSize < 128 || bitSize > 256 {
		return ErrInvalidEntropyLength
	}
	return nil
}

// +8的原因在于引入了8个bit的标记位来定义使用的密码学算法
func validateRawEntropyBitSize(bitSize int) error {
	if ((bitSize+8)%32) != 0 || (bitSize+8) < 128 || (bitSize+8) > 256 {
		return ErrInvalidRawEntropyLength
	}
	return nil
}

// 根据指定的语言类型来选择助记词list
func getWordListByLanguage(language int) ([]string, error) {
	var wordList = []string{"Not Supported Language List"}

	switch language {
	case SimplifiedChinese: // 简体中文
		wordList = wordlist.SimplifiedChineseWordList
	case English: // 英文
		wordList = wordlist.EnglishWordList
	default: // 不支持的语言类型
		wordList = nil
	}

	// 判断是否加载到了能够匹配的语言词库
	if wordList == nil {
		return nil, ErrLanguageNotSupported
	}

	return wordList, nil
}

// 根据指定的语言类型来选择反向助记词Map
func getReversedWordMapByLanguage(language int) (map[string]int, error) {
	var reversedWordMap = map[string]int{}

	switch language {
	case SimplifiedChinese: // 简体中文
		reversedWordMap = wordlist.ReversedSimplifiedChineseWordMap
	case English: // 英文
		reversedWordMap = wordlist.ReversedEnglishWordMap
	default: // 不支持的语言类型
		reversedWordMap = nil
	}

	// 判断是否加载到了能够匹配的语言词库
	if reversedWordMap == nil {
		return nil, ErrLanguageNotSupported
	}

	return reversedWordMap, nil
}

// GenerateMnemonic 根据参数中提供的熵来生成一串助记词。
// 参数中的熵应该是调用GenerateEntropy函数生成的熵。
func GenerateMnemonic(entropy []byte, language int) (string, error) {
	// 先获得参数中熵对应的比特位长度，1个字节=8个比特
	entropyBitLength := len(entropy) * 8

	// 万一有人不按照函数说明先调用GenerateEntropy函数来生成熵呢？
	// 拖出去TJJTDS
	// 这里还要再校验一遍熵的长度是否符合规范
	err := validateEntropyBitSize(entropyBitLength)
	if err != nil {
		return "", err
	}

	// 根据指定的语言类型来选择助记词词库
	wordList, err := getWordListByLanguage(language)

	// 判断是否加载到了能够匹配的语言词库
	if err != nil {
		return "", err
	}

	// 再根据熵的比特位长度来计算其校验值所需要的比特位长度
	checksumBitLength := entropyBitLength / 32

	// 然后计算拼接后的字符串能转换为多少个助记词
	// 注意：每11个比特位对应一个数字，数字范围是0-2047，数字会再转换为对应的助记词
	sentenceLength := (entropyBitLength + checksumBitLength) / 11

	// 熵的后面带上一段校验位
	//	log.Printf("entropy before add:%v", entropy)
	entropyWithChecksum := addChecksum(entropy)

	// 把熵切分为11个比特长度的片段
	// 把最右侧的11个比特长度的片段转化为数字，再匹配到对应的助记词
	// 然后再右移11个比特，再把最右侧的11个比特长度的片段转化为数字，再匹配到对应的助记词
	// 重复以上过程，直到熵被全部处理完成

	// 把带有校验值的熵转化为一个bigint，方便后续做比特位运算（主要是移位操作）
	entropyInt := new(big.Int).SetBytes(entropyWithChecksum)

	//	log.Printf("entropyInt now is %b", entropyInt)

	// 创建一个string slice来为保存助记词做准备
	words := make([]string, sentenceLength)

	// 创建一个比特位全是0的空词，为后面通过比特位“And与”运算来获取熵的11个比特长度的片段做准备
	word := big.NewInt(0)

	// 填充助记词slice
	for i := sentenceLength - 1; i >= 0; i-- {
		// 获取最右边的11个比特
		word.And(entropyInt, Last11BitsMask)

		// 将熵右移11个比特
		entropyInt.Div(entropyInt, RightShift11BitsDivider)

		// 把11个比特补齐为 2个字节
		wordBytes := padByteSlice(word.Bytes(), 2)

		// 将2个字节编码为Uint16格式，然后在word list里面寻找对应的助记词
		words[i] = wordList[binary.BigEndian.Uint16(wordBytes)]
	}

	// 返回助记词列表，空格分隔
	return strings.Join(words, " "), nil
}

// GenerateOldMnemonic 根据参数中提供的熵来生成一串助记词。
// 参数中的熵应该是调用GenerateEntropy函数生成的熵。
func GenerateOldMnemonic(entropy []byte, language int) (string, error) {
	// 先获得参数中熵对应的比特位长度，1个字节=8个比特
	entropyBitLength := len(entropy) * 8

	// 万一有人不按照函数说明先调用GenerateEntropy函数来生成熵呢？
	// 拖出去TJJTDS
	// 这里还要再校验一遍熵的长度是否符合规范
	err := validateEntropyBitSize(entropyBitLength)
	if err != nil {
		return "", err
	}

	// 根据指定的语言类型来选择助记词词库
	wordList, err := getWordListByLanguage(language)

	// 判断是否加载到了能够匹配的语言词库
	if err != nil {
		return "", err
	}

	// 再根据熵的比特位长度来计算其校验值所需要的比特位长度
	checksumBitLength := entropyBitLength / 32

	// 然后计算拼接后的字符串能转换为多少个助记词
	// 注意：每11个比特位对应一个数字，数字范围是0-2047，数字会再转换为对应的助记词
	sentenceLength := (entropyBitLength + checksumBitLength) / 11

	// 熵的后面带上一段校验位
	//	log.Printf("entropy before add:%v", entropy)
	entropyWithChecksum := addOldChecksum(entropy)

	// 把熵切分为11个比特长度的片段
	// 把最右侧的11个比特长度的片段转化为数字，再匹配到对应的助记词
	// 然后再右移11个比特，再把最右侧的11个比特长度的片段转化为数字，再匹配到对应的助记词
	// 重复以上过程，直到熵被全部处理完成

	// 把带有校验值的熵转化为一个bigint，方便后续做比特位运算（主要是移位操作）
	entropyInt := new(big.Int).SetBytes(entropyWithChecksum)

	//	log.Printf("entropyInt now is %b", entropyInt)

	// 创建一个string slice来为保存助记词做准备
	words := make([]string, sentenceLength)

	// 创建一个比特位全是0的空词，为后面通过比特位“And与”运算来获取熵的11个比特长度的片段做准备
	word := big.NewInt(0)

	// 填充助记词slice
	for i := sentenceLength - 1; i >= 0; i-- {
		// 获取最右边的11个比特
		word.And(entropyInt, Last11BitsMask)

		// 将熵右移11个比特
		entropyInt.Div(entropyInt, RightShift11BitsDivider)

		// 把11个比特补齐为 2个字节
		wordBytes := padByteSlice(word.Bytes(), 2)

		// 将2个字节编码为Uint16格式，然后在word list里面寻找对应的助记词
		words[i] = wordList[binary.BigEndian.Uint16(wordBytes)]
	}

	// 返回助记词列表，空格分隔
	return strings.Join(words, " "), nil
}

// GetEntropyFromMnemonic 从助记词提取原始熵的byte数组
func GetEntropyFromMnemonic(mnemonic string, language int) ([]byte, error) {
	// 先判断助记词是否合法，也就是判断是否每个词都存在于助记词列表中
	mnemonicSlice, err := GetWordsFromValidMnemonicSentence(mnemonic, language)
	if err != nil {
		return nil, err
	}

	// 再判断助记词的校验位是否合法
	mnemonicBitSize := len(mnemonicSlice) * 11

	// 进一步计算出校验位的比特位长度
	checksumBitSize := mnemonicBitSize % 32

	b := big.NewInt(0)
	//	modulo := big.NewInt(2048)
	// 根据语言加载对应的反向助记词map
	reversedWordMap, err := getReversedWordMapByLanguage(language)
	if err != nil {
		return nil, err
	}

	// 判断是否每个助记词
	for _, v := range mnemonicSlice {
		index, found := reversedWordMap[v]
		// 按理说通过了上面的检查GetWordsFromValidMnemonicSentence，这个问题不可能出现
		if found == false {
			return nil, fmt.Errorf("Word `%v` not found in the reversed map", v)
		}
		//		add := big.NewInt(int64(index))
		//		b = b.Mul(b, modulo)
		//		b = b.Add(b, add)

		var wordBytes [2]byte
		binary.BigEndian.PutUint16(wordBytes[:], uint16(index))
		// 左移11位，腾出11位的空间来
		b = b.Mul(b, RightShift11BitsDivider)
		// 给最右边的11位空间进行赋值
		b = b.Or(b, big.NewInt(0).SetBytes(wordBytes[:]))
	}

	// 从助记词+校验值组成的byte数组中计算出原始的随机熵
	checksumModulo := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(checksumBitSize)), nil)
	// 右移11位，来剔除掉校验位，获得原始的熵值
	entropy := big.NewInt(0).Div(b, checksumModulo)

	// 校验位最多有8个比特，计算出完整的字节长度
	// 计算出被用来计算校验位的原始内容的字节长度
	entropyByteSize := (mnemonicBitSize - checksumBitSize) / 8

	// 计算出包含计算出的校验位的内容的字节长度，校验位最多有8个比特，也就是一个字节
	fullByteSize := entropyByteSize + 1

	entropyBytes := padByteSlice(entropy.Bytes(), entropyByteSize)
	entropyWithChecksumBytes := padByteSlice(b.Bytes(), fullByteSize)

	// 检查校验位是否正确
	newEntropyWithChecksumBytes := padByteSlice(addChecksum(entropyBytes), fullByteSize)
	if !compareByteSlices(entropyWithChecksumBytes, newEntropyWithChecksumBytes) {
		//		return nil, ErrMnemonicChecksumIncorrect
		return nil, fmt.Errorf("The checksum within the new Mnemonic sentence incorrect fake:%v - real:%v, entropy:%v, mnemonic:%v", entropyWithChecksumBytes, newEntropyWithChecksumBytes, entropyBytes, mnemonic)
	}

	return entropy.Bytes(), nil
}

// GetEntropyFromOldMnemonic 从助记词提取原始熵的byte数组
func GetEntropyFromOldMnemonic(mnemonic string, language int) ([]byte, error) {
	// 先判断助记词是否合法，也就是判断是否每个词都存在于助记词列表中
	mnemonicSlice, err := GetWordsFromValidMnemonicSentence(mnemonic, language)
	if err != nil {
		return nil, err
	}

	// 再判断助记词的校验位是否合法
	mnemonicBitSize := len(mnemonicSlice) * 11

	// 进一步计算出校验位的比特位长度
	checksumBitSize := mnemonicBitSize % 32

	b := big.NewInt(0)
	//	modulo := big.NewInt(2048)
	// 根据语言加载对应的反向助记词map
	reversedWordMap, err := getReversedWordMapByLanguage(language)
	if err != nil {
		return nil, err
	}

	// 判断是否每个助记词
	for _, v := range mnemonicSlice {
		index, found := reversedWordMap[v]
		// 按理说通过了上面的检查GetWordsFromValidMnemonicSentence，这个问题不可能出现
		if found == false {
			return nil, fmt.Errorf("Word `%v` not found in the reversed map", v)
		}
		//		add := big.NewInt(int64(index))
		//		b = b.Mul(b, modulo)
		//		b = b.Add(b, add)

		var wordBytes [2]byte
		binary.BigEndian.PutUint16(wordBytes[:], uint16(index))
		// 左移11位，腾出11位的空间来
		b = b.Mul(b, RightShift11BitsDivider)
		// 给最右边的11位空间进行赋值
		b = b.Or(b, big.NewInt(0).SetBytes(wordBytes[:]))
	}

	// 从助记词+校验值组成的byte数组中计算出原始的随机熵
	checksumModulo := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(int64(checksumBitSize)), nil)
	// 右移11位，来剔除掉校验位，获得原始的熵值
	entropy := big.NewInt(0).Div(b, checksumModulo)

	// 校验位最多有8个比特，计算出完整的字节长度
	// 计算出被用来计算校验位的原始内容的字节长度
	entropyByteSize := (mnemonicBitSize - checksumBitSize) / 8

	// 计算出包含计算出的校验位的内容的字节长度，校验位最多有8个比特，也就是一个字节
	fullByteSize := entropyByteSize + 1

	entropyBytes := padByteSlice(entropy.Bytes(), entropyByteSize)
	entropyWithChecksumBytes := padByteSlice(b.Bytes(), fullByteSize)

	// 检查校验位是否正确
	newEntropyWithChecksumBytes := padByteSlice(addOldChecksum(entropyBytes), fullByteSize)
	if !compareByteSlices(entropyWithChecksumBytes, newEntropyWithChecksumBytes) {
		//		return nil, ErrMnemonicChecksumIncorrect
		return nil, fmt.Errorf("The checksum within the old Mnemonic sentence incorrect fake:%v - real:%v, entropy:%v, mnemonic:%v", entropyWithChecksumBytes, newEntropyWithChecksumBytes, entropyBytes, mnemonic)
	}

	return entropy.Bytes(), nil
}

// 比较两个字节数组的内容是否完全一致
func compareByteSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// GenerateSeedWithErrorChecking 带有错误检查。通过用户输入的助记词串（之前函数生成的）和用户指定的密码，来生成一个随机数种子
// 会校验助记词串是否
func GenerateSeedWithErrorChecking(mnemonic string, password string, keyLen int, language int) ([]byte, error) {
	_, err := GetEntropyFromMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
	}
	return generateSeed(mnemonic, password, keyLen), nil
}

// GenerateOldSeedWithErrorChecking 带有错误检查。通过用户输入的助记词串（之前函数生成的）和用户指定的密码，来生成一个随机数种子
// 老版本的助记词检查方法
func GenerateOldSeedWithErrorChecking(mnemonic string, password string, keyLen int, language int) ([]byte, error) {
	_, err := GetEntropyFromOldMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
	}
	return generateSeed(mnemonic, password, keyLen), nil
}

// 通过用户输入的助记词串（之前函数生成的）和用户指定的密码，来生成一个随机数种子
func generateSeed(mnemonic string, password string, keyLen int) []byte {
	salt := "mnemonic" + password
	//	seed := pbkdf2.Key([]byte(mnemonic), []byte(salt), 2048, 64, sha512.New)
	seed := pbkdf2.Key([]byte(mnemonic), []byte(salt), 2048, keyLen, sha512.New)

	return seed
}

// 计算 sha256(data)的前(len(data)/32)比特位的值作为校验码，
// 并将其加到data后面，然后返回新的带有校验码的data
func addChecksum(data []byte) []byte {
	// 获取sha256处理后的第二个字节作为校验码
	hashByte := hash.UsingSha256(data)
	firstChecksumByte := hashByte[1]

	//	log.Printf("addChecksum now... entropy:%v, hashByte:%v, firstChecksumByte:%v", data, hashByte, firstChecksumByte)

	// CS = ENT / 32
	// len() 相当于/8，所以这里再除以4就行了
	// 计算出校验位的比特长度
	checksumBitLength := uint(len(data) / 4)

	dataBigInt := new(big.Int).SetBytes(data)
	// 执行校验位长度N的循环，来生成长度N的校验位
	for i := uint(0); i < checksumBitLength; i++ {
		// 乘以10等于比特位运算左移一位，将原始熵全部左移一位
		dataBigInt.Mul(dataBigInt, BigTwo)

		// Set rightmost bit if leftmost checksum bit is set
		if uint8(firstChecksumByte&(1<<(7-i))) > 0 {
			// 与00000001进行异或，相当于对最右边的那个比特位进行计算，算出校验位
			dataBigInt.Or(dataBigInt, BigOne)
		}
	}

	//	log.Printf("checksum:%v", dataBigInt.Bytes())

	return dataBigInt.Bytes()
}

// 计算 sha256(data)的前(len(data)/32)比特位的值作为校验码，
// 并将其加到data后面，然后返回新的带有校验码的data
func addOldChecksum(data []byte) []byte {
	// 获取sha256处理后的第一个字节作为校验码
	hashByte := hash.UsingSha256(data)
	firstChecksumByte := hashByte[0]

	//	log.Printf("addChecksum now... entropy:%v, hashByte:%v, firstChecksumByte:%v", data, hashByte, firstChecksumByte)

	// CS = ENT / 32
	// len() 相当于/8，所以这里再除以4就行了
	// 计算出校验位的比特长度
	checksumBitLength := uint(len(data) / 4)

	dataBigInt := new(big.Int).SetBytes(data)
	// 执行校验位长度N的循环，来生成长度N的校验位
	for i := uint(0); i < checksumBitLength; i++ {
		// 乘以10等于比特位运算左移一位，将原始熵全部左移一位
		dataBigInt.Mul(dataBigInt, BigTwo)

		// Set rightmost bit if leftmost checksum bit is set
		if uint8(firstChecksumByte&(1<<(7-i))) > 0 {
			// 与00000001进行异或，相当于对最右边的那个比特位进行计算，算出校验位
			dataBigInt.Or(dataBigInt, BigOne)
		}
	}

	//	log.Printf("checksum:%v", dataBigInt.Bytes())

	return dataBigInt.Bytes()
}

// 把slice的长度补齐到指定的长度
func padByteSlice(slice []byte, length int) []byte {
	newSlice := make([]byte, length-len(slice))
	return append(newSlice, slice...)
}

// 取出助记词字符串中的所有助记词，并且同时检查助记词字符串包含的助记词数量是否有效
//  checksum length (CS)
//  entropy length (ENT)
//  mnemonic sentence (MS)
//
//	CS = ENT / 32
//	MS = (ENT + CS) / 11
//
//	|  ENT  | CS | ENT+CS |  MS  |
//	+-------+----+--------+------+
//	|  128  |  4 |   132  |  12  |
//	|  160  |  5 |   165  |  15  |
//	|  192  |  6 |   198  |  18  |
//	|  224  |  7 |   231  |  21  |
//	|  256  |  8 |   264  |  24  |
func getWordsFromMnemonicSentence(mnemonic string) ([]string, error) {
	// 将助记词字符串以空格符分割，返回包含助记词的list
	words := strings.Fields(mnemonic)

	//统计助记词的数量
	numOfWords := len(words)

	// 助记词的数量只能是 12, 15, 18, 21, 24
	validNumSlice := []string{"12", "15", "18", "21", "24"}
	if !stringInSlice(strconv.Itoa(numOfWords), validNumSlice) {
		return nil, ErrMnemonicNumNotValid
	}

	return words, nil
}

// 再检查是否所有传入的助记词都包含在对应语言的助记词列表中
func checkWordsWithinLanguageWordList(words []string, wordList []string) error {
	//统计助记词的数量
	numOfWords := len(words)
	// 再检查是否所有传入的助记词都包含在助记词列表中
	for i := 0; i < numOfWords; i++ {
		if !stringInSlice(words[i], wordList) {
			// 助记词不合法，单词未被支持
			return fmt.Errorf("Mnemonic word [%v] is not valid", words[i])
		}
	}

	return nil
}

// GetWordsFromValidMnemonicSentence 检查助记词字符串是否有效，如果有效，返回助记词
func GetWordsFromValidMnemonicSentence(mnemonic string, language int) ([]string, error) {
	// 将助记词字符串以空格符分割，返回包含助记词的list
	words, err := getWordsFromMnemonicSentence(mnemonic)
	// 判断是否从助记词字符串中成功的取出了符合数量要求的助记词
	if err != nil {
		return nil, err
	}

	// 根据指定的语言类型来选择助记词词库
	wordList, err := getWordListByLanguage(language)
	// 判断是否加载到了能够匹配的语言词库
	if err != nil {
		return nil, err
	}

	// 判断是否在对应语言的词库里
	err = checkWordsWithinLanguageWordList(words, wordList)
	// 判断是否都在词库里
	if err != nil {
		return nil, err
	}

	return words, nil
}

//相当于php的in array函数
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
