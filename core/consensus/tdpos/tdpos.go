//Copyright 2019 Baidu, Inc.

package tdpos

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/xuperchain/log15"

	"encoding/hex"
	"encoding/json"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/common/config"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	chainedbft "github.com/xuperchain/xuperunion/consensus/common/chainedbft"
	bft_config "github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/consensus/tdpos/bft"
	"github.com/xuperchain/xuperunion/contract"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/p2pv2"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo"
)

// Init init tdpos
func (tp *TDpos) Init() {
	tp.config = tDposConfig{
		initProposer: make(map[int64][]*cons_base.CandidateInfo),
	}
	tp.isProduce = make(map[int64]bool)
	tp.candidateBallots = new(sync.Map)
	tp.candidateBallotsCache = new(sync.Map)
	tp.revokeCache = new(sync.Map)
	tp.context = &contract.TxContext{}
	tp.mutex = new(sync.RWMutex)
}

// Type return the type of TDpos consensus
func (tp *TDpos) Type() string {
	return TYPE
}

// Version return the version of TDpos consensus
func (tp *TDpos) Version() int64 {
	return tp.version
}

// Configure is the specific implementation of ConsensusInterface
func (tp *TDpos) Configure(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{},
	extParams map[string]interface{}) error {
	if xlog == nil {
		xlog = log.New("module", "consensus")
		xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}
	address, err := ioutil.ReadFile(cfg.Miner.Keypath + "/address")
	if err != nil {
		xlog.Warn("load address error", "path", cfg.Miner.Keypath+"/address")
		return err
	}
	tp.log = xlog
	tp.address = address

	switch extParams["crypto_client"].(type) {
	case crypto_base.CryptoClient:
		tp.cryptoClient = extParams["crypto_client"].(crypto_base.CryptoClient)
	default:
		errMsg := "invalid type of crypto_client"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["ledger"].(type) {
	case *ledger.Ledger:
		tp.ledger = extParams["ledger"].(*ledger.Ledger)
	default:
		errMsg := "invalid type of ledger"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["utxovm"].(type) {
	case *utxo.UtxoVM:
		tp.utxoVM = extParams["utxovm"].(*utxo.UtxoVM)
	default:
		errMsg := "invalid type of utxovm"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["bcname"].(type) {
	case string:
		tp.bcname = extParams["bcname"].(string)
	default:
		errMsg := "invalid type of bcname"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	switch extParams["timestamp"].(type) {
	case int64:
		tp.initTimestamp = extParams["timestamp"].(int64)
	default:
		errMsg := "invalid type of timestamp"
		xlog.Warn(errMsg)
		return errors.New(errMsg)
	}

	if p2psvr, ok := extParams["p2psvr"].(p2pv2.P2PServer); ok {
		tp.p2psvr = p2psvr
	}

	if height, ok := extParams["height"].(int64); ok {
		tp.height = height
	}

	if err = tp.buildConfigs(xlog, nil, consCfg); err != nil {
		return err
	}

	if err = tp.initBFT(cfg); err != nil {
		xlog.Warn("init chained-bft failed!", "error", err)
		return err
	}

	if err = tp.initCandidateBallots(); err != nil {
		return err
	}
	tp.log.Trace("Configure", "TDpos", tp)
	return nil
}

func (tp *TDpos) buildConfigs(xlog log.Logger, cfg *config.NodeConfig, consCfg map[string]interface{}) error {
	// assemble consensus config
	if consCfg["proposer_num"] == nil {
		return errors.New("Parse TDpos proposer_num error, can not be null")
	}

	if consCfg["period"] == nil {
		return errors.New("Parse TDpos period error, can not be null")
	}

	if consCfg["alternate_interval"] == nil {
		return errors.New("Parse TDpos alternate_interval error, can not be null")
	}

	if consCfg["term_interval"] == nil {
		return errors.New("Parse TDpos term_interval error, can not be null")
	}

	if consCfg["vote_unit_price"] == nil {
		return errors.New("Parse TDpos vote_unit_price error, can not be null")
	}

	if consCfg["block_num"] == nil {
		return errors.New("Parse TDpos block_num error, can not be null")
	}

	if consCfg["init_proposer"] == nil {
		return errors.New("Parse TDpos init_proposer error, can not be null")
	}

	if consCfg["version"] == nil {
		tp.version = 0
	} else {
		version, err := strconv.ParseInt(consCfg["version"].(string), 10, 64)
		if err != nil {
			xlog.Warn("Parse TDpos config version error", "error", err.Error())
			return err
		}
		tp.version = version
	}

	proposerNum, err := strconv.ParseInt(consCfg["proposer_num"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse TDpos config error", "error", err.Error())
		return err
	}
	tp.config.proposerNum = proposerNum

	period, err := strconv.ParseInt(consCfg["period"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse TDpos config period error", "error", err.Error())
		return err
	}
	tp.config.period = period * 1e6

	alternateInterval, err := strconv.ParseInt(consCfg["alternate_interval"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse TDpos config alternateInterval error", "error", err.Error())
		return err
	}
	if alternateInterval%period != 0 {
		xlog.Warn("Parse TDpos config alternateInterval error", "error", "alternateInterval should be eliminated by period")
		return errors.New("alternateInterval should be eliminated by period")
	}
	tp.config.alternateInterval = alternateInterval * 1e6

	termInterval, err := strconv.ParseInt(consCfg["term_interval"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse TDpos config termInterval error", "error", err.Error())
		return err
	}
	if termInterval%period != 0 {
		xlog.Warn("Parse TDpos config termInterval error", "error", "termInterval should be eliminated by period")
		return errors.New("termInterval should be eliminated by period")
	}
	tp.config.termInterval = termInterval * 1e6

	voteUnitPrice := big.NewInt(0)
	if _, ok := voteUnitPrice.SetString(consCfg["vote_unit_price"].(string), 10); !ok {
		xlog.Warn("Parse TDpos config vote unit price error")
		return errors.New("Parse TDpos config vote unit price error")
	}
	tp.config.voteUnitPrice = voteUnitPrice

	blockNum, err := strconv.ParseInt(consCfg["block_num"].(string), 10, 64)
	if err != nil {
		xlog.Warn("Parse TDpos block_num period error", "error", err.Error())
		return err
	}
	tp.config.blockNum = blockNum

	// read config of need_neturl
	needNetURL := false
	if needNetURLVal, ok := consCfg["need_neturl"]; ok {
		needNetURL = needNetURLVal.(bool)
	}
	tp.config.needNetURL = needNetURL

	initProposer := consCfg["init_proposer"].(map[string]interface{})
	xlog.Trace("initProposer", "initProposer", initProposer)

	if len(initProposer) != 1 {
		xlog.Warn("TDpos init proposer length error", "length", len(initProposer))
		return errors.New("TDpos init proposer length error")
	}

	// first round proposers
	if _, ok := initProposer["1"]; !ok {
		return errors.New("TDpos init proposer error, Proposer 0 not provided")
	}
	initProposer1 := initProposer["1"].([]interface{})
	if int64(len(initProposer1)) != proposerNum {
		return errors.New("TDpos init proposer info error, Proposer 0 should be equal to proposerNum")
	}

	for _, v := range initProposer1 {
		canInfo := &cons_base.CandidateInfo{}
		canInfo.Address = v.(string)
		tp.config.initProposer[1] = append(tp.config.initProposer[1], canInfo)
	}

	// if have init_proposer_neturl, this info can be used for core peers connection
	if _, ok := consCfg["init_proposer_neturl"]; ok {
		proposerNeturls := consCfg["init_proposer_neturl"].(map[string]interface{})
		if _, ok := proposerNeturls["1"]; !ok {
			return errors.New("TDpos have init_proposer_neturl but don't have term 1")
		}
		proposerNeturls1 := proposerNeturls["1"].([]interface{})
		if int64(len(proposerNeturls1)) != proposerNum {
			return errors.New("TDpos init error, Proposer neturl number should be equal to proposerNum")
		}
		for idx, v := range proposerNeturls1 {
			tp.config.initProposer[1][idx].PeerAddr = v.(string)
			tp.log.Debug("TDpos proposer info", "index", idx, "proposer", tp.config.initProposer[1][idx])
		}
	} else {
		tp.log.Warn("TDpos have no neturl info for proposers",
			"need_neturl", needNetURL)
		if needNetURL {
			return errors.New("config error, init_proposer_neturl could not be empty")
		}
	}

	// parse bft related config
	tp.config.enableBFT = false
	if bftConfData, ok := consCfg["bft_config"].(map[string]interface{}); ok {
		bftconf := bft_config.MakeConfig(bftConfData)
		// if bft_config is not empty, enable bft
		tp.config.enableBFT = true
		tp.config.bftConfig = bftconf
	}

	tp.log.Trace("TDpos after config", "TTDpos.config", tp.config)
	return nil
}

func (tp *TDpos) initCandidateBallots() error {
	it := tp.utxoVM.ScanWithPrefix([]byte(GenCandidateBallotsPrefix()))
	defer it.Release()
	for it.Next() {
		key := string(it.Key())
		address, err := ParseCandidateBallotsKey(key)
		tp.log.Trace("initCandidateBallots", "key", key, "address", address)
		if err != nil {
			tp.log.Warn("initCandidateBallots parseCandidateBallotsKey error", "key", key)
			return err
		}
		ballots, err := strconv.ParseInt(string(it.Value()), 10, 64)
		tp.log.Trace("initCandidateBallots", "key", key, "address", address, "ballots", ballots)
		if err != nil {
			return err
		}
		tp.candidateBallots.Store(key, ballots)
	}
	return nil
}

// CompeteMaster is the specific implementation of ConsensusInterface
func (tp *TDpos) CompeteMaster(height int64) (bool, bool) {
	sentNewView := false
Again:
	t := time.Now()
	un := t.UnixNano()
	key := un / tp.config.period
	sleep := tp.config.period - un%tp.config.period
	maxsleeptime := time.Millisecond * 10
	if sleep > int64(maxsleeptime) {
		sleep = int64(maxsleeptime)
	}
	v, ok := tp.isProduce[key]
	if !ok || v == false {
		tp.isProduce[key] = true
	} else {
		time.Sleep(time.Duration(sleep))
		goto Again
	}
	// 查当前时间的term 和 pos
	t2 := time.Now()
	un2 := t2.UnixNano()
	term, pos, blockPos := tp.minerScheduling(un2)
	// 查当前term 和 pos是否是自己
	tp.curTerm = term
	if blockPos > tp.config.blockNum || pos >= tp.config.proposerNum {
		if !sentNewView {
			// only run once when term or proposer change
			err := tp.notifyNewView(height)
			if err != nil {
				tp.log.Warn("proposer or term change, bft Newview failed", "error", err)
			}
			sentNewView = true
		}
		goto Again
	}
	// reset proposers when term changed
	if pos == 0 && blockPos == 1 {
		err := tp.notifyTermChanged(tp.curTerm)
		if err != nil {
			tp.log.Warn("proposer or term change, bft Update Validators failed", "error", err)
		}
	}

	// if NewView not sent, send NewView message
	if !sentNewView {
		// if no term or proposer change, run NewView before generate block
		err := tp.notifyNewView(height)
		if err != nil {
			tp.log.Warn("proposer not changed, bft Newview failed", "error", err)
		}
		sentNewView = true
	}
	// master check
	if tp.isProposer(term, pos, tp.address) {
		tp.log.Trace("CompeteMaster now xterm infos", "term", term, "pos", pos, "blockPos", blockPos, "un2", un2,
			"master", true)
		tp.curBlockNum = blockPos
		s := tp.needSync()
		return true, s
	}
	tp.log.Trace("CompeteMaster now xterm infos", "term", term, "pos", pos, "blockPos", blockPos, "un2", un2,
		"master", false)
	return false, false
}

func (tp *TDpos) needSync() bool {
	meta := tp.ledger.GetMeta()
	if meta.TrunkHeight == 0 {
		return true
	}
	blockTip, err := tp.ledger.QueryBlock(meta.TipBlockid)
	if err != nil {
		return true
	}
	if string(blockTip.Proposer) == string(tp.address) {
		return false
	}
	return true
}

func (tp *TDpos) notifyNewView(height int64) error {
	if !tp.config.enableBFT {
		// BFT not enabled, continue
		return nil
	}

	// get current proposer
	meta := tp.ledger.GetMeta()
	proposer, err := tp.getProposer(0, 0)
	if err != nil {
		return err
	}
	if meta.TrunkHeight != 0 {
		blockTip, err := tp.ledger.QueryBlock(meta.TipBlockid)
		if err != nil {
			return err
		}
		proposer = string(blockTip.GetProposer())
	}

	// get next proposer
	// 查当前时间的term 和 pos
	t2 := time.Now()
	un2 := t2.UnixNano()
	term, pos, blockPos := tp.minerScheduling(un2)
	nextProposer, err := tp.getNextProposer(term, pos, blockPos)
	if err != nil {
		return err
	}
	// old height might out-of-date, use current trunkHeight when NewView
	return tp.bftPaceMaker.NextNewView(meta.TrunkHeight+1, nextProposer, proposer)
}

func (tp *TDpos) notifyTermChanged(term int64) error {
	if !tp.config.enableBFT {
		// BFT not enabled, continue
		return nil
	}

	proposers := tp.getTermProposer(term)
	return tp.bftPaceMaker.UpdateValidatorSet(proposers)
}

// CheckMinerMatch is the specific implementation of ConsensusInterface
func (tp *TDpos) CheckMinerMatch(header *pb.Header, in *pb.InternalBlock) (bool, error) {
	// 1 验证块信息是否合法
	blkid, err := ledger.MakeBlockID(in)
	if err != nil {
		tp.log.Warn("CheckMinerMatch MakeBlockID error", "logid", header.Logid, "error", err)
		return false, nil
	}
	if !(bytes.Equal(blkid, in.Blockid)) {
		tp.log.Warn("CheckMinerMatch equal blockid error", "logid", header.Logid, "redo blockid", global.F(blkid),
			"get blockid", global.F(in.Blockid))
		return false, nil
	}

	k, err := tp.cryptoClient.GetEcdsaPublicKeyFromJSON(in.Pubkey)
	if err != nil {
		tp.log.Warn("CheckMinerMatch get ecdsa from block error", "logid", header.Logid, "error", err)
		return false, nil
	}
	chkResult, _ := tp.cryptoClient.VerifyAddressUsingPublicKey(string(in.Proposer), k)
	if chkResult == false {
		tp.log.Warn("CheckMinerMatch address is not match publickey", "logid", header.Logid)
		return false, nil
	}

	valid, err := tp.cryptoClient.VerifyECDSA(k, in.Sign, in.Blockid)
	if err != nil || !valid {
		tp.log.Warn("CheckMinerMatch VerifyECDSA error", "logid", header.Logid, "error", err)
		return false, nil
	}

	if tp.config.enableBFT && !tp.isFirstblock(in.GetHeight()) {
		// if BFT enabled and it's not the first proposal
		// check whether previous block's QuorumCert is valid
		ok, err := tp.bftPaceMaker.GetChainedBFT().IsQuorumCertValidate(in.GetJustify())
		if err != nil || !ok {
			tp.log.Warn("CheckMinerMatch bft IsQuorumCertValidate failed", "logid", header.Logid, "error", err)
			return false, nil
		}
	}

	// 2 验证轮数信息
	preBlock, err := tp.ledger.QueryBlock(in.PreHash)
	if err != nil {
		tp.log.Warn("CheckMinerMatch failed, get preblock error")
		return false, nil
	}
	tp.log.Trace("CheckMinerMatch", "preBlock.CurTerm", preBlock.CurTerm, "in.CurTerm", in.CurTerm, " in.Proposer",
		string(in.Proposer), "blockid", fmt.Sprintf("%x", in.Blockid))
	term, pos, _ := tp.minerScheduling(in.Timestamp)
	if tp.isProposer(term, pos, in.Proposer) {
		// curTermProposerProduceNumCache is not thread safe, lock before use it.
		tp.mutex.Lock()
		defer tp.mutex.Unlock()
		// 当不是第一轮时需要和前面的
		if in.CurTerm != 1 {
			// 减少矿工50%概率恶意地输入时间
			if preBlock.CurTerm > term {
				tp.log.Warn("CheckMinerMatch failed, preBlock.CurTerm is bigger than this!")
				return false, nil
			}
			// 当系统切轮时初始化 curTermProposerProduceNum
			if preBlock.CurTerm < term || (tp.curTerm == term && tp.curTermProposerProduceNumCache == nil) {
				tp.curTermProposerProduceNumCache = make(map[int64]map[string]map[string]bool)
				tp.curTermProposerProduceNumCache[in.CurTerm] = make(map[string]map[string]bool)
			}
		}
		// 判断某个矿工是否恶意出块
		if tp.curTermProposerProduceNumCache != nil && tp.curTermProposerProduceNumCache[in.CurTerm] != nil {
			if _, ok := tp.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)]; !ok {
				tp.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)] = make(map[string]bool)
				tp.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)][hex.EncodeToString(in.Blockid)] = true
			} else {
				if !tp.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)][hex.EncodeToString(in.Blockid)] {
					tp.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)][hex.EncodeToString(in.Blockid)] = true
				}
			}
			if int64(len(tp.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)])) > tp.config.blockNum+1 {
				tp.log.Warn("CheckMinerMatch failed, proposer produce more than config blockNum!", "blockNum", len(tp.curTermProposerProduceNumCache[in.CurTerm][string(in.Proposer)]))
				return false, ErrProposeBlockMoreThanConfig
			}
		}
	} else {
		tp.log.Warn("CheckMinerMatch failed, revieved block shouldn't proposed!")
		return false, nil
	}
	return true, nil
}

