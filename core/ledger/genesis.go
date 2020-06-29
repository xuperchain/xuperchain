package ledger

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/pb"
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
	NoFee        bool   `json:"nofee"`
	Award        string `json:"award"`
	AwardDecay   struct {
		HeightGap int64   `json:"height_gap"`
		Ratio     float64 `json:"ratio"`
	} `json:"award_decay"`
	GasPrice struct {
		CpuRate  int64 `json:"cpu_rate"`
		MemRate  int64 `json:"mem_rate"`
		DiskRate int64 `json:"disk_rate"`
		XfeeRate int64 `json:"xfee_rate"`
	} `json:"gas_price"`
	Decimals          string                 `json:"decimals"`
	GenesisConsensus  map[string]interface{} `json:"genesis_consensus"`
	ReservedContracts []InvokeRequest        `json:"reserved_contracts"`
	ReservedWhitelist struct {
		Account string `json:"account"`
	} `json:"reserved_whitelist"`
	ForbiddenContract InvokeRequest `json:"forbidden_contract"`
	// NewAccountResourceAmount the amount of creating a new contract account
	NewAccountResourceAmount int64 `json:"new_account_resource_amount"`
	// IrreversibleSlideWindow
	IrreversibleSlideWindow string `json:"irreversibleslidewindow"`
	// GroupChainContract
	GroupChainContract InvokeRequest `json:"group_chain_contract"`
}

// GasPrice define gas rate for utxo
type GasPrice struct {
	CpuRate  int64 `json:"cpu_rate" mapstructure:"cpu_rate"`
	MemRate  int64 `json:"mem_rate" mapstructure:"mem_rate"`
	DiskRate int64 `json:"disk_rate" mapstructure:"disk_rate"`
	XfeeRate int64 `json:"xfee_rate" mapstructure:"xfee_rate"`
}

// InvokeRequest define genesis reserved_contracts configure
type InvokeRequest struct {
	ModuleName   string            `json:"module_name" mapstructure:"module_name"`
	ContractName string            `json:"contract_name" mapstructure:"contract_name"`
	MethodName   string            `json:"method_name" mapstructure:"method_name"`
	Args         map[string]string `json:"args" mapstructure:"args"`
}

func InvokeRequestFromJSON2Pb(jsonRequest []InvokeRequest) ([]*pb.InvokeRequest, error) {
	requestsWithPb := []*pb.InvokeRequest{}
	for _, request := range jsonRequest {
		tmpReqWithPB := &pb.InvokeRequest{
			ModuleName:   request.ModuleName,
			ContractName: request.ContractName,
			MethodName:   request.MethodName,
			Args:         make(map[string][]byte),
		}
		for k, v := range request.Args {
			tmpReqWithPB.Args[k] = []byte(v)
		}
		requestsWithPb = append(requestsWithPb, tmpReqWithPB)
	}
	return requestsWithPb, nil
}

// GetIrreversibleSlideWindow get irreversible slide window
func (rc *RootConfig) GetIrreversibleSlideWindow() int64 {
	irreversibleSlideWindow, _ := strconv.Atoi(rc.IrreversibleSlideWindow)
	return int64(irreversibleSlideWindow)
}

// GetMaxBlockSizeInByte get max block size in Byte
func (rc *RootConfig) GetMaxBlockSizeInByte() (n int64) {
	maxSizeMB, _ := strconv.Atoi(rc.MaxBlockSize)
	n = int64(maxSizeMB) << 20
	return
}

// GetNewAccountResourceAmount get the resource amount of new an account
func (rc *RootConfig) GetNewAccountResourceAmount() int64 {
	return rc.NewAccountResourceAmount
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
func (rc *RootConfig) GetReservedContract() ([]*pb.InvokeRequest, error) {
	return InvokeRequestFromJSON2Pb(rc.ReservedContracts)
}

func (rc *RootConfig) GetForbiddenContract() ([]*pb.InvokeRequest, error) {
	return InvokeRequestFromJSON2Pb([]InvokeRequest{rc.ForbiddenContract})
}

func (rc *RootConfig) GetGroupChainContract() ([]*pb.InvokeRequest, error) {
	return InvokeRequestFromJSON2Pb([]InvokeRequest{rc.GroupChainContract})
}

// GetReservedWhitelistAccount return reserved whitelist account
func (rc *RootConfig) GetReservedWhitelistAccount() string {
	return rc.ReservedWhitelist.Account
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
	if config.NoFee {
		config.Award = "0"
		config.NewAccountResourceAmount = 0
		config.Predistribution = []struct {
			Address string `json:"address"`
			Quota   string `json:"quota"`
		}{}
		config.GasPrice.CpuRate = 0
		config.GasPrice.DiskRate = 0
		config.GasPrice.MemRate = 0
		config.GasPrice.XfeeRate = 0
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

// GetGasPrice get gas rate for different resource(cpu, mem, disk and xfee)
func (rc *RootConfig) GetGasPrice() *pb.GasPrice {
	gasPrice := &pb.GasPrice{
		CpuRate:  rc.GasPrice.CpuRate,
		MemRate:  rc.GasPrice.MemRate,
		DiskRate: rc.GasPrice.DiskRate,
		XfeeRate: rc.GasPrice.XfeeRate,
	}
	return gasPrice
}
