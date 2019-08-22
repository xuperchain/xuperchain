package tdpos

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/common/events"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/pb"
)

// 执行投票
func (tp *TDpos) runVote(desc *contract.TxDesc, block *pb.InternalBlock) error {
	// 验证选票信息有效性, 并解析选票参数
	tp.log.Trace("start to runVote desc", "desc", desc)
	voteInfo, err := tp.validateVote(desc)
	if err != nil {
		tp.log.Warn("runVote error", "error", err)
		return err
	}

	for i := 0; i < len(voteInfo.candidates); i++ {
		keyCanBal := genCandidateBallotsKey(voteInfo.candidates[i])
		keyCandidateVote := genCandidateVoteKey(voteInfo.candidates[i], voteInfo.voter, hex.EncodeToString(desc.Tx.Txid))
		tp.log.Trace("runVote genCandidateVoteKey", "keyCandidateVote", keyCandidateVote)
		keyVoteCandidate := genVoteCandidateKey(voteInfo.voter, voteInfo.candidates[i], hex.EncodeToString(desc.Tx.Txid))
		tp.log.Trace("runVote genVoteCandidateKey", "genVoteCandidateKey", keyCandidateVote)
		// 先看一下缓存里有没有,有的话则直接处理缓存
		if val, ok := tp.candidateBallotsCache.Load(keyCanBal); ok {
			tp.log.Trace("runVote get from cache ok", "val", val)
			canBal := val.(*candidateBallotsCacheValue)
			if !canBal.isDel {
				tp.log.Trace("runVote add ballots before cal", "ballots", canBal.ballots)
				canBal.ballots += voteInfo.ballots
				tp.log.Trace("runVote add ballots after cal", "ballots", canBal.ballots)
				tp.candidateBallotsCache.Store(keyCanBal, canBal)
			} else {
				tp.log.Warn("runVote error", "error", "the candidate was revoked!")
				return errors.New("runVote error the candidate was revoked")
			}
		} else {
			// 尝试从内存里load出来再进行记票
			if bal, ok := tp.candidateBallots.Load(keyCanBal); ok {
				tp.log.Trace("runVote get from men ok", "val", bal)
				tp.log.Trace("runVote add ballots before cal", "ballots", bal.(int64))
				bals := bal.(int64) + voteInfo.ballots
				tp.log.Trace("runVote add ballots after cal", "ballots", bals)
				canBal := &candidateBallotsCacheValue{
					ballots: bals,
					isDel:   false,
				}
				tp.candidateBallotsCache.Store(keyCanBal, canBal)
			} else {
				// 候选人不在内存中, 说明已经被删除了
				tp.log.Warn("runVote error", "error", "the candidate not found!")
				return errors.New("runVote error, the candidate not found")
			}
		}
		// 记录某个候选人被谁投了票
		tp.context.UtxoBatch.Put([]byte(keyCandidateVote), []byte(strconv.FormatInt(voteInfo.ballots, 10)))
		// 记录了某个人给谁投了票
		tp.context.UtxoBatch.Put([]byte(keyVoteCandidate), []byte(strconv.FormatInt(voteInfo.ballots, 10)))
	}
	return nil
}