// ProcessBeforeMiner is the specific implementation of ConsensusInterface
func (tp *TDpos) ProcessBeforeMiner(timestamp int64) (map[string]interface{}, bool) {
	res := make(map[string]interface{})
	term, pos, blockPos := tp.minerScheduling(timestamp)
	if term != tp.curTerm || blockPos > tp.config.blockNum || pos >= tp.config.proposerNum {
		return res, false
	}
	if !tp.isProposer(term, pos, tp.address) {
		tp.log.Warn("ProcessBeforeMiner prepare too long, omit!")
		return nil, false
	}

	// check bft status
	if tp.config.enableBFT {
		// TODO: what if IsLastViewConfirmed failed in competemaster, but succeed in ProcessBeforeMiner?
		if !tp.isFirstblock(tp.ledger.GetMeta().GetTrunkHeight() + 1) {
			if ok, _ := tp.bftPaceMaker.IsLastViewConfirmed(); !ok {
				tp.log.Warn("ProcessBeforeMiner last block not confirmed, walk to previous block")
				lastBlockid := tp.ledger.GetMeta().GetTipBlockid()
				lastBlock, err := tp.ledger.QueryBlock(lastBlockid)
				if err != nil {
					tp.log.Warn("ProcessBeforeMiner tip block query failed", "error", err)
					return nil, false
				}
				err = tp.utxoVM.Walk(lastBlock.GetPreHash(), false)
				if err != nil {
					tp.log.Warn("ProcessBeforeMiner utxo walk failed", "error", err)
					return nil, false
				}
				err = tp.ledger.Truncate(tp.utxoVM.GetLatestBlockid())
				if err != nil {
					tp.log.Warn("ProcessBeforeMiner ledger truncate failed", "error", err)
					return nil, false
				}
			}
		}

		qc, err := tp.bftPaceMaker.CurrentQCHigh([]byte(""))
		if err != nil {
			return nil, false
		}
		res["quorum_cert"] = qc
	}

	res["type"] = TYPE
	//res["curTerm"] = tp.curTerm
	//res["curBlockNum"] = tp.curBlockNum
	res["curTerm"] = term
	res["curBlockNum"] = blockPos
	tp.log.Trace("ProcessBeforeMiner", "res", res)
	return res, true
}

