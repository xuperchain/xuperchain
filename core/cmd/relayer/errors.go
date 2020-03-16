package relayer

import (
	"errors"
	"strings"
)

// 常见的合约调用错误
var (
	ErrBlockHeaderTxMissingPreHash = errors.New("missing preHash")
	ErrBlockHeaderTxExist          = errors.New("existed already")
	ErrBlockHeaderTxOnlyOnce       = errors.New("only once")
)

// NormalizedBlockHeaderTxError normalize block header tx error
func NormalizedBlockHeaderTxError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "missing preHash") {
		return ErrBlockHeaderTxMissingPreHash
	}
	if strings.Contains(err.Error(), "existed already") {
		return ErrBlockHeaderTxExist
	}
	if strings.Contains(err.Error(), "only once") {
		return ErrBlockHeaderTxOnlyOnce
	}

	return err
}
