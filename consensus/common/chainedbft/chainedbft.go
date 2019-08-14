// Package chainedbft is impliments of hotstuff for consensus common module.
package chainedbft

import (
	"os"

	log "github.com/xuperchain/log15"
	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"
	"github.com/xuperchain/xuperunion/p2pv2"
)

// ChainedBft is the implements of hotstuff
type ChainedBft struct {
	clog log.Logger
	// smr is the Smr instance of hotstuff
	smr    *Smr
	quitCh chan bool
}

// NewChainedBft create and start the chained-bft instance
func NewChainedBft(cfg Config, bcname string, p2p *p2pv2.P2PServerV2, proposalQC, generateQC, lockedQC *chainedbft_pb.QuorumCert) (*ChainedBft, error) {
	// set up log
	xlog := log.New("module", "chainedbft")
	xlog.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))

	// set up smr
	smr, err := NewSmr(cfg, bcname, p2p, proposalQC, generateQC, lockedQC)
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
	for {
		select {
		case <-cb.quitCh:
			cb.clog.Info("Quit chainedbft")
			cb.smr.quitCh <- true
			cb.Stop()
			return nil
		}
	}
}

// Stop will stop ChainedBft instance gracefully
func (cb *ChainedBft) Stop() error {
	return nil
}

// ProcessNewView used to process while view changed. There are three scenarios:
// 1 As the new leader, it will wait for (m-f) replica's new view msg and then create an new Proposers;
// 2 As a normal replica, it will send new view msg to leader;
// 3 As the previous leader, it will send new view msg to new leader with votes of its QuorumCert;
func (cb *ChainedBft) ProcessNewView() error {
	return cb.smr.processNewView()
}

// ProcessPropose used to generate new QuorumCert and broadcast to other replicas
func (cb *ChainedBft) ProcessPropose() (*chainedbft_pb.QuorumCert, error) {
	return cb.smr.processPropose()
}
