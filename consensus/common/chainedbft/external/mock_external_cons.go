package external

import pb "github.com/xuperchain/xuperunion/pb"

// MockExternalConsensus mock the ExternalInterface
// Used in unit tests
type MockExternalConsensus struct {
}

// CallPreQc is the the given QC's ProposalMsg's JustifyQC
func (mec *MockExternalConsensus) CallPreQc(qc *pb.QuorumCert) (*pb.QuorumCert, error) {
	return nil, nil
}

// CallPreProposalMsg call external consensus for the marshal format of proposalMsg's parent block
func (mec *MockExternalConsensus) CallPreProposalMsg(proposalMsg []byte) ([]byte, error) {
	return nil, nil
}

// CallPrePreProposalMsg call external consensus for the marshal format of proposalMsg's grandpa's block
func (mec *MockExternalConsensus) CallPrePreProposalMsg(proposalMsg []byte) ([]byte, error) {
	return nil, nil
}

// CallVerifyQc call external consensus for proposalMsg verify with the given QC
func (mec *MockExternalConsensus) CallVerifyQc(qc *pb.QuorumCert) (bool, error) {
	return true, nil
}

// CallProposalMsgWithProposalID call  external consensus for proposalMsg  with the given ProposalID
func (mec *MockExternalConsensus) CallProposalMsgWithProposalID(proposalID []byte) ([]byte, error) {
	return nil, nil
}
