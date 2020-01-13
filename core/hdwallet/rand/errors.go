/*
Copyright Baidu Inc. All Rights Reserved.
*/

package rand

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidRawEntropyLength 原始熵的长度不在 [120, 248]以内或者+8后的长度不是32的倍数
	ErrInvalidRawEntropyLength = errors.New("Entropy length must within [120, 248] and after +8 be multiples of 32")

	// ErrInvalidEntropyLength 熵的长度不在 [128, 256]以内或者长度不是32的倍数
	ErrInvalidEntropyLength = errors.New("Entropy length must within [128, 256] and be multiples of 32")

	// ErrStrengthNotSupported 助记词的强度暂未被支持
	// Strength required for generating Mnemonic not supported yet.
	ErrStrengthNotSupported = fmt.Errorf("This strength has not been supported yet")

	// ErrLanguageNotSupported 助记词的语言类型暂未被支持
	// Language required for generating Mnemonic not supported yet.
	ErrLanguageNotSupported = fmt.Errorf("This language has not been supported yet")

	// ErrMnemonicNumNotValid 助记词语句中包含的助记词的数量不合法，只能是12, 15, 18, 21, 24
	ErrMnemonicNumNotValid = fmt.Errorf("The number of words in the Mnemonic sentence is not valid. It must be within [12, 15, 18, 21, 24]")

	// ErrMnemonicChecksumIncorrect 助记词语句中包含的校验位的格式不合法
	ErrMnemonicChecksumIncorrect = errors.New("The checksum within the Mnemonic sentence incorrect")
)
