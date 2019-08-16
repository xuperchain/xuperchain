package smr

import (
	"bytes"
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
	// ErrSafeProposal check new proposal error
	ErrSafeProposal = errors.New("check new proposal error")
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
	if _, err := s.p2p.Register(p2pv2.NewSubscriber(s.p2pMsgChan,
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_VIEW_MSG, nil, "")); err != nil {
		return err
	}

	if _, err := s.p2p.Register(p2pv2.NewSubscriber(s.p2pMsgChan,
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_PROPOSAL_MSG, nil, "")); err != nil {
		return err
	}

	if _, err := s.p2p.Register(p2pv2.NewSubscriber(s.p2pMsgChan,
		p2p_pb.XuperMessage_CHAINED_BFT_VOTE_MSG, nil, "")); err != nil {
		return err
	}
	return nil
}

// Start used to start smr instance and process msg
func (s *Smr) Start() {
	for {
		select {
		case msg := <-s.p2pMsgChan:
			go s.handleReceivedMsg(msg)
		case <-s.QuitCh:
			s.slog.Info("Quit chainedbft smr ...")
			s.QuitCh <- true
			s.stop()
			return
		}
	}
}

// stop used to stop smr instance
func (s *Smr) stop() {
	// TODO: zq
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

	// TODO: zq sign for this msg
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

	// if as the new leader, wait for the (n-f) new view message from other replicas and call back extenal consensus
	if leader == s.address {
		s.slog.Trace("ProcessNewView as a new leader, wait for (n - f) new view messags")
		s.addViewMsg(newViewMsg)
		return s.addViewMsg(newViewMsg)
	}

	// send to next leader
	msgBuf, err := proto.Marshal(newViewMsg)
	if err != nil {
		s.slog.Error("ProcessNewView marshal msg error", "error", err)
		return err
	}

	netMsg, _ := p2p_pb.NewXuperMessage(p2p_pb.XuperMsgVersion3, s.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_VIEW_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	opts := []p2pv2.MessageOption{
		p2pv2.WithBcName(s.bcname),
		p2pv2.WithTargetPeerAddrs([]string{leader}),
	}
	go s.p2p.SendMessage(context.Background(), netMsg, opts...)
	return nil
}

// ProcessProposal used to generate new QuorumCert and broadcast to other replicas
func (s *Smr) ProcessProposal(viewNumber int64, proposalID,
	proposalMsg []byte) (*chainedbft_pb.QuorumCert, error) {

	qc := &chainedbft_pb.QuorumCert{
		ProposalId:  proposalID,
		ProposalMsg: proposalMsg,
		ViewNumber:  viewNumber,
		Type:        chainedbft_pb.QCState_PREPARE,
		SignInfos:   &chainedbft_pb.QCSignInfos{},
	}

	// TODO: zq sign for this msg
	propMsg := &chainedbft_pb.ChainedBftPhaseMessage{
		Type:       chainedbft_pb.QCState_PREPARE,
		ViewNumber: viewNumber,
		ProposalQC: qc,
	}

	// send to other replicas
	msgBuf, err := proto.Marshal(propMsg)
	if err != nil {
		s.slog.Error("ProcessProposal marshal msg error", "error", err)
		return nil, err
	}
	netMsg, _ := p2p_pb.NewXuperMessage(p2p_pb.XuperMsgVersion3, s.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_PROPOSAL_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	opts := []p2pv2.MessageOption{
		p2pv2.WithBcName(s.bcname),
		p2pv2.WithFilters([]p2pv2.FilterStrategy{p2pv2.CorePeersStrategy}),
	}
	go s.p2p.SendMessage(context.Background(), netMsg, opts...)
	return qc, nil
}

// handleReceivedMsg used to process msg received from network
func (s *Smr) handleReceivedMsg(msg *p2p_pb.XuperMessage) {
	s.slog.Info("handleReceivedMsg receive msg", "logid",
		msg.GetHeader().GetLogid(), "type", msg.GetHeader().GetType())

	// verify msg
	if !p2p_pb.VerifyDataCheckSum(msg) {
		s.slog.Warn("handleReceivedMsg verify msg data error!", "logid", msg.GetHeader().GetLogid())
		return
	}

	// filter msg from other chain
	if msg.GetHeader().GetBcname() != s.bcname {
		s.slog.Info("handleReceivedMsg msg doesn't from this chain!",
			"logid", msg.GetHeader().GetLogid(), "bcname_from", msg.GetHeader().GetBcname(), "bcname", s.bcname)
		return
	}

	// dispach msg handler
	switch msg.GetHeader().GetType() {
	case p2p_pb.XuperMessage_CHAINED_BFT_NEW_VIEW_MSG:
		go s.handleReceivedNewView(msg)
	case p2p_pb.XuperMessage_CHAINED_BFT_NEW_PROPOSAL_MSG:
		go s.handleReceivedProposal(msg)
	case p2p_pb.XuperMessage_CHAINED_BFT_VOTE_MSG:
		go s.handleReceivedVoteMsg(msg)
	default:
		s.slog.Info("handleReceivedMsg receive unknow type msg")
		return
	}
}

// handleReceivedVoteMsg used to process while receiving vote msg from network
func (s *Smr) handleReceivedVoteMsg(msg *p2p_pb.XuperMessage) error {
	voteMsg := &chainedbft_pb.ChainedBftVoteMessage{}
	if err := proto.Unmarshal(msg.GetData().GetMsgInfo(), voteMsg); err != nil {
		s.slog.Error("handleReceivedVoteMsg Unmarshal msg error",
			"logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}

	if err := s.addVoteMsg(voteMsg); err != nil {
		s.slog.Error("handleReceivedVoteMsg add vote msg error",
			"logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}

	// as a leader, if the num of votes about proposalQC more than (n -f), need to update local status
	if s.checkVoteNum(voteMsg) {
		s.votedView = s.proposalQC.GetViewNumber()
		s.lockedQC = s.generateQC
		s.generateQC = s.proposalQC
	}
	return nil
}

// handleReceivedNewView used to handle new view msg from other replicas
func (s *Smr) handleReceivedNewView(msg *p2p_pb.XuperMessage) error {
	newViewMsg := &chainedbft_pb.ChainedBftPhaseMessage{}
	if err := proto.Unmarshal(msg.GetData().GetMsgInfo(), newViewMsg); err != nil {
		s.slog.Error("handleReceivedNewView Unmarshal msg error",
			"logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}

	if err := s.addViewMsg(newViewMsg); err != nil {
		s.slog.Error("handleReceivedNewView add vote msg error",
			"logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}
	return nil
}

// handleReceivedProposal is the core function of hotstuff. It uesd to change QuorumCerts's phase.
// It will change three previous QuorumCerts's state because hotstuff is a three chained bft.
func (s *Smr) handleReceivedProposal(msg *p2p_pb.XuperMessage) error {
	propMsg := &chainedbft_pb.ChainedBftPhaseMessage{}
	if err := proto.Unmarshal(msg.GetData().GetMsgInfo(), propMsg); err != nil {
		s.slog.Error("handleReceivedProposal Unmarshal msg error",
			"logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}
	propsQC := propMsg.GetProposalQC()
	// Step1: call extenal consensus for chained proposals
	// prePropsQC is the propsQC's ProposalMsg's JustifyQC
	// prePrePropsQC is the prePropsQC's ProposalMsg's JustifyQC
	// prePrePropsQC <- prePropsQC <- propsQC
	prePropsQC, err := s.externalCons.CallPreQc(propsQC)
	if err != nil {
		s.slog.Error("handleReceivedProposal CallPreQc call prePropsQC error", "err", err)
		return err
	}
	prePrePropsQC, err := s.externalCons.CallPreQc(prePropsQC)
	if err != nil {
		s.slog.Error("handleReceivedProposal CallPreQc call prePrePropsQC error", "err", err)
		return err
	}

	// preProposalMsg is the propsQC.ProposalMsg's parent block
	// prePreProposalMsg is the propsQC.ProposalMsg's grandparent block
	preProposalMsg, err := s.externalCons.CallPreProposalMsg(propsQC.GetProposalMsg())
	if err != nil {
		s.slog.Error("handleReceivedProposal CallProposalMsg call preProposalMsg error", "err", err)
		return err
	}
	prePreProposalMsg, err := s.externalCons.CallPrePreProposalMsg(propsQC.GetProposalMsg())
	if err != nil {
		s.slog.Error("handleReceivedProposal CallProposalMsg call prePreProposalMsg error", "err", err)
		return err
	}

	// Step2: judge safety
	ok, err := s.safeProposal(propsQC, prePropsQC)
	if !ok || err != nil {
		s.slog.Error("handleReceivedProposal safeProposal error!", "ok", ok, "error", err)
		return ErrSafeProposal
	}
	// Step3: vote for this proposal
	// TODO: zq sign for this msg
	voteMsg := &chainedbft_pb.ChainedBftVoteMessage{
		ProposalId: propsQC.GetProposalId(),
	}
	// send to leader
	msgBuf, err := proto.Marshal(voteMsg)
	if err != nil {
		s.slog.Error("handleReceivedProposal marshal msg error", "error", err)
		return err
	}

	netMsg, _ := p2p_pb.NewXuperMessage(p2p_pb.XuperMsgVersion3, s.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_VOTE_MSG, msgBuf, p2p_pb.XuperMessage_NONE)

	opts := []p2pv2.MessageOption{
		p2pv2.WithBcName(s.bcname),
		p2pv2.WithTargetPeerAddrs([]string{propMsg.GetSignature().GetAddress()}),
	}
	go s.p2p.SendMessage(context.Background(), netMsg, opts...)

	// Step4: update state
	s.proposalQC = propsQC
	if bytes.Equal(preProposalMsg, prePropsQC.GetProposalMsg()) {
		s.votedView = prePropsQC.GetViewNumber()
		s.generateQC = prePropsQC
	}

	if bytes.Equal(preProposalMsg, prePropsQC.GetProposalMsg()) &&
		bytes.Equal(prePreProposalMsg, prePrePropsQC.GetProposalMsg()) {
		s.lockedQC = prePrePropsQC
	}
	return nil
}
