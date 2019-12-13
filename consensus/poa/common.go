package poa

import (
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"github.com/xuperchain/xuperunion/common"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
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
	if poa.curTerm == 0 {
		if len(poa.config.initProposer) <= 0 {
			poa.log.Warn("TDpos getTermProposer error, no proposer in term 1")
			return "", errors.New("no proposer in term 1")
		}
		return poa.config.initProposer[0].Address, nil
	}

	proposers := poa.getTermProposer(poa.curTerm)
	poa.log.Trace("TDpos getTermProposer result", "term", poa.curTerm, "proposers", proposers)
	if proposers == nil {
		poa.log.Warn("TDpos getTermProposer error", "term", poa.curTerm)
		return "", errors.New("no proposer found")
	}

	// current proposer is the last proposer of this term
	if poa.curPos >= int64(len(proposers)) {
		proposers := poa.getTermProposer(poa.curTerm + 1)
		if proposers == nil {
			poa.log.Warn("TDpos getTermProposer error", "term", poa.curTerm + 1)
			return "", errors.New("no proposer found")
		}
		return proposers[0].Address, nil
	} else if poa.curPos < 0 {
		poa.log.Warn("TDpos getTermProposer error, pos index out of range", "pos", poa.curTerm, "proposers", proposers)
		return "", errors.New("invalid pos")
	}

	// leader not changed
	if poa.curBlockNum <= poa.config.blockNum {
		return proposers[poa.curPos].Address, nil
	}

	// return next proposer of current term
	return proposers[(poa.curPos+1)%int64(len(proposers))].Address, nil
}

// 查询当前轮的验证者名单
func (poa *Poa) getTermProposer(term int64) []*cons_base.CandidateInfo {
	if term == 1 {
		return poa.config.initProposer
	}
	key := GenTermCheckKey(poa.version, term)
	val, err := poa.utxoVM.GetFromTable(nil, []byte(key))
	if err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound {
		poa.log.Error("TDpos getTermProposer vote result error", "term", term, "error", err)
		return nil
	} else if common.NormalizedKVError(err) == common.ErrKVNotFound {
		it := poa.utxoVM.ScanWithPrefix([]byte(genTermCheckKeyPrefix(poa.version)))
		defer it.Release()
		if it.Last() {
			termLast, err := parseTermCheckKey(string(it.Key()))
			poa.log.Trace("TDpos getTermProposer ", "termLast", string(it.Key()))
			if err != nil {
				poa.log.Warn("TDpos getTermProposer parseTermCheckKey error", "error", err)
				return nil
			}
			if termLast == term+1 {
				if it.Prev() {
					poa.log.Trace("TDpos getTermProposer ", "key", string(it.Key()))
				} else {
					poa.log.Trace("TDpos getTermProposer parseTermCheckKey get prev nil")
					it.Last()
					keyLast := string(it.Key())
					it.First()
					ketFirst := string(it.Key())
					if keyLast == ketFirst {
						return poa.config.initProposer
					}
					poa.log.Warn("TDpos getTermProposer parseTermCheckKey get prev error", "error", err)
					return nil
				}
			}
			val = it.Value()
		} else {
			poa.log.Warn("TDpos getTermProposer query from table is nil", "tp.config.initProposer", poa.config.initProposer)
			return poa.config.initProposer
		}
	}
	var proposers []*cons_base.CandidateInfo
	err = json.Unmarshal(val, &proposers)
	if err != nil {
		poa.log.Error("TDpos Unmarshal vote result error", "term", term, "error", err)
		return nil
	}
	return proposers

}

// 生成当前轮的验证者名单
func (poa *Poa) genTermProposer() ([]*cons_base.CandidateInfo, error) {
	return poa.proposerInfos, nil
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
