package main

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger/v2"
	//"github.com/xuperchain/xuperchain/core/kv/kvdb"
)

type BadgerIterator struct {
	badgerDB   *badger.DB
	badgerIter *badger.Iterator
	first      []byte
	last       []byte
	txn        *badger.Txn
	init       bool
	prefixIter bool
	rangeIter  bool
	opts       badger.IteratorOptions
	direction  bool
}

func NewBadgerIterator(db *badger.DB, iterOptions badger.IteratorOptions, prefixIter bool, rangeIter bool, first []byte, last []byte) *BadgerIterator {
	var it *badger.Iterator
	badgerTxn := db.NewTransaction(false)
	it = badgerTxn.NewIterator(iterOptions)
	badgerIterator := &BadgerIterator{
		badgerDB:   db,
		badgerIter: it,
		first:      first,
		last:       last,
		txn:        badgerTxn,
		prefixIter: prefixIter,
		rangeIter:  rangeIter,
		init:       false,
		opts:       iterOptions,
		direction:  false,
	}
	return badgerIterator
}

/*
func NewIterator(first []byte, last []byte) kvdb.Iterator {
    // TODO
    return nil
}
*/

func (iter *BadgerIterator) Key() []byte {
	if iter.badgerIter.Valid() {
		return iter.badgerIter.Item().Key()
	}
	return nil
}

func (iter *BadgerIterator) Value() []byte {
	if iter.badgerIter.Valid() {
		item := iter.badgerIter.Item()
		ival, err := item.ValueCopy(nil)
		if err != nil {
			return nil
		}
		return ival
	}
	fmt.Println("iterator is invalid.......")
	return nil
}

func (iter *BadgerIterator) Next() bool {
	// last time call the function of Prev, here should renew iterator
	if iter.direction == true {
		key := iter.Key()
		iter.badgerIter.Close()
		iter.opts.Reverse = false
		it := iter.txn.NewIterator(iter.opts)
		iter.badgerIter = it
		iter.badgerIter.Seek(key)
	}
	iter.direction = false
	next := iter.next()
	return next
}

func (iter *BadgerIterator) next() bool {
	if !iter.init {
		iter.badgerIter.Seek(iter.first)
		iter.init = true
	} else {
		if !iter.badgerIter.Valid() {
			return false
		}
		iter.badgerIter.Next()
	}
	if iter.prefixIter {
		return iter.badgerIter.Valid() && iter.badgerIter.ValidForPrefix(iter.first)
	}
	if iter.rangeIter {
		valid := iter.badgerIter.Valid()
		if valid == false {
			return valid
		}
		item := iter.badgerIter.Item()
		return bytes.Compare(item.Key(), iter.last) < 0 && bytes.Compare(item.Key(), iter.first) >= 0
	}
	// for general iterator
	return iter.badgerIter.Valid()
}

func (iter *BadgerIterator) Prev() bool {
	// first time to call the function of Prev, renew Iterator
	if iter.direction == false {
		key := iter.Key()
		iter.badgerIter.Close()
		//iter.txn.Discard()
		//badgerTxn := iter.badgerDB.NewTransaction(false)
		iter.opts.Reverse = true
		//it := badgerTxn.NewIterator(iter.opts)
		it := iter.txn.NewIterator(iter.opts)
		iter.badgerIter = it
		//iter.txn = badgerTxn

		iter.badgerIter.Seek(key)
	}
	iter.direction = true
	// not first time to call the function of Prev, call it directly by next
	next := iter.next()

	return next
}

// Last skip to the last iterator
func (iter *BadgerIterator) Last() bool {
	key := iter.Key()
	for iter.Next() {
		key = iter.Key()
	}
	iter.badgerIter.Seek(key)

	valid := iter.badgerIter.Valid()
	if !valid {
		return false
	}
	if iter.rangeIter {
		item := iter.badgerIter.Item()
		return bytes.Compare(item.Key(), iter.last) < 0 && bytes.Compare(item.Key(), iter.first) >= 0
	}
	// 如果根本不存在以iter.first为前缀, 那么应该返回false
	if iter.prefixIter {
		return iter.badgerIter.ValidForPrefix(key)
	}
	return true
}

// First skip to the first iterator
func (iter *BadgerIterator) First() bool {
	key := iter.Key()
	for iter.Prev() {
		key = iter.Key()
	}
	iter.badgerIter.Seek(key)

	return true
}

func (iter *BadgerIterator) Error() error {
	return nil
}

func (iter *BadgerIterator) Release() {
	iter.badgerIter.Close()
	iter.txn.Discard()
}
