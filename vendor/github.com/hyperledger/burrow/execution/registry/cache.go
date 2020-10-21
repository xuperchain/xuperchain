// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"sync"

	"github.com/hyperledger/burrow/crypto"
)

// Cache helps prevent unnecessary IAVLTree updates and garbage generation.
type Cache struct {
	sync.RWMutex
	backend  Reader
	registry map[crypto.Address]*nodeInfo
	stats    NodeStats
}

type nodeInfo struct {
	sync.RWMutex
	node    *NodeIdentity
	removed bool
	updated bool
}

var _ Writer = &Cache{}

// NewCache returns a Cache which can write to an output Writer via Sync.
// Not goroutine safe, use syncStateCache if you need concurrent access
func NewCache(backend Reader) *Cache {
	return &Cache{
		backend:  backend,
		registry: make(map[crypto.Address]*nodeInfo),
		stats:    NewNodeStats(),
	}
}

func (cache *Cache) GetNodeByID(id crypto.Address) (*NodeIdentity, error) {
	info, err := cache.get(id)
	if err != nil {
		return nil, err
	}
	info.RLock()
	defer info.RUnlock()
	if info.removed {
		return nil, nil
	}
	return info.node, nil
}

func (cache *Cache) GetNodeIDsByAddress(net string) ([]crypto.Address, error) {
	return cache.stats.GetAddresses(net), nil
}

func (cache *Cache) GetNumPeers() int {
	return len(cache.registry)
}

func (cache *Cache) UpdateNode(id crypto.Address, node *NodeIdentity) error {
	info, err := cache.get(id)
	if err != nil {
		return err
	}
	info.Lock()
	defer info.Unlock()
	if info.removed {
		return fmt.Errorf("UpdateNode on a removed node: %x", id)
	}

	info.node = node
	info.updated = true
	cache.stats.Remove(info.node)
	cache.stats.Insert(node.GetNetworkAddress(), id)
	return nil
}

func (cache *Cache) RemoveNode(id crypto.Address) error {
	info, err := cache.get(id)
	if err != nil {
		return err
	}
	info.Lock()
	defer info.Unlock()
	if info.removed {
		return fmt.Errorf("RemoveNode on removed node: %x", id)
	}
	cache.stats.Remove(info.node)
	info.removed = true
	return nil
}

// Sync writes whatever is in the cache to the output state. Does not flush the cache, to do that call Reset()
// after Sync or use Flush if your wish to use the output state as your next backend
func (cache *Cache) Sync(state Writer) error {
	cache.Lock()
	defer cache.Unlock()

	for id, info := range cache.registry {
		info.RLock()
		if info.removed {
			err := state.RemoveNode(id)
			if err != nil {
				info.RUnlock()
				return err
			}
		} else if info.updated {
			err := state.UpdateNode(id, info.node)
			if err != nil {
				info.RUnlock()
				return err
			}
		}
		info.RUnlock()
	}
	return nil
}

// Reset the cache to empty initialising the backing map to the same size as the previous iteration
func (cache *Cache) Reset(backend Reader) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.registry = make(map[crypto.Address]*nodeInfo)
}

func (cache *Cache) Backend() Reader {
	return cache.backend
}

func (cache *Cache) get(id crypto.Address) (*nodeInfo, error) {
	cache.RLock()
	info := cache.registry[id]
	cache.RUnlock()
	if info == nil {
		cache.Lock()
		defer cache.Unlock()
		info = cache.registry[id]
		if info == nil {
			node, err := cache.backend.GetNodeByID(id)
			if err != nil {
				return nil, err
			}
			info = &nodeInfo{
				node: node,
			}
			cache.registry[id] = info
		}
	}
	return info, nil
}
