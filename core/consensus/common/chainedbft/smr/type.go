package smr

import (
	"crypto/ecdsa"
	"sync"

	log "github.com/xuperchain/log15"
	cons_base "github.com/xuperchain/xuperchain/core/consensus/base"
	"github.com/xuperchain/xuperchain/core/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperchain/core/consensus/common/chainedbft/external"
	crypto_base "github.com/xuperchain/xuperchain/core/crypto/client/base"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuper_p2p "github.com/xuperchain/xuperchain/core/p2p/pb"
	pb "github.com/xuperchain/xuperchain/core/pb"
)

// Smr is the state of the node
type Smr struct {
	slog log.Logger
	// config is the config of ChainedBft
	config *config.Config
	// bcname of ChainedBft instance
	bcname string
	// the node address
	address   string
	publicKey string
	// private key
	privateKey *ecdsa.PrivateKey
	// last validates sets, changes with external layer consensus
	preValidates []*cons_base.CandidateInfo
	// validates sets, changes with external layer consensus
	validates []*cons_base.CandidateInfo
	// externalCons is the instance that chained bft communicate with
	externalCons external.ExternalInterface
	// cryptoClient is default cryptoclient of chain
	cryptoClient crypto_base.CryptoClient
	// p2p is the network instance
	p2p p2p_base.P2PServer
	// p2pMsgChan is the msg channel registered to network
	p2pMsgChan chan *xuper_p2p.XuperMessage
	// subscribeList is the Subscriber list of the srm instance
	subscribeList []p2p_base.Subscriber

	// Hotstuff State of this nodes
	// votedView is the last voted view, view changes with chain
	votedView int64
	// vscView is the last validated sets changed view number
	vscView int64
	// proposalQC is the proposalBlock's QC
	proposalQC *pb.QuorumCert
	// generateQC is the proposalBlock's QC, refer to generateBlock's votes
	generateQC *pb.QuorumCert
	// lockedQC is the generateBlock's QC, refer to lockedBlock's votes
	lockedQC *pb.QuorumCert
	// localProposal is the proposal local proposaled
	localProposal *sync.Map
	// votes of QC in mem, key: prposalID, value: *pb.QCSignInfos
	qcVoteMsgs *sync.Map
	// new view msg gathered from other replicas, key: viewNumber, value: []*pb.ChainedBftPhaseMessage
	newViewMsgs *sync.Map

	// lk lock
	lk *sync.Mutex
	// quitCh stop channel
	QuitCh chan bool
}
