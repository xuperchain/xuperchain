package smr

import chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"

// safeProposal make sure whether Proposal is safe
// 1 smr check whether the proposal is true
// 2 external consensus make sure whether the proposalMsg is safe
// TODO: zq
func (s *Smr) safeProposal(propsQC, prePropsQC *chainedbft_pb.QuorumCert) (bool, error) {
	return false, nil
}
