package poa

import (
	"errors"
	"math/big"
	"strings"

	consBase "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/pb"
)

const minNominateProportion = 100000

// 调度产生的矿工与自身进行进行比较
func (poa *Poa) isProposer(term int64, pos int64, address []byte) bool {
	if term == 0 {
		return false
	}
	proposers := poa.getTermProposer(term)
	poa.log.Trace("TDpos getTermProposer result", "term", term, "proposers", proposers)
	if proposers == nil {
		poa.log.Warn("TDpos getTermProposer error", "term", term)
		return false
	}
	if pos < 0 || pos > int64(len(proposers)-1) {
		poa.log.Warn("TDpos getTermProposer error, pos index out of range", "pos", pos, "proposers", proposers)
		return false
	}
	return string(address) == proposers[pos].Address
}

// getProposer return the proposer of given term, pos
func (poa *Poa) getProposer(term int64, pos int64) (string, error) {
	if term == 0 {
		if len(poa.config.initProposer) <= 0 {
			poa.log.Warn("TDpos getTermProposer error, no proposer in term 1")
			return "", errors.New("no proposer in term 1")
		}
		return poa.config.initProposer[0].Address, nil
	}
	proposers := poa.getTermProposer(term)
	poa.log.Trace("TDpos getTermProposer result", "term", term, "proposers", proposers)
	if proposers == nil {
		poa.log.Warn("TDpos getTermProposer error", "term", term)
		return "", errors.New("no proposer found")
	}
	if pos < 0 || pos > int64(len(proposers)-1) {
		poa.log.Warn("TDpos getTermProposer error, pos index out of range", "pos", pos, "proposers", proposers)
		return "", errors.New("invalid pos")
	}
	return proposers[pos].Address, nil
}

// getNextProposer return the next block proposer of given term,pos
func (poa *Poa) getNextProposer() (string, error) {
	proposers := poa.getTermProposer(poa.curTerm)
	poa.log.Trace("Poa getTermProposer result", "term", poa.curTerm, "proposers", proposers)
	if proposers == nil {
		poa.log.Warn("Poa getTermProposer error", "term", poa.curTerm)
		return poa.proposerInfos[0].Address, errors.New("no proposer found")
	}

	// current proposer is the last proposer of this term
	if poa.curPos >= int64(len(proposers)) {
		acl, confirmed, err := poa.aclManager.GetAccountACLWithConfirmed(poa.accountName)
		if err != nil || acl == nil {
			poa.log.Warn("Poa getConfirmedACL error", "accountName", poa.accountName, "term", poa.curTerm+1)
			return poa.proposerInfos[0].Address, errors.New("no proposer confirmed")
		}
		if confirmed {
			poa.mutex.Lock()
			defer poa.mutex.Unlock()
			l := 0
			r := len(acl.AksWeight)
			tmpSet := make([]*consBase.CandidateInfo, r)
			for address, weight := range acl.AksWeight {
				if weight > 0 {
					tmpSet[l] = &consBase.CandidateInfo{
						Address:  address,
						PeerAddr: "",
					}
					l++
				} else {
					tmpSet[r] = &consBase.CandidateInfo{
						Address:  address,
						PeerAddr: "",
					}
					r--
				}
				poa.proposerInfos = tmpSet
			}
			if l > 1 {
				poa.log.Warn("more than one weights are greater than 0, means there are more than one CAs")
			}
		}
		return proposers[0].Address, nil
	} else if poa.curPos < 0 {
		poa.log.Warn("Poa getTermProposer error, pos index out of range", "pos", poa.curTerm, "proposers", proposers)
		return "", errors.New("invalid pos")
	}

	// leader not changed
	if poa.curBlockNum <= poa.config.blockNum {
		return proposers[poa.curPos].Address, nil
	}

	// return next proposer of current term
	return proposers[poa.curPos+1].Address, nil
}

// 查询当前轮的验证者名单
func (poa *Poa) getTermProposer(term int64) []*consBase.CandidateInfo {
	if term == 1 {
		return poa.config.initProposer
	}
	return poa.proposerInfos
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

func (poa *Poa) isAuthAddress(candidate string, initiator string, authRequire []string) bool {
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
