package bft

import (
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

// PoaPaceMaker the implementation of PaceMakerInterface for PoA
type PoaPaceMaker struct {
	bcname      string
	startView   int64
	currentView int64
	address     string
	cbft        *chainedbft.ChainedBft
	log         log.Logger
	ledger      *ledger.Ledger
	cons        base.ConsensusInterface
}

// NewPoaPaceMaker create new PoaPaceMaker instance
func NewPoaPaceMaker(bcname string, startView int64, viewNum int64, address string, cbft *chainedbft.ChainedBft,
	xlog log.Logger, cons base.ConsensusInterface, ledger *ledger.Ledger) (*PoaPaceMaker, error) {
	if cbft == nil {
		return nil, fmt.Errorf("Chained-BFT instance is nil")
	}
	startView++
	return &PoaPaceMaker{
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
func (ppm *PoaPaceMaker) CurrentView() int64 {
	return ppm.currentView
}

// NextNewView is used submit NewView event to bft network
// in most case it means leader changed
func (ppm *PoaPaceMaker) NextNewView(viewNum int64, proposer, preProposer string) error {
	if viewNum < ppm.currentView {
		return fmt.Errorf("next view cannot smaller than current view number")
	}

	ppm.currentView = viewNum
	err := ppm.cbft.ProcessNewView(viewNum, proposer, preProposer)
	ppm.log.Info("bft NewView", "viewNum", viewNum, "proposer", proposer, "preProposer", preProposer)
	return err
}

// NextNewProposal used to submit new proposal to bft network
// the content is the new block
func (ppm *PoaPaceMaker) NextNewProposal(proposalID []byte, data interface{}) error {
	block, ok := data.(*pb.Block)
	if !ok {
		return fmt.Errorf("proposal data is not block")
	}
	if block.GetBlock().GetHeight() < ppm.currentView-1 {
		return fmt.Errorf("proposal height is too small")
	}
	blockid := block.GetBlockid()
	blockMsg, err := proto.Marshal(block)
	if err != nil {
		ppm.log.Warn("proposal proto marshal failed", "error", err)
		return err
	}
	// set current view number to block height
	dpm.currentView = block.GetBlock().GetHeight()
	_, err = dpm.cbft.ProcessProposal(dpm.currentView, blockid, blockMsg, true)

	if err != nil {
		ppm.log.Warn("ProcessProposal failed", "error", err)
		return err
	}
	// set current view number to block height
	ppm.currentView = block.GetBlock().GetHeight()
	ppm.log.Info("bft NewProposal", "viewNum", ppm.currentView, "blockid", hex.EncodeToString(blockid))
	return nil
}

// CurrentQCHigh get the latest QuorumCert
func (ppm *PoaPaceMaker) CurrentQCHigh(proposalID []byte) (*pb.QuorumCert, error) {
	// TODO: what would happen if current QC don't have 2/3 signature?
	return ppm.cbft.GetGenerateQC()
}

// UpdateValidatorSet update the validator set of BFT
func (ppm *PoaPaceMaker) UpdateValidatorSet(validators []*base.CandidateInfo) error {
	valStr, _ := json.Marshal(validators)
	ppm.log.Trace("bft update validator set", "validators", string(valStr))
	return ppm.cbft.UpdateValidateSets(validators)
}

// IsFirstProposal check if current view is the first view
func (ppm *PoaPaceMaker) IsFirstProposal(qc *pb.QuorumCert) bool {
	ppm.log.Trace("IsFirstProposal check", "viewNum", qc.GetViewNumber(), "startView", ppm.startView)
	if qc.GetViewNumber() == ppm.startView {
		return true
	}
	return false
}

// IsLastViewConfirmed check if last block is confirmed
func (ppm *PoaPaceMaker) IsLastViewConfirmed() (bool, error) {
	tipID := ppm.ledger.GetMeta().GetTipBlockid()
	qc, err := ppm.cbft.GetGenerateQC()
	// dpm.log.Debug("IsLastViewConfirmed get generate qc", "qc", qc,
	// 	"proposalID", hex.EncodeToString(qc.GetProposalId()))
	// qc is not valid or qc is valid but it's not the same with last block
	//if err != nil || bytes.Compare(qc.GetProposalId(), tipID) != 0 {
	if err != nil {
		ppm.log.Warn("IsLastViewConfirmed check failed", "error", err, "qc proposerID", qc.GetProposalId(), "tipID", tipID)
		return false, nil
	}
	return true, nil
}

func (ppm *PoaPaceMaker) slaveViewCheck(viewNum int64) bool {
	trunkHeight := ppm.ledger.GetMeta().GetTrunkHeight()
	return viewNum == trunkHeight+1
}

// GetChainedBFT return the chained-bft module
func (ppm *PoaPaceMaker) GetChainedBFT() *chainedbft.ChainedBft {
	return ppm.cbft
}

// Start run BFT
func (ppm *PoaPaceMaker) Start() error {
	go ppm.cbft.Start()
	return nil
}

// Stop finish running BFT
func (ppm *PoaPaceMaker) Stop() error {
	return ppm.cbft.Stop()
}
