package p2pv1

// StaticNodeStrategy a peer filter that contains strategy nodes
type StaticNodeStrategy struct {
	isBroadCast bool
	bcname      string
	pSer        *P2PServerV1
}

// Filter return static nodes peers
func (ss *StaticNodeStrategy) Filter() (interface{}, error) {
	if ss.isBroadCast {
		return ss.pSer.staticNodes["xuper"], nil
	}
	return ss.pSer.staticNodes[ss.bcname], nil
}

// BucketsFilter define filter that get all peers in buckets
type BucketsFilter struct {
}

// Filter 依据Bucket分层广播
func (bf *BucketsFilter) Filter() (interface{}, error) {
	return nil, nil
}

// NearestBucketFilter define filter that get nearest peers from a specified peer ID
type NearestBucketFilter struct {
}

// Filter 广播给最近的Bucket
func (nf *NearestBucketFilter) Filter() (interface{}, error) {
	return nil, nil
}

// BucketsFilterWithFactor define filter that get a certain percentage peers in each bucket
type BucketsFilterWithFactor struct {
}

// Filter BucketsFilterWithFactor广播
func (nf *BucketsFilterWithFactor) Filter() (interface{}, error) {
	return nil, nil
}

// CorePeersFilter define filter for core peers
type CorePeersFilter struct {
	name string
}

// SetRouteName set the core route name to filter
// in XuperChain, the route name is the blockchain name
func (cp *CorePeersFilter) SetRouteName(name string) {
	cp.name = name
}

// Filter select MaxBroadCastCorePeers random peers from core peers,
// half from current and half from next
func (cp *CorePeersFilter) Filter() (interface{}, error) {
	return nil, nil
}
