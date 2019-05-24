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
