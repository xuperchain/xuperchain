package bft

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft"
	"github.com/xuperchain/xuperunion/pb"
)

// DPoSPaceMaker the implementation of PaceMakerInterface for TDPoS
type DPoSPaceMaker struct {
	currentView int64
	cbft        *chainedbft.ChainedBft
	log         log.Logger
	cons        base.ConsensusInterface
}

// NewDPoSPaceMaker create new DPoSPaceMaker instance
func NewDPoSPaceMaker(viewNum int64, cbft *chainedbft.ChainedBft, xlog log.Logger, cons base.ConsensusInterface) (*DPoSPaceMaker, error) {
	if cbft == nil {
		return nil, fmt.Errorf("Chained-BFT instance is nil")
	}

	return &DPoSPaceMaker{
		currentView: viewNum,
		cbft:        cbft,
		log:         xlog,
		cons:        cons,
	}, nil
}

// CurrentView get current view number
func (dpm *DPoSPaceMaker) CurrentView() int64 {
	return dpm.currentView
}

// NextNewView is used submit NewView event to bft network
// in most case it means leader changed
func (dpm *DPoSPaceMaker) NextNewView(viewNum int64, proposer, preProposer string) error {
	if viewNum < dpm.currentView {
		return fmt.Errorf("next view cannot smaller than current view number")
	}
	dpm.cbft.ProcessNewView(viewNum, proposer, preProposer)
	return nil
}

// NextNewProposal used to submit new proposal to bft network
// the content is the new block
func (dpm *DPoSPaceMaker) NextNewProposal(proposalID []byte, data interface{}) error {
	block, ok := data.(*pb.Block)
	if !ok {
		return fmt.Errorf("Proposal data is not block")
	}
	blockid := block.GetBlockid()
	blockMsg, err := proto.Marshal(block)
	if err != nil {
		dpm.log.Warn("proposal proto marshal failed", "error", err)
		return err
	}
	_, err = dpm.cbft.GetGenerateQC(proposalID)
	if err != nil {
		dpm.log.Warn("proposal QC generate failed", "error", err)
		return err
	}
	// TODO: add qc into block
	_, err = dpm.cbft.ProcessProposal(dpm.currentView, blockid, blockMsg)
	if err != nil {
		dpm.log.Warn("ProcessProposal failed", "error", err)
		return err
	}
	return nil
}

// UpdateValidatorSet update the validator set of BFT
func (dpm *DPoSPaceMaker) UpdateValidatorSet(validators []*base.CandidateInfo) error {
	return dpm.cbft.UpdateValidateSets(validators)
}
