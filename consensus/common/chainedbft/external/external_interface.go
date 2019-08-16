package external

import chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"

// ExternalInterface is the interface that chainedbft can communicate with external interface
// external consensus need to implements this.
type ExternalInterface interface {
	// CallPreQc call external consensus for the PreQc with the given Qc
	//  PreQc is the the given QC's ProposalMsg's JustifyQC
	CallPreQc(*chainedbft_pb.QuorumCert) (*chainedbft_pb.QuorumCert, error)

	// CallProposalMsg call external consensus for the marshal format of proposalMsg's parent block
	CallPreProposalMsg([]byte) ([]byte, error)
	// CallPrePreProposalMsg call external consensus for the marshal format of proposalMsg's grandpa's block
	CallPrePreProposalMsg([]byte) ([]byte, error)

	// CallVerifyQc call external consensus for proposalMsg verify with the given QC
	CallVerifyQc(*chainedbft_pb.QuorumCert) (bool, error)
}
