package bft

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	log "github.com/xuperchain/log15"

	"github.com/golang/protobuf/proto"

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
	address     string
	cbft        *chainedbft.ChainedBft
	log         log.Logger
	ledger      *ledger.Ledger
	cons        base.ConsensusInterface
}

// NewDPoSPaceMaker create new DPoSPaceMaker instance
func NewDPoSPaceMaker(bcname string, startView int64, viewNum int64, address string, cbft *chainedbft.ChainedBft,
	xlog log.Logger, cons base.ConsensusInterface, ledger *ledger.Ledger) (*DPoSPaceMaker, error) {
	if cbft == nil {
		return nil, fmt.Errorf("Chained-BFT instance is nil")
	}
	startView++
	return &DPoSPaceMaker{
		bcname:      bcname,
		currentView: viewNum,
		startView:   startView,
		address:     address,
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
	dpm.log.Info("bft NewView", "viewNum", viewNum, "proposer", proposer, "preProposer", preProposer)
	return err
}

// NextNewProposal used to submit new proposal to bft network
// the content is the new block
func (dpm *DPoSPaceMaker) NextNewProposal(proposalID []byte, data interface{}) error {
	block, ok := data.(*pb.Block)
	if !ok {
		return fmt.Errorf("Proposal data is not block")
	}
	if block.GetBlock().GetHeight() < dpm.currentView-1 {
		return fmt.Errorf("Proposal height is too small")
	}
	blockid := block.GetBlockid()
	blockMsg, err := proto.Marshal(block)
	if err != nil {
		dpm.log.Warn("proposal proto marshal failed", "error", err)
		return err
	}
	// set current view number to block height
	dpm.currentView = block.GetBlock().GetHeight()
	_, err = dpm.cbft.ProcessProposal(dpm.currentView, blockid, blockMsg)
	if err != nil {
		dpm.log.Warn("ProcessProposal failed", "error", err)
		return err
	}
	// set current view number to block height
	dpm.currentView = block.GetBlock().GetHeight()
	dpm.log.Info("bft NewProposal", "viewNum", dpm.currentView, "blockid", hex.EncodeToString(blockid))
	return nil
}

// CurrentQCHigh get the latest QuorumCert
func (dpm *DPoSPaceMaker) CurrentQCHigh(proposalID []byte) (*pb.QuorumCert, error) {
	// TODO: what would happen if current QC don't have 2/3 signature?
	return dpm.cbft.GetGenerateQC()
}

// UpdateValidatorSet update the validator set of BFT
func (dpm *DPoSPaceMaker) UpdateValidatorSet(validators []*base.CandidateInfo) error {
	valStr, _ := json.Marshal(validators)
	dpm.log.Trace("bft update validator set", "validators", string(valStr))
	return dpm.cbft.UpdateValidateSets(validators)
}

// IsFirstProposal check if current view is the first view
func (dpm *DPoSPaceMaker) IsFirstProposal(qc *pb.QuorumCert) bool {
	dpm.log.Trace("IsFirstProposal check", "viewNum", qc.GetViewNumber(), "startView", dpm.startView)
	if qc.GetViewNumber() == dpm.startView {
		return true
	}
	return false
}

// IsLastViewConfirmed check if last block is confirmed
func (dpm *DPoSPaceMaker) IsLastViewConfirmed() (bool, error) {
	tipID := dpm.ledger.GetMeta().GetTipBlockid()
	qc, err := dpm.cbft.GetGenerateQC()
	// dpm.log.Debug("IsLastViewConfirmed get generate qc", "qc", qc,
	// 	"proposalID", hex.EncodeToString(qc.GetProposalId()))
	// qc is not valid or qc is valid but it's not the same with last block
	if err != nil || bytes.Compare(qc.GetProposalId(), tipID) != 0 {
		dpm.log.Warn("IsLastViewConfirmed check failed", "error", err)
		return false, nil
	}
	return true, nil
}

func (dpm *DPoSPaceMaker) slaveViewCheck(viewNum int64) bool {
	trunkHeight := dpm.ledger.GetMeta().GetTrunkHeight()
	return viewNum == trunkHeight+1
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
