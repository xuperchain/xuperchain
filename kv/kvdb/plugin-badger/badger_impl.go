//badger wrapper plugin
//so
package main

import (
	"github.com/dgraph-io/badger"
	"github.com/xuperchain/log15"
)

// BadgerDatabase define db backend based on badger
type BadgerDatabase struct {
	fn  string     // filename of db
	db  *badger.DB // db instance
	log log.Logger // logger instance
}

func GetInstance() interface{} {
	return &BadgerDatabase{}
}

// Path returns the path to the database directory
func (bdb *BadgerDatabase) Path() string {
	return bdb.fn
}

func (bdb *BadgerDatabase) Open(path string, options map[string]interface{}) error {
	logger := log.New("database", path)
	bdb.fn = path
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)
	if err != nil {
		log.Warn("badger open failed", "path", path, "err", err)
		return err
	}
	bdb.db = db
	bdb.log = logger
	return nil
}

func (bdb *BadgerDatabase) Close() {
	err := bdb.db.Close()
	if err == nil {
		bdb.log.Info("database closed")
	} else {
		bdb.log.Error("failed to close database", "err", err)
	}
}

func (bdb *BadgerDatabase) Put(key []byte, value []byte) error {
	wb := bdb.db.NewWriteBatch()
	defer wb.Cancel()
	err := wb.Set(key, value, 0)
	if err != nil {
		return err
	}
	return wb.Flush()
}
