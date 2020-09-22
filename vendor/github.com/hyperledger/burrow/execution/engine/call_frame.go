package engine

import (
	"github.com/hyperledger/burrow/acm/acmstate"
	"github.com/hyperledger/burrow/execution/errors"
)

type CallFrame struct {
	// Cache this State wraps
	*acmstate.Cache
	// Where we sync
	backend acmstate.ReaderWriter
	// In order for nested cache to inherit any options
	cacheOptions []acmstate.CacheOption
	// Depth of the call stack
	callStackDepth uint64
	// Max call stack depth
	maxCallStackDepth uint64
}

// Create a new CallFrame to hold state updates at a particular level in the call stack
func NewCallFrame(st acmstate.ReaderWriter, cacheOptions ...acmstate.CacheOption) *CallFrame {
	return newCallFrame(st, 0, 0, cacheOptions...)
}

func newCallFrame(st acmstate.ReaderWriter, stackDepth uint64, maxCallStackDepth uint64, cacheOptions ...acmstate.CacheOption) *CallFrame {
	return &CallFrame{
		Cache:             acmstate.NewCache(st, cacheOptions...),
		backend:           st,
		cacheOptions:      cacheOptions,
		callStackDepth:    stackDepth,
		maxCallStackDepth: maxCallStackDepth,
	}
}

// Put this CallFrame in permanent read-only mode
func (st *CallFrame) ReadOnly() *CallFrame {
	acmstate.ReadOnly(st.Cache)
	return st
}

func (st *CallFrame) WithMaxCallStackDepth(max uint64) *CallFrame {
	st.maxCallStackDepth = max
	return st
}

func (st *CallFrame) NewFrame(cacheOptions ...acmstate.CacheOption) (*CallFrame, error) {
	if st.maxCallStackDepth > 0 && st.maxCallStackDepth == st.callStackDepth {
		return nil, errors.Codes.CallStackOverflow
	}
	return newCallFrame(st.Cache, st.callStackDepth+1, st.maxCallStackDepth,
		append(st.cacheOptions, cacheOptions...)...), nil
}

func (st *CallFrame) Sync() error {
	err := st.Cache.Sync(st.backend)
	if err != nil {
		return errors.AsException(err)
	}
	return nil
}

func (st *CallFrame) CallStackDepth() uint64 {
	return st.callStackDepth
}
