package relayer

import (
	"fmt"
	"path/filepath"

	"github.com/xuperchain/xuperchain/core/kv/kvdb"
)

type Storage struct {
	baseDB      kvdb.Database // 默认是leveldb实例
	blocksTable kvdb.Database // 存储区块头
	metaTable   kvdb.Database // 存储最新状态<DeliverBlockCommand>
}

func NewStorage() (*Storage, error) {
	storage := &Storage{}
	storePath := "./"
	// new kvdb instance
	kvParam := &kvdb.KVParameter{
		DBPath:                filepath.Join(storePath, "xuper"),
		KVEngineType:          "default",
		MemCacheSize:          128,
		FileHandlersCacheSize: 1024,
		OtherPaths:            []string{},
	}
	baseDB, err := kvdb.NewKVDBInstance(kvParam)
	if err != nil {
		fmt.Println("failed to open leveldb", "dbPath:", storePath+"xuper", "err:", err)
		return nil, err
	}
	storage.baseDB = baseDB
	storage.blocksTable = kvdb.NewTable(baseDB, "XUPER")
	storage.metaTable = kvdb.NewTable(baseDB, "META")

	return storage, nil
}

func (storage *Storage) Put(key, value []byte) error {
	return storage.blocksTable.Put(key, value)
}

func (storage *Storage) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (storage *Storage) LoadDeliverMeta() ([]byte, error) {
	return storage.metaTable.Get([]byte("META"))
}

func (storage *Storage) UpdateDeliverMeta(metaBuf []byte) error {
	return storage.metaTable.Put([]byte("META"), metaBuf)
}
