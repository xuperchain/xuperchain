// KV database interface
// 通用的kv库接口，背后实现可以是leveldb、rocksdb或者是分布式kv

package kvdb

// Iterator NewIteratorXX操作后得到的迭代器
type Iterator interface {
	Key() []byte
	Value() []byte
	Next() bool
	Prev() bool
	Last() bool
	First() bool
	Error() error
	Release()
}

// Database KV数据库的接口
type Database interface {
	Open(path string, options map[string]interface{}) error
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) (bool, error)
	Delete(key []byte) error
	Close()
	NewBatch() Batch
	NewIteratorWithRange(start []byte, limit []byte) Iterator
	NewIteratorWithPrefix(prefix []byte) Iterator
}

// Batch Batch操作的接口
type Batch interface {
	ValueSize() int
	Write() error
	Reset()
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	PutIfAbsent(key []byte, value []byte) error
	Exist(key []byte) bool
}