// ProcessConfirmBlock is the specific implementation of ConsensusInterface
func (tp *TDpos) ProcessConfirmBlock(block *pb.InternalBlock) error {
	// send bft NewProposal if bft enable and it's the miner
	if tp.config.enableBFT && bytes.Compare(block.GetProposer(), tp.address) == 0 {
		blockData := &pb.Block{
			Bcname:  tp.bcname,
			Blockid: block.Blockid,
			Block:   block,
		}

		err := tp.bftPaceMaker.NextNewProposal(block.Blockid, blockData)
		if err != nil {
			tp.log.Warn("ProcessConfirmBlock: bft next proposal failed", "error", err)
			return err
		}
	}
	return nil
}

// InitCurrent is the specific implementation of ConsensusInterface
func (tp *TDpos) InitCurrent(block *pb.InternalBlock) error {
	return nil
}

// Run is the specific implementation of interface contract
func (tp *TDpos) Run(desc *contract.TxDesc) error {
	switch desc.Method {
	// 进行投票
	case voteMethod:
		return tp.runVote(desc, tp.context.Block)
	case revokeVoteMethod:
		return tp.runRevokeVote(desc, tp.context.Block)
	case nominateCandidateMethod:
		return tp.runNominateCandidate(desc, tp.context.Block)
	case revokeCandidateMethod:
		return tp.runRevokeCandidate(desc, tp.context.Block)
	case checkvValidaterMethod:
		return tp.runCheckValidater(desc, tp.context.Block)
	default:
		tp.log.Warn("method not definated", "module", desc.Method, "method", desc.Method)
		return nil
	}
}