// 执行撤销投票
func (tp *TDpos) runRevokeVote(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("start to runRevokeVote desc", "desc", desc)
	voteInfo, txVote, err := tp.validateRevokeVote(desc)
	if err != nil {
		tp.log.Warn("runRevokeVote error", "error", err)
		return err
	}

	keyRevoke := genRevokeKey(txVote)
	if _, ok := tp.revokeCache.Load(txVote); ok {
		tp.log.Warn("runRevokeVote error", "error", "revoke repeated")
		return errors.New("runRevokeVote error revoke repeated")
	}
	val, err := tp.utxoVM.GetFromTable(nil, []byte(keyRevoke))
	if (err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound) || val != nil {
		tp.log.Warn("runRevokeVote error revoke repeated or get revoke key from db error", "val", hex.EncodeToString(val),
			"error", err)
		return errors.New("runRevokeVote error revoke repeated or get revoke key from db error")
	}

	for i := 0; i < len(voteInfo.candidates); i++ {
		keyCanBal := genCandidateBallotsKey(voteInfo.candidates[i])
		keyCandidateVote := genCandidateVoteKey(voteInfo.candidates[i], voteInfo.voter, txVote)
		tp.log.Trace("runRevokeVote genCandidateVoteKey", "keyCandidateVote", keyCandidateVote)
		keyVoteCandidate := genVoteCandidateKey(voteInfo.voter, voteInfo.candidates[i], txVote)
		tp.log.Trace("runRevokeVote genVoteCandidateKey", "genVoteCandidateKey", keyCandidateVote)
		// 先看一下缓存里有没有,有的话则直接处理缓存
		if val, ok := tp.candidateBallotsCache.Load(keyCanBal); ok {
			tp.log.Trace("runRevokeVote get from cache ok", "val", val)
			canBal := val.(*candidateBallotsCacheValue)
			if !canBal.isDel {
				tp.log.Trace("runRevokeVote minus ballots before cal", "ballots", canBal.ballots)
				canBal.ballots -= voteInfo.ballots
				tp.log.Trace("runRevokeVote minus ballots after cal", "ballots", canBal.ballots)
				tp.candidateBallotsCache.Store(keyCanBal, canBal)
			} else {
				tp.log.Warn("runRevokeVote error", "error", "the candidate was revoked!")
				return errors.New("runRevokeVote error, the candidate was revoked")
			}
		} else {
			// 尝试从内存里load出来再进行票撤销
			if bal, ok := tp.candidateBallots.Load(keyCanBal); ok {
				tp.log.Trace("runRevokeVote get from men ok", "val", bal)
				tp.log.Trace("runRevokeVote add ballots before cal", "ballots", bal.(int64))
				bals := bal.(int64) - voteInfo.ballots
				tp.log.Trace("runRevokeVote add ballots after cal", "ballots", bals)
				canBal := &candidateBallotsCacheValue{
					ballots: bals,
					isDel:   false,
				}
				tp.candidateBallotsCache.Store(keyCanBal, canBal)
			} else {
				// 候选人不在内存中, 说明已经被删除了
				tp.log.Warn("runRevokeVote error", "error", "the candidate not found!")
				return errors.New("runRevokeVote error, the candidate not found")
			}
		}
		// 清除某个候选人被谁投了票的记录
		tp.context.UtxoBatch.Delete([]byte(keyCandidateVote))
		// 清除某个人给谁投了票的记录
		tp.context.UtxoBatch.Delete([]byte(keyVoteCandidate))
		// 记录撤销记录
		tp.revokeCache.Store(keyRevoke, true)
		tp.context.UtxoBatch.Put([]byte(keyRevoke), desc.Tx.Txid)
	}
	return nil
}

