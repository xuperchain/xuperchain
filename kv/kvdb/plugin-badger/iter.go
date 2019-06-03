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
}

func NewBadgerIterator(db *badger.DB, iterOptions badger.IteratorOptions, first []byte, last []byte) *BadgerIterator {
	var it *badger.Iterator
	db.View(func(txn *badger.Txn) error {
		it = txn.NewIterator(iterOptions)
		defer it.Close()
		return nil
	})

	badgerIterator := &BadgerIterator{
		badgerDB:   db,
		badgerIter: it,
		first:      first,
		last:       last,
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
	if iter.badgerIter.Valid() {
		iter.badgerIter.Next()
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
	// TODO
}
