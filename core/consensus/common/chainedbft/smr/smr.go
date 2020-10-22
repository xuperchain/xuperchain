package smr

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"

	"github.com/golang/protobuf/proto"
	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperchain/core/consensus/base"
	"github.com/xuperchain/xuperchain/core/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperchain/core/consensus/common/chainedbft/external"
	"github.com/xuperchain/xuperchain/core/consensus/common/chainedbft/utils"
	crypto_base "github.com/xuperchain/xuperchain/core/crypto/client/base"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	p2p_pb "github.com/xuperchain/xuperchain/core/p2p/pb"
	pb "github.com/xuperchain/xuperchain/core/pb"
)

const (
	// DefaultNetMsgChanSize is the default size of network msg channel
	DefaultNetMsgChanSize = 1000
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
	// ErrCheckDataSum return check data sum error
	ErrCheckDataSum = errors.New("check data sum error")
	// ErrParams return params error
	ErrParams = errors.New("params error")
	// ErrCallPreQcStatus return call pre qc status error
	ErrCallPreQcStatus = errors.New("call pre qc status error")
	// ErrGetLocalProposalQC return LocalProposalQC error
	ErrGetLocalProposalQC = errors.New("get local proposalQC error")
)

// NewSmr return smr instance
func NewSmr(
	slog log.Logger,
	cfg *config.Config,
	bcname string,
	address string,
	publicKey string,
	privateKey *ecdsa.PrivateKey,
	validates []*cons_base.CandidateInfo,
	externalCons external.ExternalInterface,
	cryptoClient crypto_base.CryptoClient,
	p2p p2p_base.P2PServer,
	proposalQC,
	generateQC,
	lockedQC *pb.QuorumCert,
	effectiveDelay int64) (*Smr, error) {

	// set up smr
	smr := &Smr{
		slog:           slog,
		bcname:         bcname,
		address:        address,
		publicKey:      publicKey,
		privateKey:     privateKey,
		preValidates:   []*cons_base.CandidateInfo{},
		validates:      validates,
		externalCons:   externalCons,
		cryptoClient:   cryptoClient,
		p2p:            p2p,
		p2pMsgChan:     make(chan *p2p_pb.XuperMessage, DefaultNetMsgChanSize),
		subscribeList:  []p2p_base.Subscriber{},
		localProposal:  &sync.Map{},
		qcVoteMsgs:     &sync.Map{},
		newViewMsgs:    &sync.Map{},
		effectiveDelay: effectiveDelay,
		lk:             &sync.Mutex{},
		QuitCh:         make(chan bool, 1),
	}
	if err := smr.updateQcStatus(proposalQC, generateQC, lockedQC); err != nil {
		slog.Error("smr updateQcStatus error", "error", err)
		return nil, err
	}
	return smr, nil
}

// RegisterToNetwork register msg handler to p2p network
func (s *Smr) RegisterToNetwork() error {
	s.lk.Lock()
	defer s.lk.Unlock()
	subNewView, err := s.p2p.Register(s.p2p.NewSubscriber(s.p2pMsgChan,
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_VIEW_MSG, nil, "", s.slog))
	if err != nil {
		return err
	}
	s.subscribeList = append(s.subscribeList, subNewView)

	subNewProposal, err := s.p2p.Register(s.p2p.NewSubscriber(s.p2pMsgChan,
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_PROPOSAL_MSG, nil, "", s.slog))
	if err != nil {
		return err
	}
	s.subscribeList = append(s.subscribeList, subNewProposal)

	subVote, err := s.p2p.Register(s.p2p.NewSubscriber(s.p2pMsgChan,
		p2p_pb.XuperMessage_CHAINED_BFT_VOTE_MSG, nil, "", s.slog))
	if err != nil {
		return err
	}
	s.subscribeList = append(s.subscribeList, subVote)
	return nil
}

// UnRegisterToNetwork unregister msg handler to p2p network
func (s *Smr) UnRegisterToNetwork() error {
	s.lk.Lock()
	defer s.lk.Unlock()
	for i := 0; i < len(s.subscribeList); i++ {
		s.p2p.UnRegister(s.subscribeList[i])
	}
	s.subscribeList = []p2p_base.Subscriber{}
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
			s.stop()
			return
		}
	}
}

