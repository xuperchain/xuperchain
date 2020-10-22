// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"net"

	"github.com/hyperledger/burrow/crypto"
)

type NodeStats struct {
	Addresses map[string]map[crypto.Address]struct{}
}

func NewNodeStats() NodeStats {
	return NodeStats{
		Addresses: make(map[string]map[crypto.Address]struct{}),
	}
}

func (ns *NodeStats) GetAddresses(net string) []crypto.Address {
	nodes := ns.Addresses[net]
	addrs := make([]crypto.Address, 0, len(nodes))
	for id := range nodes {
		addrs = append(addrs, id)
	}
	return addrs
}

func (ns *NodeStats) Insert(net string, id crypto.Address) {
	_, ok := ns.Addresses[net]
	if !ok {
		ns.Addresses[net] = make(map[crypto.Address]struct{})
	}
	ns.Addresses[net][id] = struct{}{}
}

func (ns *NodeStats) Remove(node *NodeIdentity) bool {
	_, ok := ns.Addresses[node.GetNetworkAddress()]
	if !ok {
		return false
	}
	_, ok = ns.Addresses[node.GetNetworkAddress()][node.TendermintNodeID]
	if ok {
		delete(ns.Addresses[node.GetNetworkAddress()], node.TendermintNodeID)
		return true
	}
	return false
}

type NodeFilter struct {
	state IterableReader
}

func NewNodeFilter(state IterableReader) *NodeFilter {
	return &NodeFilter{
		state: state,
	}
}

func (nf *NodeFilter) QueryPeerByID(id string) bool {
	addr, err := crypto.AddressFromHexString(id)
	if err != nil {
		return false
	}

	node, err := nf.state.GetNodeByID(addr)
	if err != nil || node == nil {
		return false
	}
	return true
}

func (nf *NodeFilter) findByAddress(addr string) bool {
	nodes, err := nf.state.GetNodeIDsByAddress(addr)
	if err != nil {
		panic(err)
	} else if len(nodes) == 0 {
		return false
	}
	return true
}

func (nf *NodeFilter) QueryPeerByAddress(addr string) bool {
	// may have different outbound port in address, so fallback to host
	ok := nf.findByAddress(addr)
	if ok {
		return ok
	}
	host, _, _ := net.SplitHostPort(addr)
	return nf.findByAddress(host)
}

func (nf *NodeFilter) NumPeers() int {
	return nf.state.GetNumPeers()
}

func (rn *NodeIdentity) String() string {
	return fmt.Sprintf("RegisterNode{%v -> %v @ %v}", rn.ValidatorPublicKey, rn.TendermintNodeID, rn.NetworkAddress)
}

type Reader interface {
	GetNodeByID(crypto.Address) (*NodeIdentity, error)
	GetNodeIDsByAddress(net string) ([]crypto.Address, error)
	GetNumPeers() int
}

type Writer interface {
	// Updates the node, creating it if it does not exist
	UpdateNode(crypto.Address, *NodeIdentity) error
	// Remove the node by address
	RemoveNode(crypto.Address) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Iterable interface {
	IterateNodes(consumer func(crypto.Address, *NodeIdentity) error) (err error)
}

type IterableReader interface {
	Iterable
	Reader
}

type IterableReaderWriter interface {
	Iterable
	ReaderWriter
}
