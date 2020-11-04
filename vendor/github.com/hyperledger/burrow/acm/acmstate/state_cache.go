// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package acmstate

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/errors"
)

type Cache struct {
	sync.RWMutex
	name     string
	backend  Reader
	accounts map[crypto.Address]*accountInfo
	readonly bool
}

type accountInfo struct {
	sync.RWMutex
	account *acm.Account
	storage map[binary.Word256][]byte
	removed bool
	updated bool
}

type CacheOption func(*Cache) *Cache

// Returns a Cache that wraps an underlying Reader to use on a cache miss, can write to an output Writer
// via Sync. Goroutine safe for concurrent access.
func NewCache(backend Reader, options ...CacheOption) *Cache {
	cache := &Cache{
		backend:  backend,
		accounts: make(map[crypto.Address]*accountInfo),
	}
	for _, option := range options {
		option(cache)
	}
	return cache
}

func Named(name string) CacheOption {
	return func(cache *Cache) *Cache {
		cache.name = name
		return cache
	}
}

var ReadOnly CacheOption = func(cache *Cache) *Cache {
	cache.readonly = true
	return cache
}

func (cache *Cache) GetAccount(address crypto.Address) (*acm.Account, error) {
	accInfo, err := cache.get(address)
	if err != nil {
		return nil, err
	}
	accInfo.RLock()
	defer accInfo.RUnlock()
	if accInfo.removed {
		return nil, nil
	}
	return accInfo.account.Copy(), nil
}

func (cache *Cache) UpdateAccount(account *acm.Account) error {
	if account == nil {
		return errors.Errorf(errors.Codes.IllegalWrite, "UpdateAccount called with nil account")
	}
	if cache.readonly {
		return errors.Errorf(errors.Codes.IllegalWrite,
			"UpdateAccount called in a read-only context on account %v", account.GetAddress())
	}
	accInfo, err := cache.get(account.GetAddress())
	if err != nil {
		return err
	}
	accInfo.Lock()
	defer accInfo.Unlock()
	if accInfo.removed {
		return errors.Errorf(errors.Codes.IllegalWrite, "UpdateAccount on a removed account: %s", account.GetAddress())
	}
	accInfo.account = account.Copy()
	accInfo.updated = true
	return nil
}

func (cache *Cache) RemoveAccount(address crypto.Address) error {
	if cache.readonly {
		return errors.Errorf(errors.Codes.IllegalWrite, "RemoveAccount called on read-only account %v", address)
	}
	accInfo, err := cache.get(address)
	if err != nil {
		return err
	}
	accInfo.Lock()
	defer accInfo.Unlock()
	if accInfo.removed {
		return fmt.Errorf("RemoveAccount on a removed account: %s", address)
	}
	accInfo.removed = true
	return nil
}

func (cache *Cache) Transfer(from, to crypto.Address, amount *big.Int) error {
	return nil
}

// Iterates over all cached accounts first in cache and then in backend until consumer returns true for 'stop'
func (cache *Cache) IterateCachedAccount(consumer func(*acm.Account) (stop bool)) (stopped bool, err error) {
	// Try cache first for early exit
	cache.RLock()
	for _, info := range cache.accounts {
		if consumer(info.account) {
			cache.RUnlock()
			return true, nil
		}
	}
	cache.RUnlock()
	return false, nil
}

func (cache *Cache) GetStorage(address crypto.Address, key binary.Word256) ([]byte, error) {
	accInfo, err := cache.get(address)
	if err != nil {
		return []byte{}, err
	}
	// Check cache
	accInfo.RLock()
	value, ok := accInfo.storage[key]
	accInfo.RUnlock()
	if !ok {
		accInfo.Lock()
		defer accInfo.Unlock()
		value, ok = accInfo.storage[key]
		if !ok {
			// Load from backend
			value, err = cache.backend.GetStorage(address, key)
			if err != nil {
				return []byte{}, err
			}
			accInfo.storage[key] = value
		}
	}
	return value, nil
}

