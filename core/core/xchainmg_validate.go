package xchaincore

import (
	"errors"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/pb"

	"github.com/golang/protobuf/proto"
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
	// ErrTxInvalid is returned when tx invaild
	ErrTxInvalid = errors.New("validation error: tx info is invaild")
)

func validatePostTx(ts *pb.TxStatus) error {
	if ts == nil || ts.Tx == nil || len(ts.Txid) == 0 {
		return ErrTxNil
	}
	if len(ts.Bcname) == 0 {
		return ErrBlockChainNameEmpty
	}

	// 为了兼容pb和json序列化时，对于空byte数组的处理行为不同导致txid计算错误的问题
	// 先对输入参数统一做一次序列化，防止交易被打包入块，utxoVM校验不通过，阻塞walk
	// 可能会导致一些语言的sdk受影响，需要在计算txid时统一把空byte数组明确置null处理
	prtBuf, err := proto.Marshal(ts.Tx)
	if err != nil {
		return ErrTxInvalid
	}
	err = proto.Unmarshal(prtBuf, ts.Tx)
	if err != nil {
		return ErrTxInvalid
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
