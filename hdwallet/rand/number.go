/*
Copyright Baidu Inc. All Rights Reserved.
*/

package rand

import (
	"crypto/rand"
	"crypto/sha512"

	"golang.org/x/crypto/pbkdf2"
)

// 定义不同int类型对应的key length
const (
	// int8 类型
	KeyLengthInt8 = 8

	// int16 类型
	KeyLengthInt16 = 16

	// int32 类型
	KeyLengthInt32 = 32

	// int64 类型
	KeyLengthInt64 = 64
)

const (
	// KeyStrengthEasy 安全强度低
	KeyStrengthEasy = iota

	// KeyStrengthMiddle 安全强度中
	KeyStrengthMiddle

	// KeyStrengthHard 安全强度高
	KeyStrengthHard
)

// 底层调用跟操作系统相关的函数（读取系统熵）来产生一些伪随机数，
// 对外建议管这个返回值叫做“熵”
func generateEntropy(bitSize int) ([]byte, error) {
	err := validateEntropyBitSize(bitSize)
	if err != nil {
		return nil, err
	}

	entropy := make([]byte, bitSize/8)
	_, err = rand.Read(entropy)
	return entropy, err
}

//  检查试图获取的Entropy的比特大小是否符合规范要求：
//  在8-64之间，并且是8的倍数
//
//  |  类型               | 字节 | 比特位 |
//	+-------------+-----+-------+
//	|  char       |  1  |   8   |
//	|  short int  |  2  |   16  |
//	|  int        |  4  |   32  |
//	|  in64       |  8  |   64  |
//func validateEntropyBitSizeForSeed(bitSize int) error {
//	if (bitSize%8) != 0 || bitSize < 8 || bitSize > 64 {
//		return ErrInvalidEntropyLength
//	}
//	return nil
//}

// generateSeedWithRandomPassword 生成一个指定长度的随机数种子
func generateSeedWithRandomPassword(randomPassword []byte, keyLen int) []byte {
	salt := "jingbo is handsome."
	seed := pbkdf2.Key(randomPassword, []byte(salt), 2048, keyLen, sha512.New)

	return seed
}

// GenerateSeedWithStrengthAndKeyLen generates key seed with specified strength and length
func GenerateSeedWithStrengthAndKeyLen(strength int, keyLength int) ([]byte, error) {
	var entropyBitLength = 0
	//根据强度来判断随机数长度
	switch strength {
	case KeyStrengthEasy: // 弱
		entropyBitLength = 128
	case KeyStrengthMiddle: // 中
		entropyBitLength = 192
	case KeyStrengthHard: // 高
		entropyBitLength = 256
	default: // 不支持的语言类型
		entropyBitLength = 0
	}

	// 判断强度是否合法
	if entropyBitLength == 0 {
		return nil, ErrStrengthNotSupported
	}

	// 产生随机熵
	entropyByte, err := generateEntropy(entropyBitLength)
	if err != nil {
		return nil, err
	}

	return generateSeedWithRandomPassword(entropyByte, keyLength), nil
}
