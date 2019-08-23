package smr

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/external"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/utils"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/p2pv2"
	p2p_pb "github.com/xuperchain/xuperunion/p2pv2/pb"
	pb "github.com/xuperchain/xuperunion/pb"
)

var (
	// ErrNewViewNum used to return error new view number
	ErrNewViewNum = errors.New("new view number error")
	// ErrSafeProposal check new proposal error
	ErrSafeProposal = errors.New("check new proposal error")
	// ErrGetVotes get votes error
	ErrGetVotes = errors.New("get votes error")
	// ErrPropsViewNum return proposal view number error
	ErrPropsViewNum = errors.New("proposal view number error")
	// ErrJustifySignNotEnough return justify sign not enough error
	ErrJustifySignNotEnough = errors.New("proposal justify sign not enough error")
	// ErrVerifyVoteSign return verify vote sign error
	ErrVerifyVoteSign = errors.New("verify justify sign error")
	// ErrInValidateSets return in validate sets error
	ErrInValidateSets = errors.New("in validate sets error")
)

// NewSmr return smr instance
func NewSmr(
	cfg *config.Config,
	bcname string,
	address string,
	publicKey string,
	privateKey *ecdsa.PrivateKey,
	validates []*cons_base.CandidateInfo,
	externalCons external.ExternalInterface,
	cryptoClient crypto_base.CryptoClient,
	p2p p2pv2.P2PServer,
	proposalQC,
	generateQC,
	lockedQC *pb.QuorumCert) (*Smr, error) {

	xlog := log.New("module", "smr")
	// set up smr
	//todo: validate params

	smr := &Smr{
		slog:         xlog,
		bcname:       bcname,
		address:      address,
		publicKey:    publicKey,
		privateKey:   privateKey,
		validates:    validates,
		externalCons: externalCons,
		cryptoClient: cryptoClient,
		p2p:          p2p,
		p2pMsgChan:   make(chan *p2p_pb.XuperMessage, cfg.NetMsgChanSize),
		votedView:    generateQC.ViewNumber,
		proposalQC:   proposalQC,
		generateQC:   generateQC,
		lockedQC:     lockedQC,
		QuitCh:       make(chan bool, 1),
	}
	if err := smr.registerToNetwork(); err != nil {
		xlog.Error("smr registerToNetwork error", "error", err)
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

	newViewMsg := &pb.ChainedBftPhaseMessage{
		Type:       pb.QCState_NEW_VIEW,
		ViewNumber: viewNumber,
		Signature: &pb.SignInfo{
			Address:   s.address,
			PublicKey: s.publicKey,
		},
	}

	if preLeader == s.address {
		newViewMsg.JustifyQC = s.generateQC
	}

	newViewMsg, err := utils.MakePhaseMsgSign(s.cryptoClient, s.privateKey, newViewMsg)
	if err != nil {
		s.slog.Error("ProcessNewView MakePhaseMsgSign error", "error", err)
		return err
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

// GetGenerateQC get latest GenerateQC while dominer
func (s *Smr) GetGenerateQC(proposalID []byte) (*pb.QuorumCert, error) {
	return s.generateQC, nil
}

// ProcessProposal used to generate new QuorumCert and broadcast to other replicas
func (s *Smr) ProcessProposal(viewNumber int64, proposalID,
	proposalMsg []byte) (*pb.QuorumCert, error) {

	qc := &pb.QuorumCert{
		ProposalId:  proposalID,
		ProposalMsg: proposalMsg,
		ViewNumber:  viewNumber,
		Type:        pb.QCState_PREPARE,
		SignInfos:   &pb.QCSignInfos{},
	}

	propMsg := &pb.ChainedBftPhaseMessage{
		Type:       pb.QCState_PREPARE,
		ViewNumber: viewNumber,
		ProposalQC: qc,
		Signature: &pb.SignInfo{
			Address:   s.address,
			PublicKey: s.publicKey,
		},
	}

	propMsg, err := utils.MakePhaseMsgSign(s.cryptoClient, s.privateKey, propMsg)
	if err != nil {
		s.slog.Error("ProcessProposal MakePhaseMsgSign error", "error", err)
		return nil, err
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
	voteMsg := &pb.ChainedBftVoteMessage{}
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
	if s.checkVoteNum(voteMsg.GetProposalId()) {
		s.votedView = s.proposalQC.GetViewNumber()
		s.lockedQC = s.generateQC
		s.generateQC = s.proposalQC
		v, ok := s.qcVoteMsgs.Load(string(s.proposalQC.GetProposalId()))
		if !ok {
			s.slog.Error("handleReceivedVoteMsg get votes error")
			return ErrGetVotes
		}
		s.generateQC.SignInfos = v.(*pb.QCSignInfos)
	}
	return nil
}

// handleReceivedNewView used to handle new view msg from other replicas
func (s *Smr) handleReceivedNewView(msg *p2p_pb.XuperMessage) error {
	newViewMsg := &pb.ChainedBftPhaseMessage{}
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
	propMsg := &pb.ChainedBftPhaseMessage{}
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
	voteMsg := &pb.ChainedBftVoteMessage{
		ProposalId: propsQC.GetProposalId(),
		Signature: &pb.SignInfo{
			Address:   s.address,
			PublicKey: s.publicKey,
		},
	}
	_, err = utils.MakeVoteMsgSign(s.cryptoClient, s.privateKey, voteMsg.GetSignature(), propsQC.GetProposalId())
	if err != nil {
		s.slog.Error("ProcessProposal MakeVoteMsgSign error", "error", err)
		return err
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

// addViewMsg check and add new view msg to smr
// 1: check sign of msg
// 2: check if the msg from validate sets replica
func (s *Smr) addViewMsg(msg *pb.ChainedBftPhaseMessage) error {
	// check msg sign
	ok, err := utils.VerifyPhaseMsgSign(s.cryptoClient, msg)
	if !ok || err != nil {
		s.slog.Error("addViewMsg VerifyPhaseMsgSign error", "ok", ok, "error", err)
		return errors.New("addViewMsg VerifyPhaseMsgSign error")
	}
	// check whether view outdate
	if msg.GetViewNumber() < s.votedView {
		s.slog.Error("addViewMsg view outdate", "votedView", s.votedView, "viewRecivied", msg.GetViewNumber())
		return errors.New("addViewMsg view outdate")
	}

	// check in ValidateSets
	if !utils.IsInValidateSets(s.validates, msg.GetSignature().GetAddress()) {
		s.slog.Error("addViewMsg checkValidateSets error")
		return errors.New("addViewMsg checkValidateSets error")
	}
	// add JustifyQC
	if msg.GetJustifyQC() != nil {
		s.slog.Info("addViewMsg GetJustifyQC not nil", "GetJustifyQC.SignInfos", msg.GetJustifyQC().GetSignInfos())
		s.qcVoteMsgs.LoadOrStore(string(msg.GetJustifyQC().GetProposalId()), msg.GetJustifyQC().GetSignInfos())
	}

	// add View msg
	v, ok := s.newViewMsgs.Load(msg.GetViewNumber())
	if !ok {
		viewMsgs := []*pb.ChainedBftPhaseMessage{}
		viewMsgs = append(viewMsgs, msg)
		s.newViewMsgs.Store(msg.GetViewNumber(), viewMsgs)
		return nil
	}

	viewMsgs := v.([]*pb.ChainedBftPhaseMessage)
	viewMsgs = append(viewMsgs, msg)
	s.newViewMsgs.Store(msg.GetViewNumber(), viewMsgs)
	return nil
}

// addVoteMsg check and add vote msg to smr
// 1: check sign of msg
// 2: check if the msg from validate sets
func (s *Smr) addVoteMsg(msg *pb.ChainedBftVoteMessage) error {
	// check in ValidateSets
	if !utils.IsInValidateSets(s.validates, msg.GetSignature().GetAddress()) {
		s.slog.Error("addVoteMsg IsInValidateSets error")
		return ErrInValidateSets
	}

	// check msg sign
	ok, err := utils.VerifyVoteMsgSign(s.cryptoClient, msg.GetSignature(), msg.GetProposalId())
	if !ok || err != nil {
		s.slog.Error("addVoteMsg VerifyVoteMsgSign error", "ok", ok, "error", err)
		return ErrVerifyVoteSign
	}

	// add vote msg
	v, ok := s.qcVoteMsgs.Load(string(msg.GetProposalId()))
	if !ok {
		voteMsgs := &pb.QCSignInfos{}
		voteMsgs.QCSignInfos = append(voteMsgs.QCSignInfos, msg.GetSignature())
		s.qcVoteMsgs.Store(string(msg.GetProposalId()), voteMsgs)
		return nil
	}

	voteMsgs := v.(*pb.QCSignInfos)
	if utils.CheckIsVoted(voteMsgs, msg.GetSignature()) {
		s.slog.Error("addVoteMsg CheckIsVoted error, this address have voted")
		return errors.New("addVoteMsg CheckIsVoted error")
	}
	voteMsgs.QCSignInfos = append(voteMsgs.QCSignInfos, msg.GetSignature())
	s.qcVoteMsgs.Store(string(msg.GetProposalId()), voteMsgs)
	return nil
}

// checkVoteNum leader will check whether the vote nums more than (n-f)
func (s *Smr) checkVoteNum(proposalID []byte) bool {
	v, ok := s.qcVoteMsgs.Load(string(proposalID))
	if !ok {
		s.slog.Error("smr checkVoteNum error, voteMsgs not found!")
		return false
	}
	voteMsgs := v.(*pb.QCSignInfos)

	if len(voteMsgs.GetQCSignInfos()) > (len(s.validates)-1)*2/3 {
		return true
	}
	return false
}

// UpdateValidateSets update current ValidateSets by ex
func (s *Smr) UpdateValidateSets(validates []*cons_base.CandidateInfo) error {
	s.validates = validates
	return nil
}
