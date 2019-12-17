package poa

import (
	"errors"
	"github.com/xuperchain/xuperunion/consensus/poa/bft"
	"github.com/xuperchain/xuperunion/permission/acl/impl"
	"sync"

	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	bft_config "github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/p2pv2"
	"github.com/xuperchain/xuperunion/utxo"
)

var (
	// ErrProposerNotEnough proposer not enough
	ErrProposerNotEnough = errors.New("Term publish proposer num less than config")
	// ErrProposeBlockMoreThanConfig propose block more than config
	ErrProposeBlockMoreThanConfig = errors.New("Propose block more than config num error")
)

const (
	// TYPE the type of tdpos
	TYPE = "poa"
	// 验证人生成
	checkvValidaterMethod = "check_validater"
)

// TDpos is struct of tdpos consensus
type Poa struct {
	// tdpos共识配置
	config PoaConfig
	// tpos 版本信息, 要求是数字版本号, 避免由于用户指定字符版本导致取前缀有误
	version int64
	// tdpos start height, 共识起始高度
	height int64
	log    log.Logger
	// 共识作用的链名
	bcname string
	// 节点矿工address
	address []byte
	// 账本实例
	ledger *ledger.Ledger
	// utxo实例
	utxoVM *utxo.UtxoVM
	// 切换为tdpos的系统初始时间
	initTimestamp int64
	// 当前时间的轮数
	curTerm int64
	// 当前时间的候选人顺位
	curPos int64
	// 当前时间的块数
	curBlockNum int64
	// 验证者集合信息 address -> nodeInfo
	proposerInfos []*cons_base.CandidateInfo
	proposerNum int64
	// 记录某一轮内某个候选人出块是否大于系统限制, 以此避免矿工恶意出块, 切轮时进行初始化 map[term_num]map[proposer]map[blockid]bool
	curTermProposerProduceNumCache map[int64]map[string]map[string]bool
	// 此链使用的加密模块
	cryptoClient crypto_base.CryptoClient
	// revokeCache 撤销记录缓存, 内存状态, 记录每个block中撤销的记录, key: txid, value: true
	revokeCache *sync.Map
	isProduce   map[int64]bool

	mutex   *sync.RWMutex
	// BFT module
	bftPaceMaker *bft.PoaPaceMaker
	p2psvr       p2pv2.P2PServer
	// ACLManager
	accountName string
	aclManager impl.Manager
	// interval timer
	intervalT *MyTimer
}

// tdpos 共识机制的配置
type PoaConfig struct {
	// 出块间隔
	period int64
	// 更换候选人时间间隔
	alternateInterval int64
	// 每轮每个候选人最多出多少块
	blockNum int64
	// account name used for acl
	accountName string
	// initial proposers
	initProposer []*cons_base.CandidateInfo
	// is proposers' netURL needed for nomination and tdpos config
	// this is read from config need_neturl
	needNetURL bool
	// BTF related config
	enableBFT bool
	bftConfig *bft_config.Config
}