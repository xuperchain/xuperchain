package relayer

import (
	"fmt"
	//"path/filepath"

	"github.com/golang/protobuf/proto"

	relayerpb "github.com/xuperchain/xuperchain/core/cmd/relayer/pb"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
)

/*
// StorageConfig的一个默认值
const DefaultStorageConfig = StorageConfig{
	storePath: "./",
	fileName: "xuper",
	kvConfig: &kvdb.KVParameter{
		DBPath:					filepath.Join(storePath, fileName),
		KVEngineType:			"defualt",
		MemCacheSize:			128,
		FileHandlersCacheSize:	1024,
		OtherPaths:				[]string{},
	},
}*/

// Storage 存储区块头的句柄
// 跟踪deliverBlockCommand, queryBlockCommand的最新状态
// 存储height --> blockid的映射
// 存储实际的区块头
type Storage struct {
	baseDB           kvdb.Database
	blocksTable      kvdb.Database
	deliverMetaTable kvdb.Database
	queryMetaTable   kvdb.Database
	heightTable      kvdb.Database
}

// StorageConfig Storage实例的传入参数
// 包括底层KVDB的参数, 外层存储路径参数
type StorageConfig struct {
	KVConfig  *kvdb.KVParameter
	StorePath string
	FileName  string
}

// NewStorage 创建一个Storage实例
func NewStorage(storageConfig *StorageConfig) (*Storage, error) {
	baseDB, err := kvdb.NewKVDBInstance(storageConfig.KVConfig)
	if err != nil {
		fmt.Println("failed to open leveldb, err:", err)
		return nil, err
	}
	return &Storage{
		baseDB:           baseDB,
		blocksTable:      kvdb.NewTable(baseDB, "XUPER"),
		deliverMetaTable: kvdb.NewTable(baseDB, "DELIVER_META"),
		queryMetaTable:   kvdb.NewTable(baseDB, "QUERY_META"),
		heightTable:      kvdb.NewTable(baseDB, "HEIGHT_META"),
	}, nil
}

// PutBlockHeader 存储本地区块头
func (storage *Storage) PutBlockHeader(key, value []byte) error {
	return storage.blocksTable.Put(key, value)
}

func (storage *Storage) PutHeightBlockid(height int64, blockid []byte) error {
	sHeight := []byte(fmt.Sprintf("%020d", height))
	return storage.heightTable.Put(sHeight, blockid)
}

// GetBlockHeader 按照blockid查询本地区块头
func (storage *Storage) GetBlockHeader(key []byte) ([]byte, error) {
	return storage.blocksTable.Get(key)
}

// GeBlockHeaderByHeight 按照高度查询本地区块头
func (storage *Storage) GetBlockHeaderByHeight(height int64) ([]byte, error) {
	sHeight := []byte(fmt.Sprintf("%020d", height))
	blockID, kvErr := storage.heightTable.Get(sHeight)
	if kvErr != nil {
		return nil, kvErr
	}
	return storage.GetBlockHeader(blockID)
}

// LoadDeliverMeta 更新DeliverMeta
func (storage *Storage) LoadDeliverMeta() (*relayerpb.DeliverMeta, error) {
	metaBuf, findErr := storage.deliverMetaTable.Get([]byte("DELIVER_META"))
	if findErr == nil {
		meta := &relayerpb.DeliverMeta{}
		err := proto.Unmarshal(metaBuf, meta)
		return meta, err
	}
	return nil, findErr
}

// UpdateDeliverMeta 加载最新持久化的DeliverMeta
func (storage *Storage) UpdateDeliverMeta(meta *relayerpb.DeliverMeta) error {
	metaBuf, pbErr := proto.Marshal(meta)
	if pbErr != nil {
		return pbErr
	}
	return storage.deliverMetaTable.Put([]byte("DELIVER_META"), metaBuf)
}

// LoadQueryMeta 加载最新持久化的QueryMeta
func (storage *Storage) LoadQueryMeta() (*relayerpb.QueryMeta, error) {
	metaBuf, findErr := storage.queryMetaTable.Get([]byte("QUERY_META"))
	if findErr == nil {
		meta := &relayerpb.QueryMeta{}
		err := proto.Unmarshal(metaBuf, meta)
		return meta, err
	}
	return nil, findErr
}

// UpdateQueryMeta 更新QueryMeta
func (storage *Storage) UpdateQueryMeta(meta *relayerpb.QueryMeta) error {
	metaBuf, pbErr := proto.Marshal(meta)
	if pbErr != nil {
		return pbErr
	}
	return storage.queryMetaTable.Put([]byte("QUERY_META"), metaBuf)
}
