// +build cloud

// go-leveldb wrapper plugin
// so，package必须是main
// build.sh: go build --buildmode=plugin --tags cloud -o core/plugins/kv/kv-ldb-cloud.so.1.0.0 github.com/xuperchain/xuperchain/core/kv/kvdb/plugin-ldb
package main

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/kv/s3"
	pt "path"
)

// Open opens an instance of LDB with parameters (ldb path and other options)
func (ldb *LDBDatabase) Open(path string, options map[string]interface{}) error {
	setDefaultOptions(options)
	logger := log.New("database", path)
	cache := options["cache"].(int)
	fds := options["fds"].(int)
	logger.Info("Allocated cache and path fds", "cache", cache, "fds", fds)
	cfg := config.NewNodeConfig()
	cfg.LoadConfig()
	//cloud storage
	s3opt := levels3.OpenOption{
		Bucket:        cfg.CloudStorage.Bucket,
		Path:          path,
		Ak:            cfg.CloudStorage.Ak,
		Sk:            cfg.CloudStorage.Sk,
		Region:        cfg.CloudStorage.Region,
		Endpoint:      cfg.CloudStorage.Endpoint,
		LocalCacheDir: pt.Join(cfg.CloudStorage.LocalCacheDir, path),
	}
	st, err := levels3.NewS3Storage(s3opt)
	if err != nil {
		return err
	}
	db, err := leveldb.Open(st, &opt.Options{
		OpenFilesCacheCapacity: fds,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		//db, err = leveldb.Recover(store, nil)
		return err
	}
	// (Re)check for errors and abort if opening of the db failed
	if err != nil {
		return err
	}
	ldb.fn = path
	ldb.db = db
	ldb.log = logger
	return nil
}
