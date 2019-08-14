package chainedbft

import (
	"context"
	"errors"

	log "github.com/xuperchain/log15"
	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// Smr is the state of the node
type Smr struct {
	slog log.Logger
	// config is the config of ChainedBft
	config Config
	// bcname of ChainedBft instance
	bcname string
	// p2p is the network instance
	p2p *p2pv2.P2PServerV2
	// p2pMsgChan is the msg channel registered to network
	p2pMsgChan chan *xuper_p2p.XuperMessage

	// Hotstuff State of this nodes
	// proposalQC is the proposalBlock's QC
	proposalQC *chainedbft_pb.QuorumCert
	// generateQC is the proposalBlock's QC, refer to generateBlock's votes
	generateQC *chainedbft_pb.QuorumCert
	// lockedQC is the generateBlock's QC, refer to lockedBlock's votes
	lockedQC *chainedbft_pb.QuorumCert
	// votes of QC in mem
	qcVotes map[string]*chainedbft_pb.QCSignInfos
	// new view msg gathered from other replicas
	newViewMsgs map[int64]*chainedbft_pb.ChainedBftPhaseMessage
	// quitCh stop channel
	quitCh chan bool
}

// NewSmr return smr instance
func NewSmr(cfg Config, bcname string, p2p *p2pv2.P2PServerV2, proposalQC, generateQC, lockedQC *chainedbft_pb.QuorumCert) (*Smr, error) {
	xlog := log.New("module", "smr")
	// set up smr
	if proposalQC == nil || generateQC == nil || lockedQC == nil {
		xlog.Error("NewSmr params error, init QC status can not be nil")
		return nil, errors.New("NewSmr QC params error")
	}

	smr := &Smr{
		slog:       xlog,
		bcname:     bcname,
		p2p:        p2p,
		p2pMsgChan: make(chan *xuper_p2p.XuperMessage, cfg.NetMsgChanSize),
		proposalQC: proposalQC,
		generateQC: generateQC,
		lockedQC:   lockedQC,
		quitCh:     make(chan bool, 1),
	}
	// register to p2p network
	if err := smr.registerToNetwork(); err != nil {
		xlog.Error("NewSmr register to network error", "error", err)
		return nil, err
	}
	return smr, nil
}

// registerToNetwork register msg handler to p2p network
func (s *Smr) registerToNetwork() error {
	if _, err := s.p2p.Register(p2pv2.NewSubscriber(nil, xuper_p2p.XuperMessage_CHAINED_BFT_PHASE_MSG, s.HandleReceivePhaseMsg, "")); err != nil {
		return err
	}

	if _, err := s.p2p.Register(p2pv2.NewSubscriber(s.p2pMsgChan, xuper_p2p.XuperMessage_CHAINED_BFT_VOTE_MSG, nil, "")); err != nil {
		return err
	}
	return nil
}

// ProcessNewView used to process while view changed. There are three scenarios:
// 1 As the new leader, it will wait for (m-f) replica's new view msg and then create an new Proposers;
// 2 As a normal replica, it will send new view msg to leader;
// 3 As the previous leader, it will send new view msg to new leader with votes of its QuorumCert;
func (s *Smr) processNewView() error {
	return nil
}

// ProcessPropose used to generate new QuorumCert and broadcast to other replicas
func (s *Smr) processPropose() (*chainedbft_pb.QuorumCert, error) {
	return nil, nil
}

// HandleReceiveVoteMsg used to process while receiving vote msg from network
func (s *Smr) HandleReceiveVoteMsg() error {
	return nil
}

// HandleReceivePhaseMsg used to process while receiving proposal msg from network
func (s *Smr) HandleReceivePhaseMsg(ctx context.Context, msg *xuper_p2p.XuperMessage) (*xuper_p2p.XuperMessage, error) {
	return nil, nil
}

// handleReceiveNewView used to handle new view msg from other replicas
func (s *Smr) handleReceiveNewView() error {
	return nil
}

// handleReceiveProposal is the core function of hotstuff. It uesd to change QuorumCerts's phase.
// It will change three previous QuorumCerts's state because hotstuff is a three chained bft.
func (s *Smr) handleReceiveProposal() error {
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
