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
	Address  string
	PeerAddr string
}