// 执行提名候选人
func (tp *TDpos) runNominateCandidate(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("start to runNominateCandidate", "desc", desc)
	canInfo, fromAddr, err := tp.validateNominateCandidate(desc)
	if err != nil {
		tp.log.Warn("run to validate nominate error", "error", err.Error())
		return err
	}
	candidate := canInfo.Address
	key := GenCandidateNominateKey(candidate)
	keyCanBal := genCandidateBallotsKey(candidate)
	keyCanInfo := genCandidateInfoKey(candidate)
	keyNominateRecord := GenNominateRecordsKey(fromAddr, candidate, hex.EncodeToString(desc.Tx.Txid))

	canInfoValue, err := json.Marshal(canInfo)
	if err != nil {
		tp.log.Warn("runNominateCandidate json marshal failed", "err", err)
		return err
	}

	// 判断内存中是否已经提过名
	val, ok := tp.candidateBallotsCache.Load(keyCanBal)
	if ok {
		tp.log.Trace("runNominateCandidate get from cache ok", "val", val)
		canBal := val.(*candidateBallotsCacheValue)
		if !canBal.isDel {
			tp.log.Warn("runNominateCandidate this candidate had been nominate!")
			return errors.New("runNominateCandidate this candidate had been nominate!")
		}
		tp.log.Trace("runNominateCandidate recover candidate!", "key", keyCanBal)
		canBal.isDel = false
		canBal.ballots = 0
		tp.candidateBallotsCache.Store(keyCanBal, canBal)
		tp.context.UtxoBatch.Put([]byte(key), desc.Tx.Txid)
		tp.context.UtxoBatch.Put([]byte(keyNominateRecord), desc.Tx.Txid)
		tp.context.UtxoBatch.Put([]byte(keyCanInfo), []byte(canInfoValue))
		return nil

	}
	// 从内存中load出该候选人的记录
	_, ok = tp.candidateBallots.Load(keyCanBal)
	if !ok {
		// check if the address nominated exists in the initiator or slice of auth_require
		initiator := desc.Tx.GetInitiator()
		authRequire := desc.Tx.GetAuthRequire()
		if ok := tp.isAuthAddress(candidate, initiator, authRequire); !ok {
			tp.log.Warn("candidate has not been authenticated", "candidate:", candidate)
			return errors.New("candidate has not been authenticated")
		}
		tp.log.Trace("runNominateCandidate candidate!", "key", keyCanBal)
		// 如果内存中没有, 则说明该候选人可以被提名并进行提名
		canBal := &candidateBallotsCacheValue{}
		canBal.isDel = false
		canBal.ballots = 0
		tp.candidateBallotsCache.Store(keyCanBal, canBal)
		tp.context.UtxoBatch.Put([]byte(key), desc.Tx.Txid)
		tp.context.UtxoBatch.Put([]byte(keyNominateRecord), desc.Tx.Txid)
		tp.context.UtxoBatch.Put([]byte(keyCanInfo), []byte(canInfoValue))
		return nil
	}
	// 内存中已经存在了, 说明被重复提名
	tp.log.Warn("This candidate had been nominate!")
	return nil
}

// 执行候选人撤销
func (tp *TDpos) runRevokeCandidate(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("start to runRevokeCandidate", "desc", desc)
	candidate, fromAddr, txNom, err := tp.validateRevokeCandidate(desc)
	if err != nil {
		tp.log.Warn("runRevokeCandidate to validate Revoke error", "error", err.Error())
		return err
	}

	keyRevoke := genRevokeKey(txNom)
	if _, ok := tp.revokeCache.Load(txNom); ok {
		tp.log.Warn("runRevokeCandidate error", "error", "revoke repeated")
		return errors.New("runRevokeCandidate error revoke repeated")
	}
	val, err := tp.utxoVM.GetFromTable(nil, []byte(keyRevoke))
	if (err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound) || val != nil {
		tp.log.Warn("runRevokeCandidate error revoke repeated or get revoke key from db error", "val", hex.EncodeToString(val),
			"error", err)
		return errors.New("runRevokeCandidate error revoke repeated or get revoke key from db error")
	}

	key := GenCandidateNominateKey(candidate)
	keyBal := genCandidateBallotsKey(candidate)
	revokeKey := genRevokeCandidateKey(candidate, hex.EncodeToString(desc.Tx.Txid))
	keyNominateRecord := GenNominateRecordsKey(fromAddr, candidate, txNom)

	txid, _ := tp.utxoVM.GetFromTable(nil, []byte(key))
	if hex.EncodeToString(txid) != txNom {
		tp.log.Warn("runRevokeCandidate GetFromTable error, txid not match!", "txid", hex.EncodeToString(txid), "txNom", txNom)
		return errors.New("runRevokeCandidate GetFromTable error, txid not match")
	}

	kal, ok := tp.candidateBallotsCache.Load(keyBal)
	if ok {
		blVal := kal.(*candidateBallotsCacheValue)
		tp.log.Trace("runRevokeCandidate get from cache ok", "kal", blVal)
		tp.context.UtxoBatch.Delete([]byte(key))
		tp.context.UtxoBatch.Delete([]byte(keyNominateRecord))
		tp.context.UtxoBatch.Put([]byte(revokeKey), []byte(strconv.FormatInt(blVal.ballots, 10)))
		blVal.isDel = true
		blVal.ballots = 0
		tp.candidateBallotsCache.Store(keyBal, blVal)
		// 记录撤销记录
		tp.revokeCache.Store(keyRevoke, true)
		tp.context.UtxoBatch.Put([]byte(keyRevoke), desc.Tx.Txid)
		tp.log.Trace("runRevokeCandidate success")
		return nil
	}

	bal, ok := tp.candidateBallots.Load(keyBal)
	if ok {
		val := bal.(int64)
		tp.log.Trace("runRevokeCandidate get from mem ok", "val", val)
		blVal := &candidateBallotsCacheValue{}
		tp.context.UtxoBatch.Delete([]byte(key))
		tp.context.UtxoBatch.Delete([]byte(keyNominateRecord))
		tp.context.UtxoBatch.Put([]byte(revokeKey), []byte(strconv.FormatInt(val, 10)))
		blVal.isDel = true
		blVal.ballots = 0
		tp.candidateBallotsCache.Store(keyBal, blVal)
		// 记录撤销记录
		tp.revokeCache.Store(keyRevoke, true)
		tp.context.UtxoBatch.Put([]byte(keyRevoke), desc.Tx.Txid)
		tp.log.Trace("runRevokeCandidate success")
		return nil
	}
	return nil
}