// Rollback is the specific implementation of interface contract
func (tp *TDpos) Rollback(desc *contract.TxDesc) error {
	switch desc.Method {
	// 回滚投票
	case voteMethod:
		return tp.rollbackVote(desc, tp.context.Block)
	case revokeVoteMethod:
		return tp.rollbackRevokeVote(desc, tp.context.Block)
	case nominateCandidateMethod:
		return tp.rollbackNominateCandidate(desc, tp.context.Block)
	case revokeCandidateMethod:
		return tp.rollbackRevokeCandidate(desc, tp.context.Block)
	case checkvValidaterMethod:
		return tp.rollbackCheckValidater(desc, tp.context.Block)
	default:
		tp.log.Warn("method not definated", "module", desc.Method, "method", desc.Method)
		return nil
	}
}

// Finalize is the specific implementation of interface contract
func (tp *TDpos) Finalize(blockid []byte) error {
	tp.candidateBallotsCache.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*candidateBallotsCacheValue)
		if value.isDel {
			tp.context.UtxoBatch.Delete([]byte(key))
			tp.candidateBallots.Delete(key)
		} else {
			tp.context.UtxoBatch.Put([]byte(key), []byte(strconv.FormatInt(value.ballots, 10)))
			tp.candidateBallots.Store(key, value.ballots)
		}
		return true
	})
	return nil
}

