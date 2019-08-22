package liveness

// PacemakerInterface is the interface of Pacemaker. It responsible for generating a new round.
// We assume Pacemaker in all correct replicas will have synchronized leadership after GST.
// Safty is entirely decoupled from liveness by any potential instantiation of Packmaker.
// Different consensus have different pacemaker implement
type PacemakerInterface interface {
	// NextNewView sends new view msg to next leader
	// It used while leader changed.
	NextNewView() error
	// NextNewProposal generate new proposal directly while the leader haven't changed.
	NextNewProposal() error
	// UpdateQCHigh update QuorumCert high of this node.
	UpdateQCHigh() error
	// CurretQCHigh return current QuorumCert high of this node.
	CurretQCHigh() error
	// CurrentView return current vie of this node.
	CurrentView() error
}

// TODO @yucao: DPoS need to implement this
// // Pacemaker the pacemaker struct in
// type Pacemaker struct {
// 	// State of liveness
// 	// view of hotstuff
// 	view int64
// 	// highestQC
// 	highestQC *chainedbft_pb.QuorumCert
// }
