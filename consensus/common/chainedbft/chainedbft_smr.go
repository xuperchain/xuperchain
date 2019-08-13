package chainedbft

import (
	log "github.com/xuperchain/log15"
	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"
)

// Smr is the state of the node
type Smr struct {
	slog log.Logger
	// config is the config of ChainedBft
	config Config

	// proposalQC is the proposalBlock's QC
	proposalQC *chainedbft_pb.QuorumCert
	// generateQC is the proposalBlock's QC, refer to generateBlock's votes
	generateQC *chainedbft_pb.QuorumCert
	// lockedQC is the generateBlock's QC, refer to lockedBlock's votes
	lockedQC *chainedbft_pb.QuorumCert
	// votes of QC in mem
	qcVotes map[string]*chainedbft_pb.QCSignInfos
	// quitCh stop channel
	quitCh chan bool
}

// NewSmr return smr instance of hotstuff
func NewSmr(cfg Config) (*Smr, error) {
	return nil, nil
}

// ProcessNewView used to process while view changed. There are three scenarios:
// 1 As the new leader, it will wait for (m-f) replica's new view msg and then create an new Proposers;
// 2 As a normal replica, it will send new view msg to leader;
// 3 As the previous leader, it will send new view msg to new leader with votes of its QuorumCert;
func (s *Smr) processNewView() error {
	return nil
}

// ProcessPropose used to generate new QuorumCert and broadcast to other replicas
func (s *Smr) processPropose() error {
	return nil
}

// HandleReceiveVote used to process while receiving vote
func (s *Smr) HandleReceiveVote() error {
	return nil
}

// HandleReceiveNewView used to handle new view msg from other replicas
func (s *Smr) HandleReceiveNewView() error {
	return nil
}

// HandleReceiveProposal is the core function of hotstuff. It uesd to change QuorumCerts's phase.
// It will change three previous QuorumCerts's state because hotstuff is a three chained bft.
func (s *Smr) HandleReceiveProposal() error {
	return nil
}

// handleOnReceiveProposal used to process while receiving Proposal
func (s *Smr) handleStateChangeOnReceiveProposal() error {
	return nil
}

// handleStateChangeOnPreCommit used to process while
func (s *Smr) handleStateChangeOnPreCommit() error {
	return nil
}

// handleOnCommit
func (s *Smr) handleStateChangeOnCommit() error {
	return nil
}