// stop used to stop smr instance
func (s *Smr) stop() {
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

	// 作为上一个Leader，会将当前自己的qc(上一个区块)写入这个NewView的上一个qc发出去
	if preLeader == s.address {
		gQC, _ := s.GetGenerateQC()
		newViewMsg.JustifyQC = gQC
	}

	newViewMsg, err := utils.MakePhaseMsgSign(s.cryptoClient, s.privateKey, newViewMsg)
	if err != nil {
		s.slog.Error("ProcessNewView MakePhaseMsgSign error", "error", err)
		return err
	}

	msgBuf, err := proto.Marshal(newViewMsg)
	if err != nil {
		s.slog.Error("ProcessNewView marshal msg error", "error", err)
		return err
	}

	netMsg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion3, s.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_VIEW_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	opts := []p2p_base.MessageOption{
		p2p_base.WithBcName(s.bcname),
		p2p_base.WithTargetPeerAddrs([]string{s.getAddressPeerURL(leader)}),
	}
	go s.p2p.SendMessage(context.Background(), netMsg, opts...)
	return nil
}

// GetGenerateQC get latest GenerateQC while dominer
func (s *Smr) GetGenerateQC() (*pb.QuorumCert, error) {
	res := s.generateQC
	if res != nil {
		res.ProposalMsg = nil
		if len(res.SignInfos.GetQCSignInfos()) == 0 {
			v, ok := s.qcVoteMsgs.Load(string(res.GetProposalId()))
			if !ok {
				s.slog.Error("handleReceivedVoteMsg get votes error")
				return nil, ErrGetVotes
			}
			res.SignInfos = v.(*pb.QCSignInfos)
		}
	}
	s.slog.Info("GetGenerateQC res", "ProposalId", hex.EncodeToString(res.GetProposalId()),
		"lockedQCId", hex.EncodeToString(s.lockedQC.GetProposalId()))
	return res, nil
}

// ProcessProposal used to generate new QuorumCert and broadcast to other replicas
func (s *Smr) ProcessProposal(viewNumber int64, proposalID, proposalMsg []byte, validatesInfos []*cons_base.CandidateInfo) (*pb.QuorumCert, error) {
	qc := &pb.QuorumCert{
		ProposalId:  proposalID,
		ProposalMsg: proposalMsg,
		ViewNumber:  viewNumber,
		Type:        pb.QCState_PREPARE,
		SignInfos:   &pb.QCSignInfos{},
	}
	s.addLocalProposal(qc)
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

	netMsg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion3, s.bcname, "",
		p2p_pb.XuperMessage_CHAINED_BFT_NEW_PROPOSAL_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	s.slog.Info("ProcessProposal: Send proposal msg", "netMsg", netMsg)
	opts := []p2p_base.MessageOption{
		p2p_base.WithBcName(s.bcname),
		p2p_base.WithTargetPeerAddrs(s.getReplicasURL(validatesInfos)),
	}
	go s.p2p.SendMessage(context.Background(), netMsg, opts...)
	return qc, nil
}

