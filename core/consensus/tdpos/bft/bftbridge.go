package bft

import (
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/consensus/base"
	"github.com/xuperchain/xuperchain/core/ledger"
	"github.com/xuperchain/xuperchain/core/pb"
)

// CbftBridge implements ExternalInterface that chainedbft can communicate with TDPoS
type CbftBridge struct {
	bcname    string
	ledger    *ledger.Ledger
	log       log.Logger
	cons      base.ConsensusInterface
	paceMaker *DPoSPaceMaker
}

// NewCbftBridge create new instance of CbftBridge
func NewCbftBridge(bcname string, ledger *ledger.Ledger, xlog log.Logger, cons base.ConsensusInterface) *CbftBridge {
	return &CbftBridge{
		bcname: bcname,
		ledger: ledger,
		log:    xlog,
		cons:   cons,
	}
}

// SetPaceMaker set pacemaker
func (cb *CbftBridge) SetPaceMaker(paceMaker *DPoSPaceMaker) {
	cb.paceMaker = paceMaker
}

// CallPreQc call external consensus for the PreQc with the given Qc
//  PreQc is the the given QC's ProposalMsg's JustifyQC
func (cb *CbftBridge) CallPreQc(qc *pb.QuorumCert) (*pb.QuorumCert, error) {
	if qc == nil {
		return nil, fmt.Errorf("invalid params")
	}

	block := &pb.Block{}
	err := proto.Unmarshal(qc.GetProposalMsg(), block)
	if err != nil {
		return nil, err
	}

	if block.GetBlock() == nil {
		return nil, fmt.Errorf("CallPreQC block content is not complete")
	}

	return block.Block.Justify, nil
}

// CallPreProposalMsg call external consensus for the marshal format of proposalMsg's parent block
func (cb *CbftBridge) CallPreProposalMsg(proposalMsg []byte) ([]byte, error) {
	block := &pb.Block{}
	err := proto.Unmarshal(proposalMsg, block)
	if err != nil || block.GetBlock() == nil {
		cb.log.Warn("CallPreProposalMsg cannot unmarshal block", "block", proposalMsg)
		return nil, err
	}
	preHash := block.GetBlock().GetPreHash()
	preBlockContent, err := cb.ledger.QueryBlock(preHash)
	if err != nil {
		cb.log.Warn("CallPreProposalMsg cannot found pre block", "block",
			hex.EncodeToString(preHash))
		return nil, err
	}
	preBlock := &pb.Block{
		Block:   preBlockContent,
		Blockid: preBlockContent.GetBlockid(),
		Bcname:  cb.bcname,
	}
	msg, err := proto.Marshal(preBlock)
	if err != nil {
		cb.log.Warn("CallPreProposalMsg marshal data failed", "block",
			hex.EncodeToString(preHash))
		return nil, err
	}
	return msg, nil
}

// CallPrePreProposalMsg call external consensus for the marshal format of proposalMsg's grandpa's block
func (cb *CbftBridge) CallPrePreProposalMsg(proposalMsg []byte) ([]byte, error) {
	block := &pb.Block{}
	err := proto.Unmarshal(proposalMsg, block)
	if err != nil || block.GetBlock() == nil {
		cb.log.Warn("CallPrePreProposalMsg cannot unmarshal block", "block", proposalMsg)
		return nil, err
	}

	// get the previous block of current proposal message
	preHash := block.GetBlock().GetPreHash()
	preBlockContent, err := cb.ledger.QueryBlock(preHash)
	if err != nil {
		cb.log.Warn("CallPrePreProposalMsg cannot found previous block", "block",
			hex.EncodeToString(preHash))
		return nil, err
	}

	// get the previous block of current proposal message
	penultimateBlock, err := cb.ledger.QueryBlock(preBlockContent.GetPreHash())
	if err != nil {
		cb.log.Warn("CallPrePreProposalMsg cannot found penultimate block", "block",
			hex.EncodeToString(preHash))
		return nil, err
	}
	penulBlock := &pb.Block{
		Block:   penultimateBlock,
		Blockid: penultimateBlock.GetBlockid(),
		Bcname:  cb.bcname,
	}
	msg, err := proto.Marshal(penulBlock)
	if err != nil {
		cb.log.Warn("CallPrePreProposalMsg marshal data failed", "block",
			hex.EncodeToString(preBlockContent.GetPreHash()))
		return nil, err
	}
	return msg, nil
}

// CallVerifyQc call external consensus for proposalMsg verify with the given QC
func (cb *CbftBridge) CallVerifyQc(qc *pb.QuorumCert) (bool, error) {
	if qc == nil || qc.GetProposalMsg() == nil {
		return false, fmt.Errorf("invalid params")
	}
	msg := qc.GetProposalMsg()
	block := &pb.Block{}
	err := proto.Unmarshal(msg, block)
	if err != nil {
		cb.log.Warn("CallVerifyQc ummarshal data failed", "qc", qc)
		return false, err
	}
	header := &pb.Header{
		Logid: "",
	}
	ok, err := cb.cons.CheckMinerMatch(header, block.GetBlock())
	if err != nil {
		cb.log.Warn("CallVerifyQc check miner match failed", "qc", qc)
		return false, err
	}
	return ok, nil
}

// CallProposalMsgWithProposalID call  external consensus for proposalMsg  with the given ProposalID
func (cb *CbftBridge) CallProposalMsgWithProposalID(proposalID []byte) ([]byte, error) {
	blockContent, err := cb.ledger.QueryBlock(proposalID)
	if err != nil {
		cb.log.Warn("CallProposalMsgWithProposalID cannot found block", "block",
			hex.EncodeToString(proposalID))
		return nil, err
	}
	block := &pb.Block{
		Block:   blockContent,
		Blockid: blockContent.GetBlockid(),
		Bcname:  cb.bcname,
	}
	msg, err := proto.Marshal(block)
	if err != nil {
		cb.log.Warn("CallProposalMsgWithProposalID marshal data failed", "block",
			hex.EncodeToString(proposalID))
		return nil, err
	}
	return msg, nil
}

// IsFirstProposal return true if current proposal is the first proposal of bft
// First proposal could have empty or nil PreQC
func (cb *CbftBridge) IsFirstProposal(qc *pb.QuorumCert) (bool, error) {
	if qc == nil {
		return false, fmt.Errorf("invalid params")
	}
	return cb.paceMaker.IsFirstProposal(qc), nil
}
