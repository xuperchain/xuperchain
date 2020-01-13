package xchaincore

import (
	"errors"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
	"testing"
)

func TestHandlerUtxoError(t *testing.T) {
	testCases := map[string]struct {
		in       error
		expected pb.XChainErrorEnum
	}{
		"1": {
			in:       utxo.ErrAlreadyInUnconfirmed,
			expected: pb.XChainErrorEnum_UTXOVM_ALREADY_UNCONFIRM_ERROR,
		},
		"2": {
			in:       utxo.ErrNoEnoughUTXO,
			expected: pb.XChainErrorEnum_NOT_ENOUGH_UTXO_ERROR,
		},
		"3": {
			in:       utxo.ErrUTXONotFound,
			expected: pb.XChainErrorEnum_UTXOVM_NOT_FOUND_ERROR,
		},
		"4": {
			in:       utxo.ErrInputOutputNotEqual,
			expected: pb.XChainErrorEnum_INPUT_OUTPUT_NOT_EQUAL_ERROR,
		},
		"5": {
			in:       utxo.ErrTxNotFound,
			expected: pb.XChainErrorEnum_TX_NOT_FOUND_ERROR,
		},
		"6": {
			in:       utxo.ErrTxSizeLimitExceeded,
			expected: pb.XChainErrorEnum_TX_SLE_ERROR,
		},
		"8": {
			in:       utxo.ErrRWSetInvalid,
			expected: pb.XChainErrorEnum_RWSET_INVALID_ERROR,
		},
		"9": {
			in:       errors.New("default"),
			expected: pb.XChainErrorEnum_UNKNOW_ERROR,
		},
	}
	for testName, testCase := range testCases {
		if actual := HandlerUtxoError(testCase.in); testCase.expected != actual {
			t.Errorf("%s expected: %v, actual: %v", testName, testCase.expected, actual)
		}
	}
}

func TestHandlerLedgerError(t *testing.T) {
	testCases := map[string]struct {
		in       error
		expected pb.XChainErrorEnum
	}{
		"1": {
			in:       ledger.ErrRootBlockAlreadyExist,
			expected: pb.XChainErrorEnum_ROOT_BLOCK_EXIST_ERROR,
		},
		"2": {
			in:       ledger.ErrTxDuplicated,
			expected: pb.XChainErrorEnum_TX_DUPLICATE_ERROR,
		},
		"3": {
			in:       ledger.ErrTxNotFound,
			expected: pb.XChainErrorEnum_TX_NOT_FOUND_ERROR,
		},
		"4": {
			in:       errors.New("default"),
			expected: pb.XChainErrorEnum_UNKNOW_ERROR,
		},
	}
	for testName, testCase := range testCases {
		if actual := HandlerLedgerError(testCase.in); testCase.expected != actual {
			t.Errorf("%s expected: %v, actual: %v", testName, testCase.expected, actual)
		}
	}

}
