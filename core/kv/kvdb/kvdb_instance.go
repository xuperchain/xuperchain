package kvdb

import (
	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/pluginmgr"
)

// KVParameter structure for kv instance parameters
type KVParameter struct {
	DBPath                string
	KVEngineType          string
	MemCacheSize          int
	FileHandlersCacheSize int
	OtherPaths            []string
}

// GetDBPath return the value of DBPath
func (param *KVParameter) GetDBPath() string {
	return param.DBPath
}

// GetKVEngineType return the value of KVEngineType
func (param *KVParameter) GetKVEngineType() string {
	return param.KVEngineType
}

// GetMemCacheSize return the value of MemCacheSize
func (param *KVParameter) GetMemCacheSize() int {
	return param.MemCacheSize
}

// GetFileHandlersCacheSize return the value of FileHandlersCacheSize
func (param *KVParameter) GetFileHandlersCacheSize() int {
	return param.FileHandlersCacheSize
}

// GetOtherPaths return the value of OtherPaths
func (param *KVParameter) GetOtherPaths() []string {
	return param.OtherPaths
}

// NewKVDBInstance instance an object of kvdb
func NewKVDBInstance(param *KVParameter) (Database, error) {
	plgMgr, plgErr := pluginmgr.GetPluginMgr()
	if plgErr != nil {
		log.Warn("fail to get plugin manager")
		return nil, plgErr
	}
	var baseDB Database
	soInst, err := plgMgr.PluginMgr.CreatePluginInstance("kv", param.GetKVEngineType())
	if err != nil {
		log.Warn("fail to create plugin instance", "kvtype", param.GetKVEngineType())
		return nil, err
	}
	baseDB = soInst.(Database)
	err = baseDB.Open(param.GetDBPath(), map[string]interface{}{
		"cache":     param.GetMemCacheSize(),
		"fds":       param.GetFileHandlersCacheSize(),
		"dataPaths": param.GetOtherPaths(),
	})
	if err != nil {
		log.Warn("fail to open leveldb", "dbPath", param.GetDBPath(), "err", err)
		return nil, err
	}

	return baseDB, nil
}
