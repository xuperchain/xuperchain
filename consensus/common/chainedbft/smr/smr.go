package smr

import (
	"context"
	"errors"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"

	"github.com/xuperchain/xuperunion/p2pv2"
	p2p_pb "github.com/xuperchain/xuperunion/p2pv2/pb"
)

var (
	// ErrNewViewNum used to return error new view number
	ErrNewViewNum = errors.New("new view number error")
)

// NewSmr return smr instance
func NewSmr(cfg config.Config, bcname string, p2p *p2pv2.P2PServerV2, proposalQC, generateQC, lockedQC *chainedbft_pb.QuorumCert) (*Smr, error) {
	xlog := log.New("module", "smr")
	// set up smr
	if proposalQC == nil || generateQC == nil || lockedQC == nil {
		xlog.Error("NewSmr params error, init QC status can not be nil")
		return nil, errors.New("NewSmr QC params error")
	}

	smr := &Smr{
		// TODO: zq check init all member variables
		slog:       xlog,
		bcname:     bcname,
		p2p:        p2p,
		p2pMsgChan: make(chan *p2p_pb.XuperMessage, cfg.NetMsgChanSize),
		votedView:  generateQC.ViewNumber,
		proposalQC: proposalQC,
		generateQC: generateQC,
		lockedQC:   lockedQC,
		QuitCh:     make(chan bool, 1),
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
	if _, err := s.p2p.Register(p2pv2.NewSubscriber(nil, p2p_pb.XuperMessage_CHAINED_BFT_PHASE_MSG, s.HandleReceivePhaseMsg, "")); err != nil {
		return err
	}

	if _, err := s.p2p.Register(p2pv2.NewSubscriber(s.p2pMsgChan, p2p_pb.XuperMessage_CHAINED_BFT_VOTE_MSG, nil, "")); err != nil {
		return err
	}
	return nil
}

// ProcessNewView used to process while view changed. There are three scenarios:
// 1 As the new leader, it will wait for (m-f) replica's new view msg and then create an new Proposers;
// 2 As a normal replica, it will send new view msg to leader;
// 3 As the previous leader, it will send new view msg to new leader with votes of its QuorumCert;
func (s *Smr) ProcessNewView(viewNumber int64, leader, preLeader string) error {
	// if new view number less than voted view number, return error
	if viewNumber < s.votedView {
		s.slog.Error("ProcessNewView error", "error", ErrNewViewNum.Error())
		return ErrNewViewNum
	}

	newViewMsg := &chainedbft_pb.ChainedBftPhaseMessage{
		Type:       chainedbft_pb.QCState_NEW_VIEW,
		ViewNumber: viewNumber,
	}

	if preLeader == s.address {
		newViewMsg.JustifyQC = &chainedbft_pb.QuorumCert{
			ProposalId: s.proposalQC.GetProposalId(),
			Type:       s.proposalQC.GetType(),
			ViewNumber: s.proposalQC.GetViewNumber(),
			SignInfos:  s.proposalQC.GetSignInfos(),
		}
	}

	// TODO: zq sign for this msg

	// if as the new leader, wait for the (n-f) new view message from other replicas and call back extenal consensus
	if leader == s.address {
		// TODO: zq register call back function to extenal consensus
		s.slog.Trace("ProcessNewView as a new leader, wait for (n - f) new view messags")
		s.addViewMsgs(newViewMsg)
		return s.addViewMsgs(newViewMsg)
	}

	// send to next leader
	msgBuf, err := proto.Marshal(newViewMsg)
	if err != nil {
		s.slog.Error("ProcessNewView marshal msg error", "error", err)
		return err
	}

	netMsg, _ := p2p_pb.NewXuperMessage(p2p_pb.XuperMsgVersion3, s.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_PHASE_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	opts := []p2pv2.MessageOption{
		p2pv2.WithBcName(s.bcname),
		p2pv2.WithTargetPeerAddrs([]string{leader}),
	}
	go s.p2p.SendMessage(context.Background(), netMsg, opts...)
	return nil
}

// ProcessPropose used to generate new QuorumCert and broadcast to other replicas
func (s *Smr) ProcessPropose() (*chainedbft_pb.QuorumCert, error) {
	return nil, nil
}

// HandleReceiveVoteMsg used to process while receiving vote msg from network
func (s *Smr) HandleReceiveVoteMsg() error {
	return nil
}

// HandleReceivePhaseMsg used to process while receiving proposal msg from network
func (s *Smr) HandleReceivePhaseMsg(ctx context.Context, msg *p2p_pb.XuperMessage) (*p2p_pb.XuperMessage, error) {
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
