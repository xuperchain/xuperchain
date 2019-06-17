package ledger

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/pb"
)

// awardCacheSize system award cache, in avoid of double computing
const awardCacheSize = 1000

// RootConfig genesis block configure
type RootConfig struct {
	Version   string `json:"version"`
	Consensus struct {
		Type  string `json:"type"`
		Miner string `json:"miner"`
	} `json:"consensus"`
	Predistribution []struct {
		Address string `json:"address"`
		Quota   string `json:"quota"`
	}
	// max block size in MB
	MaxBlockSize string `json:"maxblocksize"`
	Period       string `json:"period"`
	Award        string `json:"award"`
	AwardDecay   struct {
		HeightGap int64   `json:"height_gap"`
		Ratio     float64 `json:"ratio"`
	} `json:"award_decay"`
	Decimals          string                 `json:"decimals"`
	GenesisConsensus  map[string]interface{} `json:"genesis_consensus"`
	ReservedContracts []string               `json:"reserved_contracts"`
}

// GetMaxBlockSizeInByte get max block size in Byte
func (rc *RootConfig) GetMaxBlockSizeInByte() (n int64) {
	maxSizeMB, _ := strconv.Atoi(rc.MaxBlockSize)
	n = int64(maxSizeMB) << 20
	return
}

// GetGenesisConsensus get consensus config of genesis block
func (rc *RootConfig) GetGenesisConsensus() (map[string]interface{}, error) {
	if rc.GenesisConsensus == nil {
		consCfg := map[string]interface{}{}
		consCfg["name"] = rc.Consensus.Type
		consCfg["config"] = map[string]interface{}{
			"miner":  rc.Consensus.Miner,
			"period": rc.Period,
		}
		return consCfg, nil
	}
	return rc.GenesisConsensus, nil
}

// GetReservedContract get default contract config of genesis block
func (rc *RootConfig) GetReservedContract() ([]string, error) {
	return rc.ReservedContracts, nil
}

// GenesisBlock genesis block data structure
type GenesisBlock struct {
	ib         *pb.InternalBlock
	config     *RootConfig
	awardCache *common.LRUCache
}

func getRootTx(ib *pb.InternalBlock) *pb.Transaction {
	for _, tx := range ib.Transactions {
		if tx.Coinbase {
			return tx
		}
	}
	return nil
}

// NewGenesisBlock new a genesis block
func NewGenesisBlock(ib *pb.InternalBlock) (*GenesisBlock, error) {
	gb := &GenesisBlock{
		awardCache: common.NewLRUCache(awardCacheSize),
	}
	gb.ib = ib
	config := &RootConfig{}
	rootTx := getRootTx(ib)
	if rootTx == nil {
		return nil, fmt.Errorf("genesis tx can not be found in the block: %x", ib.Blockid)
	}
	jsErr := json.Unmarshal(rootTx.Desc, config)
	if jsErr != nil {
		return nil, jsErr
	}
	gb.config = config
	return gb, nil
}

// GetInternalBlock returns internal block of genesis block
func (gb *GenesisBlock) GetInternalBlock() *pb.InternalBlock {
	return gb.ib
}

// GetConfig get config of genesis block
func (gb *GenesisBlock) GetConfig() *RootConfig {
	return gb.config
}

// CalcAward calc system award by block height
func (gb *GenesisBlock) CalcAward(blockHeight int64) *big.Int {
	award := big.NewInt(0)
	award.SetString(gb.config.Award, 10)
	if gb.config.AwardDecay.HeightGap == 0 { //无衰减策略
		return award
	}
	period := blockHeight / gb.config.AwardDecay.HeightGap
	if awardRemember, ok := gb.awardCache.Get(period); ok {
		return awardRemember.(*big.Int) //加个记忆，避免每次都重新算
	}
	var realAward = float64(award.Int64())
	for i := int64(0); i < period; i++ { //等比衰减
		realAward = realAward * gb.config.AwardDecay.Ratio
	}
	N := int64(math.Round(realAward)) //四舍五入
	award.SetInt64(N)
	gb.awardCache.Add(period, award)
	return award
}
