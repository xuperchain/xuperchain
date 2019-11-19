package smr

import (
	"encoding/hex"

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
	s.slog.Warn("safeProposal proposal justify sign", "autual", len(justifySigns), "require", (len(s.validates)-1)*2/3)
	if len(justifySigns) <= (len(s.validates)-1)*2/3 {
		return false, ErrJustifySignNotEnough
	}
	// verify justify sign
	for _, v := range justifySigns {
		if !utils.IsInValidateSets(s.validates, v.GetAddress()) {
			s.slog.Error("safeProposal IsInValidateSets error")
			return false, ErrInValidateSets
		}

		ok, err := utils.VerifyVoteMsgSign(s.cryptoClient, v, justify.GetProposalId())
		if !ok || err != nil {
			s.slog.Error("safeProposal VerifyVoteMsgSign error", "ok", ok, "error", err)
			return false, ErrVerifyVoteSign
		}
	}
	return true, nil
}
