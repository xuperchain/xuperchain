package poa

import (
	"errors"
)

// 调度产生的矿工与自身进行进行比较
func (poa *Poa) isProposer(term int64, pos int64, address []byte) bool {
	poa.mutex.RLock()
	defer poa.mutex.RUnlock()
	miner, err := poa.getProposer(term, pos)
	if err != nil {
		return false
	}
	return string(address) == miner
}

// getProposer return the proposer of given term, pos
func (poa *Poa) getProposer(term int64, pos int64) (string, error) {
	if term == 1 {
		if len(poa.config.initProposer) <= 0 {
			poa.log.Warn("poa getTermProposer error, no proposer in term 1")
			return "", errors.New("no proposer in term 1")
		}
		return poa.config.initProposer[0].Address, nil
	}
	poa.log.Trace("poa getTermProposer result", "term", term, "proposers", poa.proposerInfos)
	if poa.proposerInfos == nil {
		poa.log.Warn("poa getTermProposer error", "term", term)
		return "", errors.New("no proposer found")
	}
	if pos < 0 || pos > poa.proposerNum-1 {
		poa.log.Warn("poa getTermProposer error, pos index out of range", "pos", pos, "total num", poa.proposerNum)
		return "", errors.New("invalid pos")
	}
	return poa.proposerInfos[pos].Address, nil
}
