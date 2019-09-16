package liveness

import (
	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/pb"
)

// PacemakerInterface is the interface of Pacemaker. It responsible for generating a new round.
// We assume Pacemaker in all correct replicas will have synchronized leadership after GST.
// Safty is entirely decoupled from liveness by any potential instantiation of Packmaker.
// Different consensus have different pacemaker implement
type PacemakerInterface interface {
	// NextNewView sends new view msg to next leader
	// It used while leader changed.
	NextNewView(viewNum int64, proposer, preProposer string) error
	// NextNewProposal generate new proposal directly while the leader haven't changed.
	NextNewProposal(proposalID []byte, data interface{}) error
	// UpdateQCHigh update QuorumCert high of this node.
	//UpdateQCHigh() error
	// CurretQCHigh return current QuorumCert high of this node.
	CurrentQCHigh(proposalID []byte) (*pb.QuorumCert, error)
	// CurrentView return current vie of this node.
	CurrentView() int64
	// UpdateValidatorSet update the validator set of BFT
	UpdateValidatorSet(validators []*cons_base.CandidateInfo) error
}
