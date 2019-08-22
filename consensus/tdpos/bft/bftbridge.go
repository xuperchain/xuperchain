package bft

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"
	"github.com/xuperchain/xuperunion/pb"
)

// CbftBridge implements ExternalInterface that chainedbft can communicate with TDPoS
type CbftBridge struct {
}

// CallPreQc call external consensus for the PreQc with the given Qc
//  PreQc is the the given QC's ProposalMsg's JustifyQC
func (cb *CbftBridge) CallPreQc(qc *chainedbft_pb.QuorumCert) (*chainedbft_pb.QuorumCert, error) {
	if qc == nil {
		return nil, fmt.Errorf("invalid params")
	}

	block := &pb.Block{}
	err := proto.Unmarshal(qc.GetProposalMsg(), block)
	if err != nil {
		return nil, err
	}

	// TODO get pre qc from block content

	return nil, nil
}

// CallPreProposalMsg call external consensus for the marshal format of proposalMsg's parent block
func (cb *CbftBridge) CallPreProposalMsg(proposalID []byte) ([]byte, error) {
	return nil, nil
}

// CallPrePreProposalMsg call external consensus for the marshal format of proposalMsg's grandpa's block
func (cb *CbftBridge) CallPrePreProposalMsg([]byte) ([]byte, error) {
	return nil, nil
}

// CallVerifyQc call external consensus for proposalMsg verify with the given QC
func (cb *CbftBridge) CallVerifyQc(*chainedbft_pb.QuorumCert) (bool, error) {
	return true, nil
}

// CallProposalMsgWithProposalID call  external consensus for proposalMsg  with the given ProposalID
func (cb *CbftBridge) CallProposalMsgWithProposalID([]byte) ([]byte, error) {
	return nil, nil
}
