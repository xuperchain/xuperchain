/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 *
 *
 * lru_cache.go - an implementation of LRU cache
 *
 * DESCRIPTION
 * Usage:
 * import (
 *     "github.com/xuperchain/xuperchain/core/common"
 * )
 * // Create lru cache
 * cache := common.NewLRUCache(10000)
 * // Add a key-value pair to cache
 * cache.Add("key1", "val1")
 * // Get value by key
 * value, ok := cache.Get("Key1")
 * // Delete value by key
 * cache.Del("key1")
 * // Get count of items in cache
 * count := cache.Len()
 */

package common

import (
	"container/list"
	"fmt"
	"sync"
)

// LRUCache cache struct
type LRUCache struct {
	lock     sync.Mutex
	capacity int                           // maximum number of key-value pairs
	cache    map[interface{}]*list.Element // map for cached key-value pairs
	lru      *list.List                    // LRU list
}

// Pair <key, value>
type Pair struct {
	key   interface{} // cache key
	value interface{} // cache value
}

// NewLRUCache New function
func NewLRUCache(capacity int) *LRUCache {
	c := new(LRUCache)
	c.capacity = capacity
	c.cache = make(map[interface{}]*list.Element)
	c.lru = list.New()
	return c
}

// Get get cached value from LRU cache
// Params:
//     - key: cache key
// Return:
//     - value: cache value
//     - ok   : true if found, false if not
func (c *LRUCache) Get(key interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if elem, ok := c.cache[key]; ok {
		c.lru.MoveToFront(elem) // move node to head of lru list
		return elem.Value.(*Pair).value, true
	}
	return nil, false
}

// Add add a key-value pair to LRU cache
// Params:
//     - key  : cache key
//     - value: cache value
// Return:
//     - evictOrNot: true if eviction occurs, false if not
func (c *LRUCache) Add(key interface{}, value interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	// update item if found in cache
	if elem, ok := c.cache[key]; ok {
		c.lru.MoveToFront(elem) // update lru list
		elem.Value.(*Pair).value = value
		return false
	}
	// add item if not found
	elem := c.lru.PushFront(&Pair{key, value})
	c.cache[key] = elem
	// evict item if needed
	if c.lru.Len() > c.capacity {
		c.evict()
		return true
	}
	return false
}

// evict a key-value pair from LRU cache
func (c *LRUCache) evict() {
	elem := c.lru.Back()
	if elem == nil {
		return
	}
	// remove item at the end of lru list
	c.lru.Remove(elem)
	delete(c.cache, elem.Value.(*Pair).key)
}

// Del delete cached value from cache
// Params:
//     - key: cache key
func (c *LRUCache) Del(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if elem, ok := c.cache[key]; ok {
		c.lru.Remove(elem)
		delete(c.cache, key)
	}
}

// Len get number of items in cache
func (c *LRUCache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.lru.Len()
}

// Keys get keys of items in cache
func (c *LRUCache) Keys() []interface{} {
	var keyList []interface{}
	c.lock.Lock()
	for key := range c.cache {
		keyList = append(keyList, key)
	}
	c.lock.Unlock()
	return keyList
}

// EnlargeCapacity enlarge the capacity of cache
func (c *LRUCache) EnlargeCapacity(newCapacity int) error {
	// lock
	c.lock.Lock()
	defer c.lock.Unlock()
	// check newCapacity
	if newCapacity < c.capacity {
		return fmt.Errorf("newCapacity[%d] must be larger than current[%d]",
			newCapacity, c.capacity)
	}
	c.capacity = newCapacity
	return nil
}
