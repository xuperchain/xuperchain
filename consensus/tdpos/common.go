package tdpos

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/xuperchain/xuperunion/common"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/pb"
)

const minNominateProportion = 100000

// miner 调度算法, 依据时间进行矿工节点调度
func (tp *TDpos) minerScheduling(timestamp int64) (term int64, pos int64, blockPos int64) {
	if timestamp < tp.initTimestamp {
		return
	}
	tp.log.Trace("getTermPos", "timestamp", timestamp, "inittimestamp", tp.initTimestamp)
	// 每一轮的时间
	termTime := tp.config.termInterval + (tp.config.proposerNum-1)*tp.config.alternateInterval +
		tp.config.proposerNum*tp.config.period*(tp.config.blockNum-1)

	// 每个矿工轮值时间
	posTime := tp.config.alternateInterval + tp.config.period*(tp.config.blockNum-1)

	term = (timestamp-tp.initTimestamp)/termTime + 1
	resTime := (timestamp - tp.initTimestamp) - (term-1)*termTime
	pos = resTime / posTime
	resTime = resTime - (resTime/posTime)*posTime
	blockPos = resTime/tp.config.period + 1
	tp.log.Trace("getTermPos", "timestamp", timestamp, "term", term, "pos", pos, "blockPos", blockPos)
	return
}

// 调度产生的矿工与自身进行进行比较
func (tp *TDpos) isProposer(term int64, pos int64, address []byte) bool {
	if term == 0 {
		return false
	}
	proposers := tp.getTermProposer(term)
	tp.log.Trace("TDpos getTermProposer result", "term", term, "proposers", proposers)
	if proposers == nil {
		tp.log.Warn("TDpos getTermProposer error", "term", term)
		return false
	}
	if pos < 0 || pos > int64(len(proposers)-1) {
		tp.log.Warn("TDpos getTermProposer error, pos index out of range", "pos", pos, "proposers", proposers)
		return false
	}
	return string(address) == proposers[pos].Address
}

// 查询当前轮的验证者名单
func (tp *TDpos) getTermProposer(term int64) []*cons_base.CandidateInfo {
	if term == 1 {
		return tp.config.initProposer[1]
	}
	key := GenTermCheckKey(tp.version, term)
	val, err := tp.utxoVM.GetFromTable(nil, []byte(key))
	if err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound {
		tp.log.Error("TDpos getTermProposer vote result error", "term", term, "error", err)
		return nil
	} else if common.NormalizedKVError(err) == common.ErrKVNotFound {
		it := tp.utxoVM.ScanWithPrefix([]byte(genTermCheckKeyPrefix(tp.version)))
		if it.Last() {
			termLast, err := parseTermCheckKey(string(it.Key()))
			tp.log.Trace("TDpos getTermProposer ", "termLast", string(it.Key()))
			if err != nil {
				tp.log.Warn("TDpos getTermProposer parseTermCheckKey error", "error", err)
				return nil
			}
			if termLast == term+1 {
				if it.Prev() {
					tp.log.Trace("TDpos getTermProposer ", "key", string(it.Key()))
				} else {
					tp.log.Trace("TDpos getTermProposer parseTermCheckKey get prev nil")
					it.Last()
					keyLast := string(it.Key())
					it.First()
					ketFirst := string(it.Key())
					if keyLast == ketFirst {
						return tp.config.initProposer[1]
					}
					tp.log.Warn("TDpos getTermProposer parseTermCheckKey get prev error", "error", err)
					return nil
				}
			}
			val = it.Value()
		} else {
			tp.log.Warn("TDpos getTermProposer query from table is nil", "tp.config.initProposer[1]", tp.config.initProposer[1])
			return tp.config.initProposer[1]
		}
	}
	proposers := []*cons_base.CandidateInfo{}
	err = json.Unmarshal(val, &proposers)
	if err != nil {
		tp.log.Error("TDpos Unmarshal vote result error", "term", term, "error", err)
		return nil
	}
	return proposers

}

