package event

import (
	"errors"
	"time"

	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
)

var _ Iterator = (*BlockIterator)(nil)

// BlockIterator wraps around ledger as a iterator style interface
type BlockIterator struct {
	currNum    int64
	endNum     int64
	blockStore BlockStore
	block      *pb.InternalBlock

	closed bool
	err    error
}

func NewBlockIterator(blockStore BlockStore, startNum, endNum int64) *BlockIterator {
	return &BlockIterator{
		currNum:    startNum,
		endNum:     endNum,
		blockStore: blockStore,
	}
}

func (b *BlockIterator) Next() bool {
	if b.closed && b.err != nil {
		return false
	}
	if b.endNum != -1 && b.currNum >= b.endNum {
		return false
	}

	block, err := b.fetchBlock(b.currNum)
	if err != nil {
		b.err = err
		return false
	}

	b.block = block
	b.currNum += 1
	return true
}

func (b *BlockIterator) fetchBlock(num int64) (*pb.InternalBlock, error) {
	for !b.closed {
		// 确保utxo更新到了对应的高度
		b.blockStore.WaitBlockHeight(num)
		block, err := b.blockStore.QueryBlockByHeight(num)
		if err == nil {
			return block, err
		}
		if err != ledger.ErrBlockNotExist {
			return nil, err
		}
		// TODO：utxo更新了，但账本找不到区块的情况，应该不会发生，发生了只能重试
		time.Sleep(time.Second)
	}
	return nil, errors.New("fetchBlock: code unreachable")
}

func (b *BlockIterator) Block() *pb.InternalBlock {
	return b.block
}

func (b *BlockIterator) Data() interface{} {
	return b.Block()
}

func (b *BlockIterator) Error() error {
	return b.err
}

func (b *BlockIterator) Close() {
	b.closed = true
}
