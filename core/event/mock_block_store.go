package event

import (
	"errors"
	"sync"

	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
)

type mockBlockStore struct {
	mutex  sync.Mutex
	blocks []*pb.InternalBlock

	heightNotifier *utxo.BlockHeightNotifier
}

func newMockBlockStore() *mockBlockStore {
	return &mockBlockStore{
		heightNotifier: utxo.NewBlockHeightNotifier(),
	}
}

// TipBlockHeight returns the tip block height
func (m *mockBlockStore) TipBlockHeight() (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return int64(len(m.blocks)), nil
}

// WaitBlockHeight wait until the height of current block height >= target
func (m *mockBlockStore) WaitBlockHeight(target int64) int64 {
	return m.heightNotifier.WaitHeight(target)
}

// QueryBlockByHeight returns block at given height
func (m *mockBlockStore) QueryBlockByHeight(height int64) (*pb.InternalBlock, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if height < 0 {
		return nil, errors.New("bad height")
	}
	if height >= int64(len(m.blocks)) {
		return nil, ledger.ErrBlockNotExist
	}
	return m.blocks[int(height)], nil
}

func (m *mockBlockStore) AppendBlock(block *pb.InternalBlock) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	nblock := *block
	nblock.Height = int64(len(m.blocks))
	m.blocks = append(m.blocks, &nblock)
	m.heightNotifier.UpdateHeight(nblock.Height)
}

// GetBlockStore get BlockStore base bcname(the name of block chain)
func (m *mockBlockStore) GetBlockStore(bcname string) (BlockStore, error) {
	return m, nil
}
