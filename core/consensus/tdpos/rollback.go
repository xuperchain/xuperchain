package tdpos

import (
	"strconv"

	"encoding/hex"
	"encoding/json"

	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/pb"
)

// 回滚投票
func (tp *TDpos) rollbackVote(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("Start to rollbackVote", "desc", desc)
	// 验证选票信息有效性, 并解析选票参数
	voteInfo, err := tp.validateVote(desc, true)
	if err != nil {
		tp.log.Warn("rollbackVote error", "error", err)
		return nil
	}

	for i := 0; i < len(voteInfo.candidates); i++ {
		keyCanBal := GenCandidateBallotsKey(voteInfo.candidates[i])
		keyCandidateVote := GenCandidateVoteKey(voteInfo.candidates[i], voteInfo.voter, hex.EncodeToString(desc.Tx.Txid))
		keyVoteCandidate := GenVoteCandidateKey(voteInfo.voter, voteInfo.candidates[i], hex.EncodeToString(desc.Tx.Txid))
		if val, ok := tp.candidateBallotsCache.Load(keyCanBal); ok {
			// 先看一下缓存里有没有,有的话则直接处理缓存
			canBal := val.(*candidateBallotsValue)
			canBal.ballots -= voteInfo.ballots
			tp.candidateBallotsCache.Store(keyCanBal, canBal)
		} else {
			// 从内存中load出来再处理
			if bal, ok := tp.candidateBallots.Load(keyCanBal); ok {
				bals := bal.(int64) - voteInfo.ballots
				canBal := &candidateBallotsValue{
					ballots: bals,
					isDel:   false,
				}
				tp.candidateBallotsCache.Store(keyCanBal, canBal)
			} else {
				// 不在候选人池中
				tp.log.Trace("rollbackVote no need to process")
				return nil
			}
		}
		tp.context.UtxoBatch.Delete([]byte(keyCandidateVote))
		tp.context.UtxoBatch.Delete([]byte(keyVoteCandidate))
	}
	return nil
}

// 回滚撤销投票
func (tp *TDpos) rollbackRevokeVote(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("Start to rollbackRevokeVote", "desc", desc)
	voteInfo, txVote, err := tp.validateRevokeVote(desc)
	if err != nil {
		tp.log.Warn("rollbackRevokeVote error", "error", err)
		return nil
	}

	keyRevoke := GenRevokeKey(txVote)
	val, err := tp.utxoVM.GetFromTable(nil, []byte(keyRevoke))
	if val == nil {
		tp.log.Warn("rollbackRevokeVote error get revoke from db is nil!")
		return nil
	}

	if string(val) != hex.EncodeToString(desc.Tx.Txid) {
		tp.log.Warn("rollbackRevokeVote omit val not equal this tx", "val", string(val), "txid", hex.EncodeToString(desc.Tx.Txid))
		return nil
	}

	for i := 0; i < len(voteInfo.candidates); i++ {
		keyCanBal := GenCandidateBallotsKey(voteInfo.candidates[i])
		keyCandidateVote := GenCandidateVoteKey(voteInfo.candidates[i], voteInfo.voter, txVote)
		keyVoteCandidate := GenVoteCandidateKey(voteInfo.voter, voteInfo.candidates[i], txVote)
		if val, ok := tp.candidateBallotsCache.Load(keyCanBal); ok {
			// 先看一下缓存里有没有,有的话则直接处理缓存
			canBal := val.(*candidateBallotsValue)
			canBal.ballots += voteInfo.ballots
			tp.candidateBallotsCache.Store(keyCanBal, canBal)

		} else {
			// 从内存中load出来再处理
			if bal, ok := tp.candidateBallots.Load(keyCanBal); ok {
				bals := bal.(int64) + voteInfo.ballots
				canBal := &candidateBallotsValue{
					ballots: bals,
					isDel:   false,
				}
				tp.candidateBallotsCache.Store(keyCanBal, canBal)
			} else {
				// 不在候选人池子里，则从db中恢复
				tp.log.Trace("rollbackRevokeVote load from db")
				balVal, _ := tp.utxoVM.GetFromTable(nil, []byte(keyCanBal))
				if val != nil {
					ballots, err := strconv.ParseInt(string(balVal), 10, 64)
					if err != nil {
						tp.log.Warn("rollbackRevokeVote load from db parse error")
					} else {
						canBal := &candidateBallotsValue{
							ballots: ballots + voteInfo.ballots,
							isDel:   true,
						}
						tp.candidateBallotsCache.Store(keyCanBal, canBal)
					}
				}
			}
		}
		// 删除revoke记录
		tp.context.UtxoBatch.Delete([]byte(keyRevoke))
		// 增加投票记录
		tp.context.UtxoBatch.Put([]byte(keyCandidateVote), []byte(strconv.FormatInt(voteInfo.ballots, 10)))
		tp.context.UtxoBatch.Put([]byte(keyVoteCandidate), []byte(strconv.FormatInt(voteInfo.ballots, 10)))
	}
	return nil
}

