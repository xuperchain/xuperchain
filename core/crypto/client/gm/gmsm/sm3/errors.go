/*
Copyright Baidu Inc. All Rights Reserved.
*/

package sm3

import (
	"errors"
	"fmt"
)

var (
	// 熵的长度不在 [128, 256]以内或者长度不是32的倍数
	ErrInvalidEntropyLength = errors.New("Entropy length must within [128, 256] and be multiples of 32")

	// 助记词的强度暂未被支持
	// Strength required for generating Mnemonic not supported yet.
	ErrStrengthNotSupported = fmt.Errorf("This strength has not been supported yet.")

	// 助记词的语言类型暂未被支持
	// Language required for generating Mnemonic not supported yet.
	ErrLanguageNotSupported = fmt.Errorf("This language has not been supported yet.")

	// 助记词语句中包含的助记词的数量不合法，只能是12, 15, 18, 21, 24
	ErrMnemonicNumNotValid = fmt.Errorf("The number of words in the Mnemonic sentence is not valid. It must be within [12, 15, 18, 21, 24]")
)