// NOTE: Set value to zero to remove.
func (cache *Cache) SetStorage(address crypto.Address, key binary.Word256, value []byte) error {
	if cache.readonly {
		return errors.Errorf(errors.Codes.IllegalWrite,
			"SetStorage called in a read-only context on account %v", address)
	}
	accInfo, err := cache.get(address)
	if accInfo.account == nil {
		return errors.Errorf(errors.Codes.IllegalWrite,
			"SetStorage called on an account that does not exist: %v", address)
	}
	accInfo.Lock()
	defer accInfo.Unlock()
	if err != nil {
		return err
	}
	if accInfo.removed {
		return errors.Errorf(errors.Codes.IllegalWrite, "SetStorage on a removed account: %s", address)
	}
	accInfo.storage[key] = value
	accInfo.updated = true
	return nil
}

// Iterates over all cached storage items first in cache and then in backend until consumer returns true for 'stop'
func (cache *Cache) IterateCachedStorage(address crypto.Address,
	consumer func(key binary.Word256, value []byte) error) error {
	accInfo, err := cache.get(address)
	if err != nil {
		return err
	}
	accInfo.RLock()
	// Try cache first for early exit
	for key, value := range accInfo.storage {
		if err := consumer(key, value); err != nil {
			accInfo.RUnlock()
			return err
		}
	}
	accInfo.RUnlock()
	return err
}

// Syncs changes to the backend in deterministic order. Sends storage updates before updating
// the account they belong so that storage values can be taken account of in the update.
func (cache *Cache) Sync(st Writer) error {
	if cache.readonly {
		// Sync is (should be) a no-op for read-only - any modifications should have been caught in respective methods
		return nil
	}
	cache.Lock()
	defer cache.Unlock()
	var addresses crypto.Addresses
	for address := range cache.accounts {
		addresses = append(addresses, address)
	}

	sort.Sort(addresses)
	for _, address := range addresses {
		accInfo := cache.accounts[address]
		accInfo.RLock()
		if accInfo.removed {
			err := st.RemoveAccount(address)
			if err != nil {
				return err
			}
		} else if accInfo.updated {
			// First update account in case it needs to be created
			err := st.UpdateAccount(accInfo.account)
			if err != nil {
				return err
			}
			// Sort keys
			var keys binary.Words256
			for key := range accInfo.storage {
				keys = append(keys, key)
			}
			sort.Sort(keys)
			// Update account's storage
			for _, key := range keys {
				value := accInfo.storage[key]
				err := st.SetStorage(address, key, value)
				if err != nil {
					return err
				}
			}

		}
		accInfo.RUnlock()
	}

	return nil
}

// Resets the cache to empty initialising the backing map to the same size as the previous iteration.
func (cache *Cache) Reset(backend Reader) {
	cache.Lock()
	defer cache.Unlock()
	cache.backend = backend
	cache.accounts = make(map[crypto.Address]*accountInfo, len(cache.accounts))
}

func (cache *Cache) String() string {
	if cache.name == "" {
		return fmt.Sprintf("StateCache{Length: %v}", len(cache.accounts))
	}
	return fmt.Sprintf("StateCache{Name: %v; Length: %v}", cache.name, len(cache.accounts))
}

// Get the cache accountInfo item creating it if necessary
func (cache *Cache) get(address crypto.Address) (*accountInfo, error) {
	cache.RLock()
	accInfo := cache.accounts[address]
	cache.RUnlock()
	if accInfo == nil {
		cache.Lock()
		defer cache.Unlock()
		accInfo = cache.accounts[address]
		if accInfo == nil {
			account, err := cache.backend.GetAccount(address)
			if err != nil {
				return nil, err
			}
			accInfo = &accountInfo{
				account: account,
				storage: make(map[binary.Word256][]byte),
			}
			cache.accounts[address] = accInfo
		}
	}
	return accInfo, nil
}
