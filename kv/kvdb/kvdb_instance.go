package kvdb

import (
	"github.com/xuperchain/xuperunion/common/log"
	"github.com/xuperchain/xuperunion/pluginmgr"
)

type KVParameter struct {
	DBPath                string
	KVEngineType          string
	MemCacheSize          int
	FileHandlersCacheSize int
	OtherPaths            []string
}

func (param *KVParameter) GetDBPath() string {
	return param.DBPath
}

func (param *KVParameter) GetKVEngineType() string {
	return param.KVEngineType
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