// 执行检票
func (tp *TDpos) runCheckValidater(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("runCheckValidater desc", "desc", desc, "txid", fmt.Sprintf("%x", desc.Tx.Txid))
	version, term, err := tp.validateCheckValidater(desc)
	if err != nil {
		tp.log.Warn("runCheckValidater error for validateCheckValidater error", "error", err)
		return err
	}
	key := GenTermCheckKey(version, term)
	_, err = tp.utxoVM.GetFromTable(nil, []byte(key))
	if common.NormalizedKVError(err) != common.ErrKVNotFound {
		return err
	}
	proposers, err := tp.genTermProposer()
	tp.log.Trace("runCheckValidater", "proposers", proposers, "err", err)
	if err == ErrProposerNotEnough {
		// 没有检出足够的候选人, 则往前回溯, 使用上一轮的候选人代替
		for i := term - 1; i >= 1; i-- {
			if i == 1 {
				proposers = tp.config.initProposer[1]
			}
			keyPre := GenTermCheckKey(version, i)
			val, err := tp.utxoVM.GetFromTable(nil, []byte(keyPre))
			tp.log.Trace("runCheckValidater from previous", "keyPre", keyPre, "val", val)
			if val != nil {
				err = json.Unmarshal(val, &proposers)
				if err == nil {
					break
				}
			}
		}
	}
	if proposers != nil {
		proposersJSON, _ := json.Marshal(proposers)
		tp.log.Info("runCheckValidater", "key", key, "proposersJson", proposersJSON, "proposers", proposers)
		tp.context.UtxoBatch.Put([]byte(key), proposersJSON)
		tp.triggerProposerChanged(proposers)
		return nil
	}
	tp.log.Warn("runCheckValidater error")
	return errors.New("runCheckValidater error")
}

// triggerProposerChanged triggers a ProposerChanged event
func (tp *TDpos) triggerProposerChanged(proposers []*cons_base.CandidateInfo) {
	em := &events.EventMessage{
		BcName:   tp.bcname,
		Type:     events.ProposerChanged,
		Priority: 0,
		Sender:   tp,
	}

	msg := &cons_base.MinersChangedEvent{
		BcName:        tp.bcname,
		CurrentMiners: tp.GetCoreMiners(),
		NextMiners:    make([]*cons_base.MinerInfo, 0),
	}

	for _, proposer := range proposers {
		miner := &cons_base.MinerInfo{
			Address:  proposer.Address,
			PeerInfo: proposer.PeerAddr,
		}
		msg.NextMiners = append(msg.NextMiners, miner)
	}

	em.Message = msg
	eb := events.GetEventBus()
	_, err := eb.FireEventAsync(em)
	if err != nil {
		tp.log.Warn("triggerProposerChanged fire event failed", "error", err)
	}
}
