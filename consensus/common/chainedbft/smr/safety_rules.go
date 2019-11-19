package smr

import (
	"encoding/hex"

	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/utils"
	"github.com/xuperchain/xuperunion/pb"
)

// safeProposal make sure whether Proposal is safe
// 1 check the proposal' view number
// 2 verify justify's votes
// 3 external consensus make sure whether the proposalMsg is safe
func (s *Smr) safeProposal(propsQC, justify *pb.QuorumCert) (bool, error) {
	// step1: liveness rule
	// the proposQC's view number need to more than lockedQC
	if propsQC.GetViewNumber() < s.lockedQC.GetViewNumber() {
		s.slog.Error("safeProposal liveness rule error",
			"propsQC.ViewNum", propsQC.GetViewNumber(), "lockedQC.ViewNum", s.lockedQC.GetViewNumber())
		return false, ErrPropsViewNum
	}
	// step2: verify justify's votes
	// verify justify sign number
	if justify == nil {
		s.slog.Warn("safeProposal justify is nil")
		return s.externalCons.CallVerifyQc(propsQC)
	}

	if ok, err := s.IsQuorumCertValidate(justify); !ok || err != nil {
		s.slog.Error("safeProposal IsQuorumCertValidate error", "ok", ok, "error", err)
		return false, err
	}
	// step3: call external consensus verify proposalMsg
	return s.externalCons.CallVerifyQc(propsQC)
}

// IsQuorumCertValidate return whether QC is validated
func (s *Smr) IsQuorumCertValidate(justify *pb.QuorumCert) (bool, error) {
	s.slog.Debug("IsQuorumCertValidate", "justify.ProposalId", hex.EncodeToString(justify.GetProposalId()))
	if justify == nil || justify.GetSignInfos() == nil || justify.GetProposalId() == nil {
		return false, ErrParams
	}
	justifySigns := justify.GetSignInfos().GetQCSignInfos()
	// verify justify sign
	if justify.GetViewNumber() == s.vscView {
		return s.verifyVotes(justifySigns, s.preValidates, justify.GetProposalId())
	}
	return s.verifyVotes(justifySigns, s.validates, justify.GetProposalId())
}

// verifyVotes verify QC sign
func (s *Smr) verifyVotes(signs []*pb.SignInfo, validateSets []*cons_base.CandidateInfo, proposalID []byte) (bool, error) {
	s.slog.Trace("safeProposal proposal justify sign", "autual", len(signs), "require", (len(validateSets)-1)*2/3)
	if len(signs) <= (len(validateSets)-1)*2/3 {
		return false, ErrJustifySignNotEnough
	}
	for _, v := range signs {
		if !utils.IsInValidateSets(validateSets, v.GetAddress()) {
			s.slog.Error("verifyVotes IsInValidateSets error")
			return false, ErrInValidateSets
		}
		ok, err := utils.VerifyVoteMsgSign(s.cryptoClient, v, proposalID)
		if !ok || err != nil {
			s.slog.Error("verifyVotes VerifyVoteMsgSign error", "ok", ok, "error", err)
			return false, ErrVerifyVoteSign
		}
	}
	return true, nil
}