// SetContext is the specific implementation of interface contract
func (tp *TDpos) SetContext(context *contract.TxContext) error {
	tp.context = context
	tp.candidateBallotsCache = &sync.Map{}
	tp.revokeCache = &sync.Map{}
	return nil
}

// Stop is the specific implementation of interface contract
func (tp *TDpos) Stop() {
	if tp.config.enableBFT && tp.bftPaceMaker != nil {
		tp.bftPaceMaker.Stop()
	}
}

// ReadOutput is the specific implementation of interface contract
func (tp *TDpos) ReadOutput(desc *contract.TxDesc) (contract.ContractOutputInterface, error) {
	return nil, nil
}

// GetVerifiableAutogenTx is the specific implementation of interface VAT
func (tp *TDpos) GetVerifiableAutogenTx(blockHeight int64, maxCount int, timestamp int64) ([]*pb.Transaction, error) {
	term, _, _ := tp.minerScheduling(timestamp)

	key := GenTermCheckKey(tp.version, term+1)
	val, err := tp.utxoVM.GetFromTable(nil, []byte(key))
	txs := []*pb.Transaction{}
	if val == nil && common.NormalizedKVError(err) == common.ErrKVNotFound {
		desc := &contract.TxDesc{
			Module: "tdpos",
			Method: checkvValidaterMethod,
			Args:   make(map[string]interface{}),
		}
		desc.Args["version"] = strconv.FormatInt(tp.version, 10)
		desc.Args["term"] = strconv.FormatInt(term+1, 10)
		descJSON, err := json.Marshal(desc)
		if err != nil {
			return nil, err
		}
		tx, err := tp.utxoVM.GenerateEmptyTx(descJSON)
		txs = append(txs, tx)
		return txs, nil
	}
	return nil, nil
}

