/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package server

import (
	"github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/pb"
)

// HandleBlockCoreError core error <=> pb.error
func HandleBlockCoreError(err error) pb.XChainErrorEnum {
	switch err {
	case xchaincore.ErrCannotSyncBlock:
		return pb.XChainErrorEnum_CANNOT_SYNC_BLOCK_ERROR
	case xchaincore.ErrConfirmBlock:
		return pb.XChainErrorEnum_CONFIRM_BLOCK_ERROR
	case xchaincore.ErrUTXOVMPlay:
		return pb.XChainErrorEnum_UTXOVM_PLAY_ERROR
	case xchaincore.ErrWalk:
		return pb.XChainErrorEnum_WALK_ERROR
	case xchaincore.ErrNotReady:
		return pb.XChainErrorEnum_NOT_READY_ERROR
	case xchaincore.ErrBlockExist:
		return pb.XChainErrorEnum_BLOCK_EXIST_ERROR
	case xchaincore.ErrServiceRefused:
		return pb.XChainErrorEnum_SERVICE_REFUSED_ERROR
	default:
		return pb.XChainErrorEnum_UNKNOW_ERROR
	}
}