// 回滚候选人提名
func (tp *TDpos) rollbackNominateCandidate(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("Start to rollbackNominateCandidate", "desc", desc)
	canInfo, fromAddr, err := tp.validateNominateCandidate(desc)
	if err != nil {
		tp.log.Warn("rollbackNominateCandidate to validate nominate error", "error", err.Error())
		return nil
	}
	candidate := canInfo.Address
	key := GenCandidateNominateKey(candidate)
	keyBl := GenCandidateBallotsKey(candidate)
	keyCanInfo := GenCandidateInfoKey(candidate)

	keyNominateRecord := GenNominateRecordsKey(fromAddr, candidate, hex.EncodeToString(desc.Tx.Txid))

	txid, _ := tp.utxoVM.GetFromTable(nil, []byte(key))
	if string(txid) != string(desc.Tx.Txid) {
		tp.log.Warn("rollbackNominateCandidate GetFromTable error, txid not match!")
		return nil
	}

	val, ok := tp.candidateBallotsCache.Load(keyBl)
	if ok {
		canBal := val.(*candidateBallotsValue)
		if !canBal.isDel {
			canBal.isDel = true
		}
		tp.candidateBallotsCache.Store(keyBl, canBal)
		tp.context.UtxoBatch.Delete([]byte(key))
		tp.context.UtxoBatch.Delete([]byte(keyNominateRecord))
		tp.context.UtxoBatch.Delete([]byte(keyCanInfo))
		return nil
	}

	ballots, ok := tp.candidateBallots.Load(keyBl)
	if ok {
		canBal := &candidateBallotsValue{}
		canBal.isDel = true
		canBal.ballots = ballots.(int64)
		tp.candidateBallotsCache.Store(keyBl, canBal)
		tp.context.UtxoBatch.Delete([]byte(key))
		tp.context.UtxoBatch.Delete([]byte(keyNominateRecord))
		tp.context.UtxoBatch.Delete([]byte(keyCanInfo))
		return nil
	}
	tp.log.Warn("rollbackNominateCandidate error, not find ballots")
	return nil
}

// 回滚撤销候选人提名
func (tp *TDpos) rollbackRevokeCandidate(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("Start to rollbackRevokeCandidate", "desc", desc)
	candidate, fromAddr, txNom, err := tp.validateRevokeCandidate(desc)
	if err != nil {
		tp.log.Warn("rollbackRevokeCandidate to validate revoke error", "error", err.Error())
		return nil
	}
	keyRevoke := GenRevokeKey(txNom)
	val, err := tp.utxoVM.GetFromTable(nil, []byte(keyRevoke))
	if val == nil {
		tp.log.Warn("rollbackRevokeCandidate error get revoke from db is nil!")
		return nil
	}

	if string(val) != string(desc.Tx.Txid) {
		tp.log.Warn("rollbackRevokeCandidate omit val not equal this tx", "val", string(val), "txid", string(desc.Tx.Txid))
		return nil
	}

	key := GenCandidateNominateKey(candidate.Address)
	keyBl := GenCandidateBallotsKey(candidate.Address)
	keyCanInfo := GenCandidateInfoKey(candidate.Address)
	keyNominateRecord := GenNominateRecordsKey(fromAddr, candidate.Address, txNom)
	revokeCandidateKey := GenRevokeCandidateKey(candidate.Address, hex.EncodeToString(desc.Tx.Txid))

	tp.context.UtxoBatch.Put([]byte(key), []byte(txNom))
	tp.context.UtxoBatch.Put([]byte(keyNominateRecord), []byte(txNom))
	canInfoValue, err := json.Marshal(candidate)
	if err != nil {
		tp.log.Warn("rollbackRevokeCandidate json marshal failed", "err", err)
		return err
	}
	tp.context.UtxoBatch.Put([]byte(keyCanInfo), []byte(canInfoValue))

	bals := int64(0)
	val, err = tp.utxoVM.GetFromTable(nil, []byte(revokeCandidateKey))
	if val != nil {
		bals, err = strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			tp.log.Warn("Parse rollbackRevokeCandidate ballots error", "error", err)
			return err
		}
	}
	blVal := &candidateBallotsValue{
		ballots: bals,
		isDel:   false,
	}
	tp.candidateBallotsCache.Store(keyBl, blVal)
	tp.context.UtxoBatch.Delete([]byte(revokeCandidateKey))
	tp.context.UtxoBatch.Delete([]byte(revokeCandidateKey))

	// 删除revoke记录
	tp.context.UtxoBatch.Delete([]byte(keyRevoke))
	return nil
}

// 回滚检票
func (tp *TDpos) rollbackCheckValidater(desc *contract.TxDesc, block *pb.InternalBlock) error {
	tp.log.Trace("Start to rollbackCheckValidater", "desc", desc)

	version, term, err := tp.validateCheckValidater(desc)
	if err != nil {
		tp.log.Warn("runCheckValidater error for validateCheckValidater error", "error", err)
		return nil
	}
	key := GenTermCheckKey(version, term)
	tp.context.UtxoBatch.Delete([]byte(key))
	return nil
}
