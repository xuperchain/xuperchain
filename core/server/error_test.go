package server

import (
	"errors"
	"github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/pb"
	"testing"
)

func TestHandleBlockCoreError(t *testing.T) {

	testCases := map[string]struct {
		in       error
		expected pb.XChainErrorEnum
	}{
		"1": {
			in:       xchaincore.ErrCannotSyncBlock,
			expected: pb.XChainErrorEnum_CANNOT_SYNC_BLOCK_ERROR,
		},
		"2": {
			in:       xchaincore.ErrConfirmBlock,
			expected: pb.XChainErrorEnum_CONFIRM_BLOCK_ERROR,
		},
		"3": {
			in:       xchaincore.ErrUTXOVMPlay,
			expected: pb.XChainErrorEnum_UTXOVM_PLAY_ERROR,
		},
		"4": {
			in:       xchaincore.ErrWalk,
			expected: pb.XChainErrorEnum_WALK_ERROR,
		},
		"5": {
			in:       xchaincore.ErrNotReady,
			expected: pb.XChainErrorEnum_NOT_READY_ERROR,
		},
		"6": {
			in:       xchaincore.ErrBlockExist,
			expected: pb.XChainErrorEnum_BLOCK_EXIST_ERROR,
		},
		"7": {
			in:       xchaincore.ErrServiceRefused,
			expected: pb.XChainErrorEnum_SERVICE_REFUSED_ERROR,
		},
		"8": {
			in:       errors.New("default"),
			expected: pb.XChainErrorEnum_UNKNOW_ERROR,
		},
	}
	for testName, testCase := range testCases {
		if actual := HandleBlockCoreError(testCase.in); testCase.expected != actual {
			t.Errorf("%s expected: %v, actual: %v", testName, testCase.expected, actual)
		}
	}
}
