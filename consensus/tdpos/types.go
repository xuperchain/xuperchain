package tdpos

import (
	"errors"
	"math/big"
	"sync"

	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/contract"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/ledger"
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
	TYPE = "tdpos"
	// 候选人投票
	voteMethod = "vote"
	// 候选人投票撤销
	revokeVoteMethod = "revoke_vote"
	// 候选人提名
	nominateCandidateMethod = "nominate_candidate"
	// 候选人罢黜
	revokeCandidateMethod = "revoke_candidate"
	// 验证人生成
	checkvValidaterMethod = "check_validater"
)

// TDpos is struct of tdpos consensus
type TDpos struct {
	// tdpos共识配置
	config tDposConfig
	// tpos 版本信息, 要求是数字版本号, 避免由于用户指定字符版本导致取前缀有误
	version int64
	log     log.Logger
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
	// 当前时间的块数
	curBlockNum int64
	// 候选人得票信息 key: address, value: ballots
	candidateBallots *sync.Map
	// candidate投票缓存, 内存状态, 记录每个block中得票数的变化, key: address, value: {ballot, isDel}, 以块为粒度, 块提交后清除
	candidateBallotsCache *sync.Map
	// 记录某一轮内某个候选人出块是否大于系统限制, 以此避免矿工恶意出块, 切轮时进行初始化 map[term_num]map[proposer]map[blockid]bool
	curTermProposerProduceNumCache map[int64]map[string]map[string]bool
	// 此链使用的加密模块
	cryptoClient crypto_base.CryptoClient
	// revokeCache 撤销记录缓存, 内存状态, 记录每个block中撤销的记录, key: txid, value: true
	revokeCache *sync.Map
	isProduce   map[int64]bool
	// 执行智能合约获取合约上下文
	context *contract.TxContext
	mutex   *sync.RWMutex
}

// tdpos 共识机制的配置
type tDposConfig struct {
	// 每轮选出的候选人个数
	proposerNum int64
	// 出块间隔
	period int64
	// 更换候选人时间间隔
	alternateInterval int64
	// 更换轮时间间隔
	termInterval int64
	// 每轮每个候选人最多出多少块
	blockNum int64
	// 投票单价
	voteUnitPrice *big.Int
	// 系统指定的前两轮的候选人名单
	initProposer map[int64][]*cons_base.CandidateInfo
	// is proposers' netURL needed for nomination and tdpos config
	// this is read from config need_neturl
	needNetURL bool
}

// 每个选票的详情, 支持一票多投
type voteInfo struct {
	// 每个选票投给的address名单, 最多不能超过proposerNum的限制
	candidates []string
	// 每一轮投多少票, 依据总的amount计算得到
	// ballots = 总金额 / 投票单价
	ballots int64
	voter   string
}

// 每个地址每一轮的总票数
type termBallots struct {
	Address string
	Ballots int64
}

type termBallotsSlice []*termBallots

func (tv termBallotsSlice) Len() int {
	return len(tv)
}

func (tv termBallotsSlice) Swap(i, j int) {
	tv[i], tv[j] = tv[j], tv[i]
}

func (tv termBallotsSlice) Less(i, j int) bool {
	if tv[j].Ballots == tv[i].Ballots {
		return tv[j].Address < tv[i].Address
	}
	return tv[j].Ballots < tv[i].Ballots
}

// candidateBallotsCacheValue
type candidateBallotsCacheValue struct {
	ballots int64
	// 是否被标记为删除
	isDel bool
}
