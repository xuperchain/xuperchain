// Package kvdb 迭代器实现
package kvdb

import (
	"github.com/syndtr/goleveldb/leveldb/util"
)

// QuickIterator define the data structure of quick iterator
type QuickIterator struct {
	iters []Iterator
	cur   int
}

// NewQuickIterator new quick iterator instance
func NewQuickIterator(ldb Database, prefix []byte, middleKey []byte) *QuickIterator {
	qi := &QuickIterator{}
	qi.iters = []Iterator{}
	if middleKey != nil {
		kRange := util.BytesPrefix(prefix)
		start := kRange.Start
		limit := kRange.Limit
		qi.iters = append(qi.iters, ldb.NewIteratorWithRange(middleKey, limit))
		qi.iters = append(qi.iters, ldb.NewIteratorWithRange(start, middleKey))
	} else {
		qi.iters = append(qi.iters, ldb.NewIteratorWithPrefix(prefix))
	}
	qi.cur = 0
	return qi
}

// Next if iterator finished
func (qi *QuickIterator) Next() bool {
	if qi.cur == 0 {
		if !qi.iters[qi.cur].Next() {
			if len(qi.iters) == 2 {
				qi.cur++
			} else {
				return false
			}
		} else {
			return true
		}
	}
	return qi.iters[qi.cur].Next()
}

// Key get key by quick iterator
func (qi *QuickIterator) Key() []byte {
	return qi.iters[qi.cur].Key()
}

// Value get value by quick iterator
func (qi *QuickIterator) Value() []byte {
	return qi.iters[qi.cur].Value()
}

// Release release iterators of quick iterator
func (qi *QuickIterator) Release() {
	for _, it := range qi.iters {
		it.Release()
	}
}

// Error return err for quick iterator
func (qi *QuickIterator) Error() error {
	return qi.iters[qi.cur].Error()
}
