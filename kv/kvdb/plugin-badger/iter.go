package main

import (
	"fmt"
	"github.com/dgraph-io/badger"
	//"github.com/xuperchain/xuperunion/kv/kvdb"
)

type BadgerIterator struct {
	badgerDB   *badger.DB
	badgerIter *badger.Iterator
	first      []byte
	last       []byte
	txn        *badger.Txn
	init       bool
	prefixIter bool
}

func NewBadgerIterator(db *badger.DB, iterOptions badger.IteratorOptions, first []byte, last []byte) *BadgerIterator {
	var it *badger.Iterator
	badgerTxn := db.NewTransaction(false)
	it = badgerTxn.NewIterator(iterOptions)
	badgerIterator := &BadgerIterator{
		badgerDB:   db,
		badgerIter: it,
		first:      first,
		last:       last,
		txn:        badgerTxn,
		prefixIter: true,
		init:       false,
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
	if !iter.init {
		iter.badgerIter.Seek(iter.first)
		iter.init = true
	} else {
		iter.badgerIter.Next()
	}
	if iter.prefixIter {
		return iter.badgerIter.ValidForPrefix(iter.first)
	}
	return iter.badgerIter.Valid()
}

func (iter *BadgerIterator) Prev() bool {
	return true
}

func (iter *BadgerIterator) Last() bool {
	return true
}

func (iter *BadgerIterator) First() bool {
	return true
}

func (iter *BadgerIterator) Error() error {
	return nil
}

func (iter *BadgerIterator) Release() {
	iter.badgerIter.Close()
	iter.txn.Discard()
}
