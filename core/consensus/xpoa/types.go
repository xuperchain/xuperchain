package xpoa

import (
	"sync"

	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"

	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperchain/core/consensus/base"
	bft "github.com/xuperchain/xuperchain/core/consensus/common/chainedbft"
	bft_config "github.com/xuperchain/xuperchain/core/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/utxo"
)

const (
	// TYPE the type of xpoa
	TYPE = "xpoa"
)

// XPoa is struct of poa consensus
type XPoa struct {
	lg log.Logger
	// state 共识实例状态, SUSPEND|RUNNING
	state cons_base.ConsensusState
	// poa共识配置
	xpoaConf Config

	// 共识作用的链名
	bcname string
	// 节点矿工address
	address string
	// 账本实例
	ledger *ledger.Ledger
	// utxo实例
	utxoVM      *utxo.UtxoVM
	p2psvr      p2p_base.P2PServer
	mutex       *sync.RWMutex
	isProduce   map[int64]bool
	startHeight int64
	// 当前的验证集合 address -> nodeInfo
	proposerInfos []*cons_base.CandidateInfo
	// BFT module
	enableBFT    bool
	bftPaceMaker bft.PacemakerInterface
}

// Config xpoa共识机制的配置
type Config struct {
	// xpoa 版本信息, xpoa共识支持升级
	version int64
	// 出块间隔
	period int64
	// 每轮每个候选人最多出多少块
	blockNum int64
	// contractName name used for get validates
	contractName string
	// methodName
	methodName string
	// 切换为xpoa的初始时间
	initTimestamp int64
	// initial proposers
	initProposers []*cons_base.CandidateInfo
	// BTF related config
	bftConfig *bft_config.Config
}