// 生成当前轮的验证者名单
func (tp *TDpos) genTermProposer() ([]*cons_base.CandidateInfo, error) {
	//var res []string
	var termBallotSli termBallotsSlice
	res := []*cons_base.CandidateInfo{}

	tp.candidateBallots.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(int64)
		tp.log.Trace("genTermProposer ", "key", key, "value", value)
		addr := strings.TrimPrefix(key, GenCandidateBallotsPrefix())
		if value == 0 {
			tp.log.Warn("genTermProposer continue", "key", key, "value", value)
			return true
		}
		tmp := &termBallots{
			Address: addr,
			Ballots: value,
		}
		termBallotSli = append(termBallotSli, tmp)
		tp.log.Trace("Term publish proposer num ", "tmp", tmp, "key", key)
		return true
	})

	if int64(termBallotSli.Len()) < tp.config.proposerNum {
		tp.log.Error("Term publish proposer num less than config", "termVotes", termBallotSli)
		return nil, ErrProposerNotEnough
	}

	sort.Stable(termBallotSli)
	for i := int64(0); i < tp.config.proposerNum; i++ {
		tp.log.Trace("genTermVote sort result", "address", termBallotSli[i].Address, "ballot", termBallotSli[i].Ballots)
		addr := termBallotSli[i].Address
		keyCanInfo := genCandidateInfoKey(addr)
		ciValue, err := tp.utxoVM.GetFromTable(nil, []byte(keyCanInfo))
		if err != nil {
			return nil, err
		}
		var canInfo *cons_base.CandidateInfo
		err = json.Unmarshal(ciValue, &canInfo)
		if err != nil {
			return nil, err
		}
		if canInfo.Address != addr {
			return nil, errors.New("candidate address not match vote address")
		}
		res = append(res, canInfo)
	}
	tp.log.Trace("genTermVote sort result", "result", res)
	return res, nil
}

// 计算tx中冻结高度为 -1 的amount 值
func calAmount(tx *pb.Transaction) (*big.Int, error) {
	sum := big.NewInt(0)
	for _, txOutput := range tx.TxOutputs {
		if txOutput.FrozenHeight == -1 {
			amount := big.NewInt(0)
			amount.SetBytes(txOutput.Amount)
			sum = sum.Add(sum, amount)
		}
	}
	return sum, nil
}

// 验证选票是否合法
func (tp *TDpos) validateVote(desc *contract.TxDesc) (*voteInfo, error) {
	voteInfo := &voteInfo{}
	// 解析冻结高度和冻结的amount
	amount, err := calAmount(desc.Tx)
	if err != nil {
		return nil, err
	}

	// 依据amount 计算票数
	ballotsBig := big.NewInt(0)
	ballotsBig = ballotsBig.Div(amount, tp.config.voteUnitPrice)
	voteInfo.ballots = ballotsBig.Int64()

	// 解析候选人信息
	if desc.Args["candidates"] == nil {
		return nil, errors.New("Vote candidates can not be null")
	}
	var candidates []interface{}
	switch desc.Args["candidates"].(type) {
	case []interface{}:
		candidates = desc.Args["candidates"].([]interface{})
	default:
		return nil, errors.New("candidates should be []interface{}")
	}
	for _, v := range candidates {
		voteInfo.candidates = append(voteInfo.candidates, v.(string))
	}
	voteInfo.candidates = common.UniqSlice(voteInfo.candidates)
	for i := 0; i < len(voteInfo.candidates); i++ {
		tp.log.Trace("validateVote inCandidate", "voteInfo.candidates", voteInfo.candidates[i], "i", i)
		ok := tp.inCandidate(voteInfo.candidates[i])
		if !ok {
			tp.log.Warn("The candidate not in candidates", "candidate", voteInfo.candidates[i])
			return nil, errors.New("The candidate not in candidates")
		}
	}

	if int64(len(voteInfo.candidates)) > tp.config.proposerNum {
		return nil, errors.New("candidates nums should less than proposer nums")
	}
	voteInfo.voter = string(desc.Tx.TxInputs[0].FromAddr)
	tp.log.Trace("validateVote success", "voteInfo", voteInfo)
	return voteInfo, nil
}

// 验证撤销合约参数是否合法
func (tp *TDpos) validRevoke(desc *contract.TxDesc) (*contract.TxDesc, error) {
	if desc.Args["txid"] == nil {
		return nil, errors.New("revoke candidate txid can not be null")
	}
	txid, ok := desc.Args["txid"].(string)
	if !ok {
		return nil, errors.New("candidates should be string")
	}
	hexTxid, err := hex.DecodeString(txid)
	tx, err := tp.ledger.QueryTransaction(hexTxid)

	if err != nil {
		tp.log.Warn("validRevoke query tx error", "txid", txid)
		return nil, err
	}
	descRev, err := contract.Parse(string(tx.Desc))
	if descRev != nil {
		descRev.Tx = tx
		return descRev, nil
	}
	return nil, errors.New("validRevoke descRev is nil")
}

// 验证撤销投票合约参数合法性
func (tp *TDpos) validateRevokeVote(desc *contract.TxDesc) (*voteInfo, string, error) {
	descRaw, err := tp.validRevoke(desc)
	if err != nil {
		tp.log.Warn("validRevoke error", "error", err)
		return nil, "", err
	}

	voteInfo, err := tp.validateVote(descRaw)
	if err != nil {
		return nil, "", err
	}
	return voteInfo, hex.EncodeToString(descRaw.Tx.Txid), nil
}

