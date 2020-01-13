package exec

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("key not found")

	// DefaultCacheStore is the default store when WithCacheStore is not set
	// and EnableLazyCompile is true.
	DefaultCacheStore FuncCacheStore = new(defaultCacheStore)
)

// FuncCacheStore used to store compiled function
type FuncCacheStore interface {
	Put(uint64, interface{})
	Get(uint64) (interface{}, bool)
}

type defaultCacheStore struct {
	store sync.Map
}

func (s *defaultCacheStore) Put(key uint64, value interface{}) {
	s.store.Store(key, value)
}

func (s *defaultCacheStore) Get(key uint64) (interface{}, bool) {
	return s.store.Load(key)
}
