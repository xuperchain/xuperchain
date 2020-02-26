package main

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/xuperchain/xuperchain/core/kv/mstorage"
)

func read() {
	store, sErr := mstorage.OpenFile("./data", true, []string{"./disks/disk1/", "./disks/disk2", "./disks/disk3"})
	if sErr != nil {
		panic(sErr)
	}
	db, err := leveldb.Open(store, &opt.Options{ReadOnly: true})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("Key_%08d", i)
		value, gErr := db.Get([]byte(key), nil)
		if gErr != nil {
			panic(gErr)
		}
		if i%10000 == 0 {
			fmt.Println(key, len(value))
		}
	}
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	j := 0
	for iter.Next() {
		if j%10000 == 0 {
			fmt.Println(string(iter.Key()), len(iter.Value()))
		}
		j++
	}
	cErr := db.Close()
	if cErr != nil {
		panic(cErr)
	}
}
