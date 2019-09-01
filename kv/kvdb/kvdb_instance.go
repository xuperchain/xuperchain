package kvdb

import (
	"path/filepath"

	"github.com/xuperchain/xuperunion/common/log"
	"github.com/xuperchain/xuperunion/pluginmgr"
)

type KVParameter struct {
	StorePath             string
	TableName             string
	KvEngineType          string
	MemCacheSize          int
	FileHandlersCacheSize int
	OtherPaths            []string
}

func (param *KVParameter) GetStorePath() string {
	return param.StorePath
}

func (param *KVParameter) GetTableName() string {
	return param.TableName
}

func (param *KVParameter) GetKvEngineType() string {
	return param.KvEngineType
}

func (param *KVParameter) GetMemCacheSize() int {
	return param.MemCacheSize
}

func (param *KVParameter) GetFileHandlersCacheSize() int {
	return param.FileHandlersCacheSize
}

func (param *KVParameter) GetOtherPaths() []string {
	return param.OtherPaths
}

func NewKVDBInstance(param *KVParameter) (Database, error) {
	dbPath := filepath.Join(param.GetStorePath(), param.GetTableName())
	plgMgr, plgErr := pluginmgr.GetPluginMgr()
	if plgErr != nil {
		log.Warn("fail to get plugin manager")
		return nil, plgErr
	}
	var baseDB Database
	soInst, err := plgMgr.PluginMgr.CreatePluginInstance("kv", param.GetKvEngineType())
	if err != nil {
		log.Warn("fail to create plugin instance", "kvtype", param.GetKvEngineType())
		return nil, err
	}
	baseDB = soInst.(Database)
	err = baseDB.Open(dbPath, map[string]interface{}{
		"cache":     param.GetMemCacheSize(),
		"fds":       param.GetFileHandlersCacheSize(),
		"dataPaths": param.GetOtherPaths(),
	})
	if err != nil {
		log.Warn("fail to open leveldb", "dbPath", dbPath, "err", err)
		return nil, err
	}

	return baseDB, nil
}
