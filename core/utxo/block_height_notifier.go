package utxo

import "sync"

// BlockHeightNotifier hold the the latest block's height information of utxovm
// and notify listeners when information changed
type BlockHeightNotifier struct {
	mutex  sync.Mutex
	cond   *sync.Cond
	height int64
}

// NewBlockHeightNotifier instances a new BlockHeightNotifier
func NewBlockHeightNotifier() *BlockHeightNotifier {
	b := &BlockHeightNotifier{}
	b.cond = sync.NewCond(&b.mutex)
	return b
}

// UpdateHeight update the height information and notify all listeners
func (b *BlockHeightNotifier) UpdateHeight(height int64) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.height = height
	b.cond.Broadcast()
}

// WaitHeight wait util the height of current block >= target
func (b *BlockHeightNotifier) WaitHeight(target int64) int64 {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for b.height < target {
		b.cond.Wait()
	}
	return b.height
}