// handleReceivedMsg used to process msg received from network
func (s *Smr) handleReceivedMsg(msg *p2p_pb.XuperMessage) error {
	s.slog.Info("handleReceivedMsg receive msg", "logid",
		msg.GetHeader().GetLogid(), "type", msg.GetHeader().GetType())

	// verify msg
	if !p2p_base.VerifyDataCheckSum(msg) {
		s.slog.Warn("handleReceivedMsg verify msg data error!", "logid", msg.GetHeader().GetLogid())
		return ErrCheckDataSum
	}

	// filter msg from other chain
	if msg.GetHeader().GetBcname() != s.bcname {
		s.slog.Warn("handleReceivedMsg msg doesn't from this chain!",
			"logid", msg.GetHeader().GetLogid(), "bcname_from", msg.GetHeader().GetBcname(), "bcname", s.bcname)
		return nil
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
		return nil
	}
	return nil
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
		v, ok := s.localProposal.Load(string(voteMsg.GetProposalId()))
		if !ok {
			s.slog.Error("checkVoteNum load proposQC error")
			return ErrGetLocalProposalQC
		}
		proposQC := v.(*pb.QuorumCert)
		// 变更QC前需要确定当前收到的proposalQC是s.lockedQC的扩展且高度更高
		if proposQC.GetViewNumber() <= s.lockedQC.GetViewNumber() {
			s.slog.Error("handleReceivedVoteMsg proposalQC too old", "proposQC=", proposQC.GetViewNumber(),
				"s.lockedQC=", s.lockedQC.GetViewNumber(), "s.generateQC=",
				s.generateQC.GetViewNumber(), "s.proposalQC=", s.proposalQC.GetViewNumber())
			return ErrGetVotes
		}
		s.slog.Debug("updateQcStatus when handleReceivedVoteMsg.")
		s.updateQcStatus(nil, proposQC, s.generateQC)
		v, ok = s.qcVoteMsgs.Load(string(voteMsg.GetProposalId()))
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
	// Step1: call extenal consensus for prePropsQC
	// prePropsQC is the propsQC's ProposalMsg's JustifyQC
	// prePrePropsQC <- prePropsQC <- propsQC
	propsQC := propMsg.GetProposalQC()
	s.slog.Info("handleReceivedProposal propsQC", "propsQC", propsQC)

	prePropsQC, isFirstProposal, err := s.callPreQcWithStatus(propsQC)
	if err != nil {
		s.slog.Error("handleReceivedProposal call prePropsQC error", "logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}

	// Step2: judge safety
	ok, err := s.safeProposal(propsQC, prePropsQC)
	if !ok || err != nil {
		s.slog.Error("handleReceivedProposal safeProposal error!", "logid", msg.GetHeader().GetLogid(), "ok", ok, "error", err)
		return ErrSafeProposal
	}

	// Step3: vote justify
	err = s.voteProposal(propsQC, propMsg.GetSignature().GetAddress(), msg.GetHeader().GetLogid())
	if err != nil {
		s.slog.Error("handleReceivedProposal voteProposal error", "logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}

	// Step4: update state as a preLeader
	if prePropsQC != nil && bytes.Equal(prePropsQC.GetProposalId(), s.generateQC.GetProposalId()) {
		s.slog.Info("handleReceivedProposal as the preleader, update propsQC.")
		s.updateQcStatus(propsQC, s.generateQC, s.lockedQC)
		return nil
	}

	// propsQC is the first QC
	if isFirstProposal {
		s.updateQcStatus(propsQC, nil, nil)
		return nil
	}

	// call extenal consensus for prePrePropsQC
	// prePrePropsQC is the prePropsQC's ProposalMsg's JustifyQC
	prePrePropsQC, isFirstProposal, err := s.callPreQcWithStatus(prePropsQC)
	if err != nil {
		s.slog.Error("handleReceivedProposal call prePrePropsQC error", "logid", msg.GetHeader().GetLogid(), "error", err)
		return err
	}
	s.updateQcStatus(propsQC, prePropsQC, prePrePropsQC)
	return nil
}

// callPreQcWithStatus call externel consensus for preQc status
func (s *Smr) callPreQcWithStatus(qc *pb.QuorumCert) (*pb.QuorumCert, bool, error) {
	ok, err := s.externalCons.IsFirstProposal(qc)
	s.slog.Trace("callPreQcWithStatus IsFirstProposal status",
		"proposalId", hex.EncodeToString(qc.GetProposalId()), "ok", ok, "err", err)
	if ok || err != nil {
		return nil, ok, err
	}

	prePropsQC, err := s.externalCons.CallPreQc(qc)
	s.slog.Info("callPreQcWithStatus get prePropsQC", "prePropsQC", prePropsQC)
	if err != nil {
		s.slog.Error("callPreQcWithStatus CallPreQc call prePropsQC error", "err", err)
		return nil, false, err
	}
	prePropslMsg, err := s.externalCons.CallProposalMsgWithProposalID(prePropsQC.GetProposalId())
	if err != nil {
		s.slog.Error("callPreQcWithStatus CallPreQc call prePropsQC ProposalMsg error", "err", err)
		return nil, false, err
	}
	prePropsQC.ProposalMsg = prePropslMsg
	return prePropsQC, false, nil
}

// updateQcStatus upstate QC status with given qc's
func (s *Smr) updateQcStatus(proposalQC, generateQC, lockedQC *pb.QuorumCert) error {
	s.lk.Lock()
	if generateQC == nil {
		s.votedView = 0
	} else {
		s.votedView = generateQC.GetViewNumber()
	}
	s.lockedQC = lockedQC
	s.generateQC = generateQC
	s.proposalQC = proposalQC
	if s.generateQC == s.proposalQC {
		s.proposalQC = nil
	}
	s.slog.Info("updateQcStatus result", "proposalQCId", hex.EncodeToString(s.proposalQC.GetProposalId()))
	s.slog.Info("updateQcStatus result", "generateQCId", hex.EncodeToString(s.generateQC.GetProposalId()))
	s.slog.Info("updateQcStatus result", "lockedQCId", hex.EncodeToString(s.lockedQC.GetProposalId()))
	s.lk.Unlock()
	return nil
}

// voteProposal vote for this proposal
func (s *Smr) voteProposal(propsQC *pb.QuorumCert, voteTo, logid string) error {
	voteMsg := &pb.ChainedBftVoteMessage{
		ProposalId: propsQC.GetProposalId(),
		Signature: &pb.SignInfo{
			Address:   s.address,
			PublicKey: s.publicKey,
		},
	}
	_, err := utils.MakeVoteMsgSign(s.cryptoClient, s.privateKey, voteMsg.GetSignature(), propsQC.GetProposalId())
	if err != nil {
		s.slog.Error("voteProposal MakeVoteMsgSign error", "error", err)
		return err
	}

	// send to leader
	msgBuf, err := proto.Marshal(voteMsg)
	if err != nil {
		s.slog.Error("voteProposal marshal msg error", "error", err)
		return err
	}
	netMsg, _ := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion3, s.bcname, logid,
		p2p_pb.XuperMessage_CHAINED_BFT_VOTE_MSG, msgBuf, p2p_pb.XuperMessage_NONE)
	s.slog.Trace("voteProposal", "msg", netMsg, "voteTo", voteTo, "logid", logid)

	opts := []p2p_base.MessageOption{
		p2p_base.WithBcName(s.bcname),
		p2p_base.WithTargetPeerAddrs([]string{s.getAddressPeerURL(voteTo)}),
	}
	go s.p2p.SendMessage(context.Background(), netMsg, opts...)
	return nil
}

func (s *Smr) getValidates(viewNumer int64) []*cons_base.CandidateInfo {
	s.slog.Error("getValidates", "effectiveDelay", s.effectiveDelay, "s.vscView", s.vscView)
	// 根据不同共识的生效延时，判断当前view是否需要使用哪个验证集合，viewNumer为新视图，vscView为上次变更候选人视图
	if s.effectiveDelay > 0 && viewNumer == s.vscView+s.effectiveDelay {
		for _, v := range s.preValidates {
			s.slog.Debug("PreValidates Check Set", "Address", v.Address)
		}
		return s.preValidates
	}
	for _, v := range s.validates {
		s.slog.Debug("Validates Check Set", "Address", v.Address)
	}
	return s.validates
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
	validates := s.getValidates(msg.GetViewNumber())
	if !utils.IsInValidateSets(validates, msg.GetSignature().GetAddress()) {
		s.slog.Debug("addViewMsg checkValidateSets error", "msg.viewNumber", msg.GetViewNumber(), "msg.GetAddress()", msg.GetSignature().GetAddress())
		return errors.New("addViewMsg checkValidateSets error")
	}
	// add JustifyQC
	justify := msg.GetJustifyQC()
	// 收到preLeader的NewViewMSG，则当前leader变更QC
	if justify != nil {
		s.slog.Info("addViewMsg GetJustifyQC not nil", "justifyId", hex.EncodeToString(justify.GetProposalId()),
			"proposalId", hex.EncodeToString(s.proposalQC.GetProposalId()), "GetJustifyQC.SignInfos", justify.GetSignInfos())
		if s.proposalQC != nil && bytes.Equal(s.proposalQC.GetProposalId(), justify.GetProposalId()) {
			if ok, _ := s.IsQuorumCertValidate(justify); ok {
				s.slog.Info("addViewMsg update local as a new leader")
				s.updateQcStatus(nil, s.proposalQC, s.generateQC)
				s.qcVoteMsgs.Store(string(justify.GetProposalId()), justify.GetSignInfos())
				viewMsgs := []*pb.ChainedBftPhaseMessage{}
				viewMsgs = append(viewMsgs, msg)
				s.newViewMsgs.Store(msg.GetViewNumber(), viewMsgs)
			}
		}
	}
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
	s.slog.Info("checkVoteNum", "actual", len(voteMsgs.GetQCSignInfos()), "require", (len(s.validates)-1)*2/3)
	if len(voteMsgs.GetQCSignInfos()) > (len(s.validates)-1)*2/3 {
		s.slog.Info("checkVoteNum", "res", true, "proposalID", hex.EncodeToString(proposalID))
		return true
	}
	s.slog.Info("checkVoteNum", "res", false, "proposalID", hex.EncodeToString(proposalID))
	return false
}

// UpdateValidateSets update current ValidateSets by ex
func (s *Smr) UpdateValidateSets(validates []*cons_base.CandidateInfo) error {
	s.lk.Lock()
	defer s.lk.Unlock()
	if !utils.IsInValidateSets(s.validates, s.address) {
		s.vscView = s.votedView + 1
	} else {
		s.vscView = s.votedView
	}
	s.preValidates = s.validates
	s.validates = validates
	preStr, _ := json.Marshal(s.preValidates)
	curStr, _ := json.Marshal(s.validates)
	s.slog.Trace("UpdateValidateSets", "preValidates", string(preStr), "validates", string(curStr))
	s.slog.Trace("UpdateValidateSets", "s.vscView", s.vscView, "s.votedView", s.votedView)
	return nil
}

// getReplicasURL return validates urls
func (s *Smr) getReplicasURL(validates []*cons_base.CandidateInfo) []string {
	validateURL := []string{}
	for _, v := range validates {
		if v.Address == s.address {
			continue
		}
		validateURL = append(validateURL, v.PeerAddr)
	}
	s.slog.Trace("getReplicasURL result", "validateURL", validateURL)
	return validateURL
}

// getAddressPeerURL get address peer url
// todo: zq consider validate sets changes
func (s *Smr) getAddressPeerURL(address string) string {
	for _, v := range s.validates {
		if v.Address == address {
			return v.PeerAddr
		}
	}
	return ""
}

// addLocalProposal add local proposal
func (s *Smr) addLocalProposal(qc *pb.QuorumCert) {
	s.localProposal.Store(string(qc.GetProposalId()), qc)
}

// UpdateSmrState 更新smr状态, 解决bpm check IsLastViewConfirmed的问题
func (s *Smr) UpdateSmrState(generateQC *pb.QuorumCert) {
	s.slog.Info("UpdateSmrState and update ProposalQCId after block confirmed", "generateQC", generateQC.GetViewNumber(),
		"s.generateQC", s.generateQC.GetViewNumber(), "s.proposalQC", s.proposalQC.GetViewNumber())
	// 此处需要确定这个generateQC一定比本地的s.generateQC后才行
	if generateQC.GetViewNumber() == s.generateQC.GetViewNumber()+1 && s.proposalQC.GetViewNumber() == generateQC.GetViewNumber()+1 {
		s.updateQcStatus(s.proposalQC, generateQC, s.generateQC)
		return
	}
	if generateQC.GetViewNumber() == s.generateQC.GetViewNumber()+1 {
		s.updateQcStatus(nil, generateQC, s.generateQC)
	}
}
