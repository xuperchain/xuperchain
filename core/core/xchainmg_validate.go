package xchaincore

import (
	"errors"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/pb"
)

var (
	// ErrBlockChainNameEmpty is returned when blockchain name is empty
	ErrBlockChainNameEmpty = errors.New("validation error: validatePostTx TxStatus.Bcname can't be null")
	// ErrTxNil is returned when tx is nil
	ErrTxNil = errors.New("validation error: validatePostTx TxStatus.Tx can't be null")
	// ErrBlockIDNil is returned when blockid is nil
	ErrBlockIDNil = errors.New("validation error: validateSendBlock Block.Blockid can't be null")
	// ErrBlockNil is returned when block is nil
	ErrBlockNil = errors.New("validation error: validateSendBlock Block.Block can't be null")
)

func validatePostTx(ts *pb.TxStatus) error {
	if len(ts.Bcname) == 0 {
		return ErrBlockChainNameEmpty
	}

	if len(ts.Txid) == 0 || ts.Tx == nil {
		return ErrTxNil
	}
	return nil
}

func checkContractAuthority(contractWhiteList map[string]map[string]bool, tx *pb.Transaction) (bool, error) {
	descParse, err := contract.Parse(string(tx.GetDesc()))
	if err != nil {
		return true, nil
	}

	if contractWhiteList[descParse.Module] == nil {
		return true, nil
	}
	return tx.FromAddrInList(contractWhiteList[descParse.Module]), nil
}

func validateSendBlock(block *pb.Block) error {
	if len(block.Blockid) == 0 {
		return ErrBlockIDNil
	}

	if nil == block.Block {
		return ErrBlockNil
	}
	return nil
}