// GetVATWhiteList the specific implementation of interface VAT
func (tp *TDpos) GetVATWhiteList() map[string]bool {
	whiteList := map[string]bool{
		checkvValidaterMethod: true,
	}
	return whiteList
}

// GetCoreMiners get the information of core miners
func (tp *TDpos) GetCoreMiners() []*cons_base.MinerInfo {
	res := []*cons_base.MinerInfo{}
	timestamp := time.Now().UnixNano()
	term, _, _ := tp.minerScheduling(timestamp)
	proposers := tp.getTermProposer(term)
	for _, proposer := range proposers {
		minerInfo := &cons_base.MinerInfo{
			Address:  proposer.Address,
			PeerInfo: proposer.PeerAddr,
		}
		res = append(res, minerInfo)
	}
	return res
}

// GetStatus get the current status of consensus
func (tp *TDpos) GetStatus() *cons_base.ConsensusStatus {
	timestamp := time.Now().UnixNano()
	term, pos, blockPos := tp.minerScheduling(timestamp)
	proposers := tp.getTermProposer(term)
	status := &cons_base.ConsensusStatus{
		Term:     term,
		BlockNum: blockPos,
	}
	if int(pos) < 0 || int(pos) >= len(proposers) {
		tp.log.Warn("current pos illegal", "pos", pos)
	} else {
		status.Proposer = proposers[int(pos)].Address
	}
	return status
}

