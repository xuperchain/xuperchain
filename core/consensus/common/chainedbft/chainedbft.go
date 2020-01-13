// Package chainedbft is impliments of hotstuff for consensus common module.
package chainedbft

import (
	"crypto/ecdsa"

	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/external"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/smr"
	crypto_base "github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/p2pv2"
	"github.com/xuperchain/xuperunion/pb"
)

// ChainedBft is the implements of hotstuff
type ChainedBft struct {
	clog log.Logger
	// smr is the Smr instance of hotstuff
	smr    *smr.Smr
	quitCh chan bool
}

// NewChainedBft create and start the chained-bft instance
func NewChainedBft(
	xlog log.Logger,
	cfg *config.Config,
	bcname string,
	address string,
	publicKey string,
	privateKey *ecdsa.PrivateKey,
	validates []*cons_base.CandidateInfo,
	externalCons external.ExternalInterface,
	cryptoClient crypto_base.CryptoClient,
	p2p p2pv2.P2PServer,
	proposalQC, generateQC, lockedQC *pb.QuorumCert) (*ChainedBft, error) {

	// set up smr
	smr, err := smr.NewSmr(xlog, cfg, bcname, address, publicKey, privateKey,
		validates, externalCons, cryptoClient, p2p, proposalQC, generateQC, lockedQC)
	if err != nil {
		xlog.Error("NewChainedBft instance error")
		return nil, err
	}
	chainedBft := &ChainedBft{
		clog:   xlog,
		smr:    smr,
		quitCh: make(chan bool, 1),
	}
	go chainedBft.Start()
	return chainedBft, nil
}

// Start will start ChainedBft instance, smr instance instance
func (cb *ChainedBft) Start() error {
	go cb.smr.Start()
	for {
		select {
		case <-cb.quitCh:
			cb.clog.Info("Quit chainedbft")
			cb.smr.QuitCh <- true
			cb.Stop()
			return nil
		}
	}
}

// Stop will stop ChainedBft instance gracefully
func (cb *ChainedBft) Stop() error {
	cb.quitCh <- true
	return nil
}

// ProcessNewView used to process while view changed. There are three scenarios:
// 1 As the new leader, it will wait for (m-f) replica's new view msg and then create an new Proposers;
// 2 As a normal replica, it will send new view msg to leader;
// 3 As the previous leader, it will send new view msg to new leader with votes of its QuorumCert;
func (cb *ChainedBft) ProcessNewView(viewNumber int64, leader, preLeader string) error {
	return cb.smr.ProcessNewView(viewNumber, leader, preLeader)
}

// GetGenerateQC get latest proposal QC
func (cb *ChainedBft) GetGenerateQC() (*pb.QuorumCert, error) {
	return cb.smr.GetGenerateQC()
}

// ProcessProposal used to generate new QuorumCert and broadcast to other replicas
func (cb *ChainedBft) ProcessProposal(viewNumber int64, proposalID, proposalMsg []byte) (*pb.QuorumCert, error) {
	return cb.smr.ProcessProposal(viewNumber, proposalID, proposalMsg)
}

// UpdateValidateSets will update the validates while
func (cb *ChainedBft) UpdateValidateSets(validates []*cons_base.CandidateInfo) error {
	return cb.smr.UpdateValidateSets(validates)
}

// IsQuorumCertValidate return whether QC is validated
func (cb *ChainedBft) IsQuorumCertValidate(qc *pb.QuorumCert) (bool, error) {
	return cb.smr.IsQuorumCertValidate(qc)
}