// 是否在候选人池中
func (tp *TDpos) inCandidate(candidate string) bool {
	keyBl := genCandidateBallotsKey(candidate)
	val, ok := tp.candidateBallotsCache.Load(keyBl)
	if ok {
		tp.log.Trace("inCandidate load candidateBallotsCache ok ", "val", val)
		blVal := val.(*candidateBallotsCacheValue)
		if blVal.isDel == true {
			return false
		}
		return true
	}
	res, ok := tp.candidateBallots.Load(keyBl)
	if !ok {
		tp.log.Trace("inCandidate load candidateBallots !ok ", "val", res)
		return false
	}
	return true
}

// 验证提名候选人合约参数是否合法
func (tp *TDpos) validateNominateCandidate(desc *contract.TxDesc) (*cons_base.CandidateInfo, string, error) {
	utxoTotal := tp.utxoVM.GetTotal()
	amount, err := calAmount(desc.Tx)
	if err != nil {
		return nil, "", err
	}
	// TODO: zq 多来源以后, 这里需要优化一下
	fromAddr := string(desc.Tx.TxInputs[0].FromAddr)
	canInfo := &cons_base.CandidateInfo{}

	utxoTotal.Div(utxoTotal, big.NewInt(minNominateProportion))
	if ok := amount.Cmp(utxoTotal) >= 0; !ok {
		tp.log.Warn("validateNominateCandidate error for amount not enough", "amount", amount.String(), "utxoNeed",
			utxoTotal.String())
		return nil, "", errors.New("validateNominateCandidate amount not enough")
	}

	// process candidate address
	if desc.Args["candidate"] == nil {
		return nil, "", errors.New("validateNominateCandidate candidate can not be null")
	}
	if candidate, ok := desc.Args["candidate"].(string); ok {
		if !checkCandidateName(candidate) {
			return nil, "", errors.New("validateNominateCandidate candidate name invalid")
		}
		canInfo.Address = candidate
	} else {
		return nil, "", errors.New("validateNominateCandidate candidates should be string")
	}

	// process candidate peerid
	if desc.Args["neturl"] == nil {
		tp.log.Warn("validateNominateCandidate candidate have no neturl info",
			"address", canInfo.Address)
		// neturl could not be empty when core peers' connection is enabled
		if tp.config.needNetURL {
			return nil, "", errors.New("validateNominateCandidate neturl could not be empty")
		}
	} else {
		if peerid, ok := desc.Args["neturl"].(string); ok {
			canInfo.PeerAddr = peerid
		} else {
			return nil, "", errors.New("validateNominateCandidate neturl should be string")
		}
	}

	return canInfo, fromAddr, nil
}

// 验证撤销候选人是否合法
func (tp *TDpos) validateRevokeCandidate(desc *contract.TxDesc) (string, string, string, error) {
	descNom, err := tp.validRevoke(desc)
	if err != nil {
		tp.log.Warn("validRevoke error", "error", err)
		return "", "", "", err
	}
	if descNom == nil || descNom.Module != "tdpos" || descNom.Method != nominateCandidateMethod {
		tp.log.Warn("validateRevokeCandidate error descNom not match", "descNom", descNom)
		return "", "", "", errors.New("validateRevokeCandidate error descNom not match")
	}

	if descNom.Args["candidate"] == nil {
		return "", "", "", errors.New("Vote candidate can not be null")
	}
	candidate, ok := descNom.Args["candidate"].(string)
	if !ok {
		return "", "", "", errors.New("candidates should be string")
	}
	fromAddr := string(descNom.Tx.TxInputs[0].FromAddr)
	return candidate, fromAddr, hex.EncodeToString(descNom.Tx.Txid), nil
}

// 验证检票参数是否合法
func (tp *TDpos) validateCheckValidater(desc *contract.TxDesc) (int64, int64, error) {
	if desc.Args["version"] == nil {
		return 0, 0, errors.New("validateCheckValidater error, args term can not be null")
	}

	if desc.Args["term"] == nil {
		return 0, 0, errors.New("validateCheckValidater error, args term can not be null")
	}
	version, err := strconv.ParseInt(desc.Args["version"].(string), 10, 64)
	if err != nil {
		tp.log.Warn("validateCheckValidater error", "version", version)
		return 0, 0, err
	}
	term, err := strconv.ParseInt(desc.Args["term"].(string), 10, 64)
	if err != nil {
		tp.log.Warn("validateCheckValidater error", "term", term)
		return 0, 0, err
	}
	return version, term, nil
}

func (tp *TDpos) isAuthAddress(candidate string, initiator string, authRequire []string) bool {
	if strings.HasSuffix(initiator, candidate) {
		return true
	}
	for _, value := range authRequire {
		if strings.HasSuffix(value, candidate) {
			return true
		}
	}
	return false
}
