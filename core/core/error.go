package xchaincore

import (
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
)

// HandlerUtxoError used to handle error of utxo
func HandlerUtxoError(err error) pb.XChainErrorEnum {
	switch err {
	case utxo.ErrAlreadyInUnconfirmed:
		return pb.XChainErrorEnum_UTXOVM_ALREADY_UNCONFIRM_ERROR
	case utxo.ErrNoEnoughUTXO:
		return pb.XChainErrorEnum_NOT_ENOUGH_UTXO_ERROR
	case utxo.ErrUTXONotFound:
		return pb.XChainErrorEnum_UTXOVM_NOT_FOUND_ERROR
	case utxo.ErrInputOutputNotEqual:
		return pb.XChainErrorEnum_INPUT_OUTPUT_NOT_EQUAL_ERROR
	case utxo.ErrTxNotFound:
		return pb.XChainErrorEnum_TX_NOT_FOUND_ERROR
	case utxo.ErrTxSizeLimitExceeded:
		return pb.XChainErrorEnum_TX_SLE_ERROR
	case utxo.ErrRWSetInvalid:
		return pb.XChainErrorEnum_RWSET_INVALID_ERROR
	default:
		return pb.XChainErrorEnum_UNKNOW_ERROR
	}
}

// HandlerLedgerError used to handle error of ledger
func HandlerLedgerError(err error) pb.XChainErrorEnum {
	switch err {
	case ledger.ErrRootBlockAlreadyExist:
		return pb.XChainErrorEnum_ROOT_BLOCK_EXIST_ERROR
	case ledger.ErrTxDuplicated:
		return pb.XChainErrorEnum_TX_DUPLICATE_ERROR
	case ledger.ErrTxNotFound:
		return pb.XChainErrorEnum_TX_NOT_FOUND_ERROR
	default:
		return pb.XChainErrorEnum_UNKNOW_ERROR
	}
}