func (tp *TDpos) initBFT(cfg *config.NodeConfig) error {
	// BFT not enabled
	if !tp.config.enableBFT {
		return nil
	}

	// read keys
	pkpath := cfg.Miner.Keypath + "/public.key"
	pkJSON, err := ioutil.ReadFile(pkpath)
	if err != nil {
		tp.log.Warn("load private key error", "path", pkpath)
		return err
	}
	skpath := cfg.Miner.Keypath + "/private.key"
	skJSON, err := ioutil.ReadFile(skpath)
	if err != nil {
		tp.log.Warn("load private key error", "path", skpath)
		return err
	}
	sk, err := tp.cryptoClient.GetEcdsaPrivateKeyFromJSON(skJSON)
	if err != nil {
		tp.log.Warn("parse private key failed", "privateKey", skJSON)
		return err
	}

	// initialize bft
	bridge := bft.NewCbftBridge(tp.bcname, tp.ledger, tp.log, tp)
	qcNeeded := 3
	qc := make([]*pb.QuorumCert, qcNeeded)
	meta := tp.ledger.GetMeta()
	if meta.TrunkHeight != 0 {
		blockid := meta.TipBlockid
		for qcNeeded > 0 {
			qcNeeded--
			block, err := tp.ledger.QueryBlock(blockid)
			if err != nil {
				tp.log.Warn("initBFT: get block failed", "error", err, "blockid", string(blockid))
				return err
			}
			qc[qcNeeded] = block.GetJustify()
			blockid = block.GetPreHash()
			if blockid == nil {
				break
			}
		}
	}

	cbft, err := chainedbft.NewChainedBft(
		tp.log,
		tp.config.bftConfig,
		tp.bcname,
		string(tp.address),
		string(pkJSON),
		sk,
		tp.config.initProposer[1],
		bridge,
		tp.cryptoClient,
		tp.p2psvr,
		qc[2], qc[1], qc[0])

	if err != nil {
		tp.log.Warn("initBFT: create ChainedBft failed", "error", err)
		return err
	}

	paceMaker, err := bft.NewDPoSPaceMaker(tp.bcname, tp.height, meta.TrunkHeight,
		string(tp.address), cbft, tp.log, tp, tp.ledger)
	if err != nil {
		if err != nil {
			tp.log.Warn("initBFT: create DPoSPaceMaker failed", "error", err)
			return err
		}
	}
	tp.bftPaceMaker = paceMaker
	bridge.SetPaceMaker(paceMaker)
	return tp.bftPaceMaker.Start()
}

func (tp *TDpos) isFirstblock(targetHeight int64) bool {
	consStartHeight := tp.height
	consStartHeight++
	tp.log.Debug("isFirstblock check", "consStartHeight", consStartHeight,
		"targetHeight", targetHeight)
	return consStartHeight == targetHeight
}
