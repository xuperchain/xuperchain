package event

import (
	"encoding/hex"

	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

var _ Iterator = (*filteredBlockIterator)(nil)

type filteredBlockIterator struct {
	biter  *BlockIterator
	filter *blockFilter
	block  *pb.FilteredBlock

	endBlockNum int64

	closed bool
	err    error
}

func newFilteredBlockIterator(iter *BlockIterator, filter *blockFilter) *filteredBlockIterator {
	return &filteredBlockIterator{
		biter:  iter,
		filter: filter,
	}
}

func (b *filteredBlockIterator) Next() bool {
	if b.closed || b.err != nil {
		return false
	}
	var cont bool
	b.block, cont, b.err = b.fetchBlock()
	if b.err != nil {
		return false
	}
	return cont
}

func (b *filteredBlockIterator) matchTx(tx *pb.Transaction) bool {
	return matchTx(b.filter, tx)
}

func (b *filteredBlockIterator) toFilteredBlock(block *pb.InternalBlock) *pb.FilteredBlock {
	fblock := new(pb.FilteredBlock)
	fblock.Bcname = b.filter.GetBcname()
	fblock.Blockid = hex.EncodeToString(block.GetBlockid())
	fblock.BlockHeight = block.GetHeight()
	if b.filter.GetExcludeTx() {
		return fblock
	}

	hasEventFilter := hasEventFilter(b.filter)
	var txs []*pb.FilteredTransaction
	for _, tx := range block.GetTransactions() {
		if !b.matchTx(tx) {
			continue
		}
		events := b.parseFilteredEvents(tx)
		// 有合约事件过滤器并且当前交易没有匹配的事件，不区分交易没有合约事件或者事件都匹配
		// 则认为当前交易不符合过滤规则
		if len(events) == 0 && hasEventFilter {
			continue
		}
		ftx := &pb.FilteredTransaction{
			Txid:   hex.EncodeToString(tx.GetTxid()),
			Events: events,
		}

		txs = append(txs, ftx)
	}
	fblock.Txs = txs
	return fblock
}

func (b *filteredBlockIterator) parseFilteredEvents(tx *pb.Transaction) []*pb.ContractEvent {
	if b.filter.GetExcludeTxEvent() {
		return nil
	}
	events, err := xmodel.ParseContractEvents(tx)
	if err != nil {
		log.Error("parse contract event error", "txid", hex.EncodeToString(tx.GetTxid()), "error", err)
		return nil
	}

	var ret []*pb.ContractEvent
	for _, event := range events {
		if !matchEvent(b.filter, event) {
			continue
		}
		ret = append(ret, event)
	}

	return ret
}

func (b *filteredBlockIterator) fetchBlock() (*pb.FilteredBlock, bool, error) {
	for b.biter.Next() {
		block := b.biter.Block()
		filteredBlock := b.toFilteredBlock(block)
		return filteredBlock, true, nil
	}
	if b.biter.Error() != nil {
		return nil, false, b.biter.Error()
	}
	return nil, false, nil
}

func (b *filteredBlockIterator) Data() interface{} {
	return b.block
}

func (b *filteredBlockIterator) Error() error {
	return b.err
}

func (b *filteredBlockIterator) Close() {
	b.closed = true
	b.biter.Close()
}
