package smr

import (
	"sync"

	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	chainedbft_pb "github.com/xuperchain/xuperunion/consensus/common/chainedbft/pb"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
)

// Smr is the state of the node
type Smr struct {
	slog log.Logger
	// config is the config of ChainedBft
	config config.Config
	// bcname of ChainedBft instance
	bcname string
	// the node address
	address string
	// validates sets, changes with external layer consensus
	validates []*cons_base.CandidateInfo

	// p2p is the network instance
	p2p *p2pv2.P2PServerV2
	// p2pMsgChan is the msg channel registered to network
	p2pMsgChan chan *xuper_p2p.XuperMessage

	// Hotstuff State of this nodes
	// votedView is the last voted view, view changes with chain
	votedView int64
	// proposalQC is the proposalBlock's QC
	proposalQC *chainedbft_pb.QuorumCert
	// generateQC is the proposalBlock's QC, refer to generateBlock's votes
	generateQC *chainedbft_pb.QuorumCert
	// lockedQC is the generateBlock's QC, refer to lockedBlock's votes
	lockedQC *chainedbft_pb.QuorumCert
	// votes of QC in mem, key: prposalID, value: *chainedbft_pb.QCSignInfos
	qcVotes *sync.Map
	// new view msg gathered from other replicas, key: viewNumber, value: []*chainedbft_pb.ChainedBftPhaseMessage
	newViewMsgs *sync.Map
	// quitCh stop channel
	QuitCh chan bool
}

// addViewMsgs add
func (s *Smr) addViewMsgs(msg *chainedbft_pb.ChainedBftPhaseMessage) error {
	return nil
}
