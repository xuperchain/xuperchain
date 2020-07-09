package base

// MinerInfo defines the essential info of miner
type MinerInfo struct {
	Address  string // xchain address
	PeerInfo string // peer info(in most cases is the network address)
}

// MinersChangedEvent define the information of miners would be changed.
// this event would be fired when DPoS proposers initialized or next round proposers are selected.
type MinersChangedEvent struct {
	BcName        string
	CurrentMiners []*MinerInfo
	NextMiners    []*MinerInfo
}

// ConsensusStatus define the status of Consensus
type ConsensusStatus struct {
	Proposer string
	Term     int64
	BlockNum int64
}

// CandidateInfo define the candidate info
type CandidateInfo struct {
	// Address of node
	Address string
	// Neturl of node
	PeerAddr string
}

// CandidateInfoEqual return whether candidate info is equal
func CandidateInfoEqual(left, right []*CandidateInfo) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := 0; idx < len(left); idx++ {
		if left[idx].Address != right[idx].Address || left[idx].PeerAddr != right[idx].PeerAddr {
			return false
		}
	}
	return true
}

// CandidateInfos define the struct of proposers
type CandidateInfos struct {
	Proposers []*CandidateInfo `json:"proposers"`
}
