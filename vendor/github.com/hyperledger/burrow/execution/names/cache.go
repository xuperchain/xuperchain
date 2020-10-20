// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package names

import (
	"fmt"
	"sort"
	"sync"
)

// The Cache helps prevent unnecessary IAVLTree updates and garbage generation.
type Cache struct {
	sync.RWMutex
	backend Reader
	names   map[string]*nameInfo
}

type nameInfo struct {
	sync.RWMutex
	entry   *Entry
	removed bool
	updated bool
}

var _ Writer = &Cache{}

// Returns a Cache that wraps an underlying NameRegCacheGetter to use on a cache miss, can write to an
// output Writer via Sync. Not goroutine safe, use syncStateCache if you need concurrent access
func NewCache(backend Reader) *Cache {
	return &Cache{
		backend: backend,
		names:   make(map[string]*nameInfo),
	}
}

func (cache *Cache) GetName(name string) (*Entry, error) {
	nameInfo, err := cache.get(name)
	if err != nil {
		return nil, err
	}
	nameInfo.RLock()
	defer nameInfo.RUnlock()
	if nameInfo.removed {
		return nil, nil
	}
	return nameInfo.entry, nil
}

func (cache *Cache) UpdateName(entry *Entry) error {
	nameInfo, err := cache.get(entry.Name)
	if err != nil {
		return err
	}
	nameInfo.Lock()
	defer nameInfo.Unlock()
	if nameInfo.removed {
		return fmt.Errorf("UpdateName on a removed name: %s", nameInfo.entry.Name)
	}

	nameInfo.entry = entry
	nameInfo.updated = true
	return nil
}

func (cache *Cache) RemoveName(name string) error {
	nameInfo, err := cache.get(name)
	if err != nil {
		return err
	}
	nameInfo.Lock()
	defer nameInfo.Unlock()
	if nameInfo.removed {
		return fmt.Errorf("RemoveName on removed name: %s", name)
	}
	nameInfo.removed = true
	return nil
}

// Writes whatever is in the cache to the output Writer state. Does not flush the cache, to do that call Reset()
// after Sync or use Flush if your wish to use the output state as your next backend
func (cache *Cache) Sync(state Writer) error {
	cache.Lock()
	defer cache.Unlock()
	// Determine order for names
	// note names may be of any length less than some limit
	var names sort.StringSlice
	for nameStr := range cache.names {
		names = append(names, nameStr)
	}
	sort.Stable(names)

	// Update or delete names
	for _, name := range names {
		nameInfo := cache.names[name]
		nameInfo.RLock()
		if nameInfo.removed {
			err := state.RemoveName(name)
			if err != nil {
				nameInfo.RUnlock()
				return err
			}
		} else if nameInfo.updated {
			err := state.UpdateName(nameInfo.entry)
			if err != nil {
				nameInfo.RUnlock()
				return err
			}
		}
		nameInfo.RUnlock()
	}
	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration
func (cache *Cache) Reset(backend Reader) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.names = make(map[string]*nameInfo)
}

func (cache *Cache) Backend() Reader {
	return cache.backend
}

// Get the cache accountInfo item creating it if necessary
func (cache *Cache) get(name string) (*nameInfo, error) {
	cache.RLock()
	nmeInfo := cache.names[name]
	cache.RUnlock()
	if nmeInfo == nil {
		cache.Lock()
		defer cache.Unlock()
		nmeInfo = cache.names[name]
		if nmeInfo == nil {
			entry, err := cache.backend.GetName(name)
			if err != nil {
				return nil, err
			}
			nmeInfo = &nameInfo{
				entry: entry,
			}
			cache.names[name] = nmeInfo
		}
	}
	return nmeInfo, nil
}
