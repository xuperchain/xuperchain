package p2pv1

import (
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
)

// StaticNodeStrategy a peer filter that contains strategy nodes
type StaticNodeStrategy struct {
	isBroadCast bool
	bcname      string
	pSer        *P2PServerV1
}

// Filter return static nodes peers
func (ss *StaticNodeStrategy) Filter() (interface{}, error) {
	peers := []string{}

	if ss.isBroadCast {
		peers = append(peers, ss.pSer.staticNodes["xuper"]...)
	} else {
		peers = append(peers, ss.pSer.staticNodes[ss.bcname]...)
	}
	if len(ss.pSer.bootNodes) != 0 {
		peers = append(peers, ss.pSer.bootNodes...)
	}
	if len(ss.pSer.dynamicNodes) != 0 {
		peers = append(peers, ss.pSer.dynamicNodes...)
	}
	return peers, nil
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

// MultiStrategy a peer filter that contains multiple filters
type MultiStrategy struct {
	filters     []p2p_base.PeersFilter
	targetPeers []string
}

// NewMultiStrategy create instance of MultiStrategy
func NewMultiStrategy(filters []p2p_base.PeersFilter, targetPeers []string) *MultiStrategy {
	return &MultiStrategy{
		filters:     filters,
		targetPeers: targetPeers,
	}
}

// Filter return peer IDs with multiple filters
func (cp *MultiStrategy) Filter() (interface{}, error) {
	res := make([]string, 0)
	dupCheck := make(map[string]bool)
	// add target peers
	for _, peer := range cp.targetPeers {
		if _, ok := dupCheck[peer]; !ok {
			dupCheck[peer] = true
			res = append(res, peer)
		}
	}
	if len(res) > 0 {
		return res, nil
	}

	// add all filters
	for _, filter := range cp.filters {
		peers, err := filter.Filter()
		if err != nil {
			return res, err
		}
		for _, peer := range peers.([]string) {
			if _, ok := dupCheck[peer]; !ok {
				dupCheck[peer] = true
				res = append(res, peer)
			}
		}
	}
	return res, nil
}
