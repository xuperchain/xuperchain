package kvdb

import (
	"strings"
)

type table struct {
	db     Database
	prefix string
}

// NewTable 基于前缀编码方式实现多表
func NewTable(db Database, prefix string) Database {
	return &table{
		db:     db,
		prefix: prefix,
	}
}

func (dt *table) Put(key []byte, value []byte) error {
	return dt.db.Put(append([]byte(dt.prefix), key...), value)
}

func (dt *table) Has(key []byte) (bool, error) {
	return dt.db.Has(append([]byte(dt.prefix), key...))
}

func (dt *table) Get(key []byte) ([]byte, error) {
	return dt.db.Get(append([]byte(dt.prefix), key...))
}

func (dt *table) Delete(key []byte) error {
	return dt.db.Delete(append([]byte(dt.prefix), key...))
}

func (dt *table) Open(path string, options map[string]interface{}) error {
	return nil
}
func (dt *table) Close() {
	// Do nothing; don't close the underlying DB.
}

func (dt *table) NewBatch() Batch {
	panic("not implemented")
}

func (dt *table) NewIteratorWithPrefix(prefix []byte) Iterator {
	return dt.db.NewIteratorWithPrefix(append([]byte(dt.prefix), prefix...))
}

func (dt *table) NewIteratorWithRange(start []byte, limit []byte) Iterator {
	return dt.db.NewIteratorWithRange(append([]byte(dt.prefix), start...), append([]byte(dt.prefix), limit...))
}

// ErrNotFound return true or false when key not found
func ErrNotFound(err error) bool {
	return strings.HasSuffix(err.Error(), "not found")
}
