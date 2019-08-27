package bft

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft"
	"github.com/xuperchain/xuperunion/ledger"
	"github.com/xuperchain/xuperunion/pb"
)

// DPoSPaceMaker the implementation of PaceMakerInterface for TDPoS
type DPoSPaceMaker struct {
	bcname      string
	startView   int64
	currentView int64
	cbft        *chainedbft.ChainedBft
	log         log.Logger
	ledger      *ledger.Ledger
	cons        base.ConsensusInterface
}

// NewDPoSPaceMaker create new DPoSPaceMaker instance
func NewDPoSPaceMaker(bcname string, startView int64, viewNum int64, cbft *chainedbft.ChainedBft,
	xlog log.Logger, cons base.ConsensusInterface, ledger *ledger.Ledger) (*DPoSPaceMaker, error) {
	if cbft == nil {
		return nil, fmt.Errorf("Chained-BFT instance is nil")
	}

	return &DPoSPaceMaker{
		bcname:      bcname,
		currentView: viewNum,
		startView:   startView,
		cbft:        cbft,
		log:         xlog,
		cons:        cons,
		ledger:      ledger,
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
	dpm.currentView = viewNum
	err := dpm.cbft.ProcessNewView(viewNum, proposer, preProposer)
	dpm.log.Trace("bft NewView", "viewNum", viewNum, "proposer", proposer, "preProposer", preProposer)
	return err
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
	_, err = dpm.cbft.ProcessProposal(dpm.currentView, blockid, blockMsg)
	if err != nil {
		dpm.log.Warn("ProcessProposal failed", "error", err)
		return err
	}
	dpm.log.Trace("bft NewProposal", "viewNum", dpm.currentView, "blockid", hex.EncodeToString(blockid))
	return nil
}

// CurrentQCHigh get the latest QuorumCert
func (dpm *DPoSPaceMaker) CurrentQCHigh(proposalID []byte) (*pb.QuorumCert, error) {
	// TODO: what would happen if current QC don't have 2/3 signature?
	return dpm.cbft.GetGenerateQC(proposalID)
}

// UpdateValidatorSet update the validator set of BFT
func (dpm *DPoSPaceMaker) UpdateValidatorSet(validators []*base.CandidateInfo) error {
	valStr, _ := json.Marshal(validators)
	dpm.log.Debug("bft update validator set", "validators", string(valStr))
	return dpm.cbft.UpdateValidateSets(validators)
}

// IsFirstProposal check if current view is the first view
func (dpm *DPoSPaceMaker) IsFirstProposal(qc *pb.QuorumCert) bool {
	if qc.GetViewNumber() == dpm.startView+1 {
		return true
	}
	return false
}

func (dpm *DPoSPaceMaker) IsLastViewConfirmed() (bool, error) {
	tipID := dpm.ledger.GetMeta().GetTipBlockid()
	qc, err := dpm.cbft.GetGenerateQC([]byte(""))
	// qc is not valid or qc is valid but it's not the same with last block
	if err != nil || bytes.Compare(qc.GetProposalId(), tipID) != 0 {
		dpm.log.Warn("ProcessBeforeMiner IsQuorumCertValidate failed", "error", err)
		tipBlock, err := dpm.ledger.QueryBlock(tipID)
		if err != nil {
			dpm.log.Warn("ProcessBeforeMiner QueryBlock failed", "error", err)
			return false, err
		}
		blockData := &pb.Block{
			Bcname:  dpm.bcname,
			Blockid: tipBlock.Blockid,
			Block:   tipBlock,
		}

		err = dpm.NextNewProposal(tipBlock.Blockid, blockData)
		if err != nil {
			dpm.log.Warn("ProcessBeforeMiner: bft next proposal failed", "error", err)
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// GetChainedBFT return the chained-bft module
func (dpm *DPoSPaceMaker) GetChainedBFT() *chainedbft.ChainedBft {
	return dpm.cbft
}

// Start run BFT
func (dpm *DPoSPaceMaker) Start() error {
	go dpm.cbft.Start()
	return nil
}

// Stop finish running BFT
func (dpm *DPoSPaceMaker) Stop() error {
	return dpm.cbft.Stop()
}
