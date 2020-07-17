package event

import (
	"fmt"

	xchaincore "github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
)

// ChainManager manage multiple block chain
type ChainManager interface {
	// GetBlockStore get BlockStore base bcname(the name of block chain)
	GetBlockStore(bcname string) (BlockStore, error)
}

// BlockStore is the interface of block store
type BlockStore interface {
	// TipBlockHeight returns the tip block height
	TipBlockHeight() (int64, error)
	// WaitBlockHeight wait until the height of current block height >= target
	WaitBlockHeight(target int64) int64
	// QueryBlockByHeight returns block at given height
	QueryBlockByHeight(int64) (*pb.InternalBlock, error)
}

type chainManager struct {
	chainmg *xchaincore.XChainMG
}

// NewChainManager returns ChainManager as the wrapper of xchaincore.XChainMG
func NewChainManager(chainmg *xchaincore.XChainMG) ChainManager {
	return &chainManager{
		chainmg: chainmg,
	}
}

func (c *chainManager) GetBlockStore(bcname string) (BlockStore, error) {
	chain := c.chainmg.Get(bcname)
	if chain == nil {
		return nil, fmt.Errorf("chain %s not found", bcname)
	}
	return NewBlockStore(chain.Ledger, chain.Utxovm), nil
}

type blockStore struct {
	*ledger.Ledger
	*utxo.UtxoVM
}

// NewBlockStore wraps ledger and utxovm as a BlockStore
func NewBlockStore(ledger *ledger.Ledger, utxovm *utxo.UtxoVM) BlockStore {
	return &blockStore{
		Ledger: ledger,
		UtxoVM: utxovm,
	}
}

func (b *blockStore) TipBlockHeight() (int64, error) {
	tipBlockid := b.Ledger.GetMeta().GetTipBlockid()
	block, err := b.Ledger.QueryBlockHeader(tipBlockid)
	if err != nil {
		return 0, err
	}
	return block.GetHeight(), nil
}
