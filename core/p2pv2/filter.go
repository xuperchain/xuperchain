package p2pv2

import (
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p-kbucket"
	peer "github.com/libp2p/go-libp2p-peer"
)

// FilterStrategy defines the supported filter strategies
type FilterStrategy string

// supported filter strategies
const (
	DefaultStrategy           FilterStrategy = "DefaultStrategy"
	BucketsStrategy                          = "BucketsStrategy"
	NearestBucketStrategy                    = "NearestBucketStrategy"
	BucketsWithFactorStrategy                = "BucketsWithFactorStrategy"
	CorePeersStrategy                        = "CorePeersStrategy"
)

// PeersFilter the interface for filter peers
type PeersFilter interface {
	Filter() ([]peer.ID, error)
}

// BucketsFilter define filter that get all peers in buckets
type BucketsFilter struct {
	node *Node
}

// Filter 依据Bucket分层广播
func (bf *BucketsFilter) Filter() ([]peer.ID, error) {
	peers := []peer.ID{}
	rt := bf.node.kdht.RoutingTable()
	for i := 0; i < len(rt.Buckets); i++ {
		peers = append(peers, rt.Buckets[i].Peers()...)
	}
	return peers, nil
}

// NearestBucketFilter define filter that get nearest peers from a specified peer ID
type NearestBucketFilter struct {
	node *Node
}

// Filter 广播给最近的Bucket
func (nf *NearestBucketFilter) Filter() ([]peer.ID, error) {
	peers := nf.node.kdht.RoutingTable().NearestPeers(kbucket.ConvertPeerID(nf.node.NodeID()), MaxBroadCastPeers)
	return peers, nil
}

// BucketsFilterWithFactor define filter that get a certain percentage peers in each bucket
type BucketsFilterWithFactor struct {
	node *Node
}

// Filter 从每个Bucket中挑选占比Factor个peers进行广播
// 对于每一个Bucket,平均分成若干块,每个块抽取若干个节点
/*
 *|<---------------- Bucket ---------------->|
 *--------------------------------------------
 *|        |        |        |        |      |
 *--------------------------------------------
 *       split1   split2    split3   split4 split5
 */
func (nf *BucketsFilterWithFactor) Filter() ([]peer.ID, error) {
	factor := 0.5
	rt := nf.node.kdht.RoutingTable()
	filterPeers := []peer.ID{}
	for i := 0; i < len(rt.Buckets); i++ {
		peers := []peer.ID{}
		peers = append(peers, rt.Buckets[i].Peers()...)
		peersSize := len(peers)
		step := int(1.0 / factor)
		splitSize := int(float64(peersSize) / (1.0 / factor))
		if peersSize == 0 {
			continue
		}
		pos := 0
		// 处理split1, split2, split3, split4
		rand.Seed(time.Now().Unix())
		for pos = 0; pos < splitSize; pos++ {
			lastPos := pos * step
			// for each split
			for b := lastPos; b < lastPos+step && b < peersSize; b += step {
				randPos := rand.Intn(step) + lastPos
				filterPeers = append(filterPeers, peers[randPos])
			}
		}
		// 处理split5, 挑选一半出来
		for a := pos * step; a < peersSize; a += 2 {
			filterPeers = append(filterPeers, peers[a])
		}
	}
	return filterPeers, nil
}

// CorePeersFilter define filter for core peers
type CorePeersFilter struct {
	name string
	node *Node
}

// SetRouteName set the core route name to filter
// in XuperChain, the route name is the blockchain name
func (cp *CorePeersFilter) SetRouteName(name string) {
	cp.name = name
}

// Filter select MaxBroadCastCorePeers random peers from core peers,
// half from current and half from next
func (cp *CorePeersFilter) Filter() ([]peer.ID, error) {
	peerids := make([]peer.ID, 0)
	cp.node.routeLock.RLock()
	bcRoute, ok := cp.node.coreRoute[cp.name]
	cp.node.routeLock.RUnlock()
	if !ok {
		return peerids, nil
	}
	currSize := len(bcRoute.CurrentPeers)
	nextSize := len(bcRoute.NextPeers)

	currIdxs := GenerateUniqueRandList(MaxBroadCastCorePeers/2, currSize)
	nextIdxs := GenerateUniqueRandList(MaxBroadCastCorePeers-MaxBroadCastCorePeers/2, nextSize)

	for _, idx := range currIdxs {
		peerids = append(peerids, bcRoute.CurrentPeers[idx].PeerInfo.ID)
	}

	for _, idx := range nextIdxs {
		peerids = append(peerids, bcRoute.NextPeers[idx].PeerInfo.ID)
	}

	return peerids, nil
}

// StaticNodeStrategy a peer filter that contains strategy nodes
type StaticNodeStrategy struct {
	bcname string
	node   *Node
}

// Filter return static nodes peers
func (ss *StaticNodeStrategy) Filter() ([]peer.ID, error) {
	return ss.node.staticNodes[ss.bcname], nil
}

// MultiStrategy a peer filter that contains multiple filters
type MultiStrategy struct {
	node       *Node
	filters    []PeersFilter
	extraPeers []peer.ID
}

// NewMultiStrategy create instance of MultiStrategy
func NewMultiStrategy(node *Node, filters []PeersFilter, extraPeers []peer.ID) *MultiStrategy {
	return &MultiStrategy{
		node:       node,
		filters:    filters,
		extraPeers: extraPeers,
	}
}

// Filter return peer IDs with multiple filters
func (cp *MultiStrategy) Filter() ([]peer.ID, error) {
	res := make([]peer.ID, 0)
	dupCheck := make(map[string]bool)
	// add all filters
	for _, filter := range cp.filters {
		peers, err := filter.Filter()
		if err != nil {
			return res, err
		}
		for _, peer := range peers {
			if _, ok := dupCheck[peer.Pretty()]; !ok {
				dupCheck[peer.Pretty()] = true
				res = append(res, peer)
			}
		}
	}
	// add extra peers
	for _, peer := range cp.extraPeers {
		if _, ok := dupCheck[peer.Pretty()]; !ok {
			dupCheck[peer.Pretty()] = true
			res = append(res, peer)
		}
	}
	return res, nil
}
