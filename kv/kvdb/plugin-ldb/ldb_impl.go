// go-leveldb wrapper plugin
// so 模式，package必须是main
package main

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/kv/kvdb"
)

// LDBDatabase define data structure of storage
type LDBDatabase struct {
	fn  string      // filename of db
	db  *leveldb.DB // LevelDB instance
	log log.Logger  // logger instance
}

// GetInstance get instance of LDBDatabase
func GetInstance() interface{} {
	return &LDBDatabase{}
}

func setDefaultOptions(options map[string]interface{}) {
	if options["cache"] == nil {
		options["cache"] = 16
	}
	if options["fds"] == nil {
		options["fds"] = 16
	}
	if options["dataPaths"] == nil {
		options["dataPaths"] = []string{}
	}
}

// Path returns the path to the database directory.
func (db *LDBDatabase) Path() string {
	return db.fn
}

// Put puts the given key / value to the queue
func (db *LDBDatabase) Put(key []byte, value []byte) error {
	// Generate the data to write to disk, update the meter and write
	//value = rle.Compress(value)

	return db.db.Put(key, value, nil)
}

// Has if the given key exists
func (db *LDBDatabase) Has(key []byte) (bool, error) {
	return db.db.Has(key, nil)
}

// Get returns the given key if it's present.
func (db *LDBDatabase) Get(key []byte) ([]byte, error) {
	// Retrieve the key and increment the miss counter if not found
	dat, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

// Delete deletes the key from the queue and database
func (db *LDBDatabase) Delete(key []byte) error {
	// Execute the actual operation
	return db.db.Delete(key, nil)
}

// NewIterator returns an instance of Iterator
func (db *LDBDatabase) NewIterator() kvdb.Iterator {
	return db.db.NewIterator(nil, nil)
}

// NewIteratorWithRange returns an instance of Iterator with range
func (db *LDBDatabase) NewIteratorWithRange(start []byte, limit []byte) kvdb.Iterator {
	keyRange := &util.Range{Start: start, Limit: limit}
	return db.db.NewIterator(keyRange, nil)
}

// NewIteratorWithPrefix returns an instance of Iterator with prefix
func (db *LDBDatabase) NewIteratorWithPrefix(prefix []byte) kvdb.Iterator {
	return db.db.NewIterator(util.BytesPrefix(prefix), nil)
}

// Close close database instance
func (db *LDBDatabase) Close() {
	err := db.db.Close()
	if err == nil {
		db.log.Info("Database closed")
	} else {
		db.log.Error("Failed to close database", "err", err)
	}
}

// LDB returns ldb instance
func (db *LDBDatabase) LDB() *leveldb.DB {
	return db.db
}

// NewBatch returns batch instance of ldb
func (db *LDBDatabase) NewBatch() kvdb.Batch {
	return &ldbBatch{db: db.db, b: new(leveldb.Batch), keys: map[string]bool{}}
}

type ldbBatch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
	keys map[string]bool
}

func (b *ldbBatch) Put(key, value []byte) error {
	b.b.Put(key, value)
	b.size += len(value)
	return nil
}

func (b *ldbBatch) Delete(key []byte) error {
	b.b.Delete(key)
	b.size += len(key)
	return nil
}

func (b *ldbBatch) PutIfAbsent(key, value []byte) error {
	if !b.keys[string(key)] {
		b.b.Put(key, value)
		b.size += len(value)
		b.keys[string(key)] = true
		return nil
	}
	return fmt.Errorf("duplicated key in batch, (HEX) %x", key)
}

func (b *ldbBatch) Exist(key []byte) bool {
	return b.keys[string(key)]
}

func (b *ldbBatch) Write() error {
	return b.db.Write(b.b, nil)
}

func (b *ldbBatch) ValueSize() int {
	return b.size
}

func (b *ldbBatch) Reset() {
	b.b.Reset()
	b.size = 0
	b.keys = map[string]bool{}
}
