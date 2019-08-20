package smr

import (
	"crypto/ecdsa"
	"errors"
	"sync"

	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/external"
	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/utils"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// Smr is the state of the node
type Smr struct {
	slog log.Logger
	// config is the config of ChainedBft
	config config.Config
	// bcname of ChainedBft instance
	bcname string
	// the node address
	address   string
	publicKey string
	// private key
	privateKey *ecdsa.PrivateKey
	// validates sets, changes with external layer consensus
	validates []*cons_base.CandidateInfo
	// externalCons is the instance that chained bft communicate with
	externalCons external.ExternalInterface
	// cryptoClient is default cryptoclient of chain
	cryptoClient crypto_base.CryptoClient
	// p2p is the network instance
	p2p *p2pv2.P2PServerV2
	// p2pMsgChan is the msg channel registered to network
	p2pMsgChan chan *xuper_p2p.XuperMessage

	// Hotstuff State of this nodes
	// votedView is the last voted view, view changes with chain
	votedView int64
	// proposalQC is the proposalBlock's QC
	proposalQC *chainedbft_pb.QuorumCert
	// generateQC is the proposalBlock's QC, refer to generateBlock's votes
	generateQC *chainedbft_pb.QuorumCert
	// lockedQC is the generateBlock's QC, refer to lockedBlock's votes
	lockedQC *chainedbft_pb.QuorumCert
	// votes of QC in mem, key: prposalID, value: *chainedbft_pb.QCSignInfos
	qcVoteMsgs *sync.Map
	// new view msg gathered from other replicas, key: viewNumber, value: []*chainedbft_pb.ChainedBftPhaseMessage
	newViewMsgs *sync.Map

	// quitCh stop channel
	QuitCh chan bool
}

// addViewMsg check and add new view msg to smr
// 1: check sign of msg
// 2: check if the msg from validate sets replica
func (s *Smr) addViewMsg(msg *chainedbft_pb.ChainedBftPhaseMessage) error {
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

	// add View msg
	v, ok := s.newViewMsgs.Load(msg.GetViewNumber())
	if !ok {
		viewMsgs := []*chainedbft_pb.ChainedBftPhaseMessage{}
		viewMsgs = append(viewMsgs, msg)
		s.newViewMsgs.Store(msg.GetViewNumber(), viewMsgs)
		return nil
	}

	viewMsgs := v.([]*chainedbft_pb.ChainedBftPhaseMessage)
	viewMsgs = append(viewMsgs, msg)
	s.newViewMsgs.Store(msg.GetViewNumber(), viewMsgs)
	return nil
}

// addVoteMsg check and add vote msg to smr
// 1: check sign of msg
// 2: check if the msg from validate sets
func (s *Smr) addVoteMsg(msg *chainedbft_pb.ChainedBftVoteMessage) error {
	// check msg sign
	proposalMsg, err := s.externalCons.CallProposalMsgWithProposalID(msg.GetProposalId())
	if err != nil {
		s.slog.Error("addVoteMsg CallProposalMsgWithProposalID error", "error", err)
		return err
	}
	ok, err := utils.VerifyVoteMsgSign(s.cryptoClient, msg.GetSignature(), proposalMsg)
	if !ok || err != nil {
		s.slog.Error("addVoteMsg VerifyVoteMsgSign error", "ok", ok, "error", err)
		return errors.New("addVoteMsg VerifyVoteMsgSign error")
	}
	// check in ValidateSets
	if !utils.IsInValidateSets(s.validates, msg.GetSignature().GetAddress()) {
		s.slog.Error("addVoteMsg IsInValidateSets error")
		return errors.New("addVoteMsg IsInValidateSets error")
	}

	// add vote msg
	v, ok := s.qcVoteMsgs.Load(string(msg.GetProposalId()))
	if !ok {
		voteMsgs := &chainedbft_pb.QCSignInfos{}
		voteMsgs.QCSignInfos = append(voteMsgs.QCSignInfos, msg.GetSignature())
		s.qcVoteMsgs.Store(string(msg.GetProposalId()), voteMsgs)
		return nil
	}

	voteMsgs := v.(*chainedbft_pb.QCSignInfos)
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
	voteMsgs := v.(*chainedbft_pb.QCSignInfos)

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
