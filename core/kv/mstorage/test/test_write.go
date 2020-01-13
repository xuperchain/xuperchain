package main

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/xuperchain/xuperunion/kv/mstorage"
	"strings"
)

func main() {
	store, sErr := mstorage.OpenFile("./data", false, []string{"./disks/disk1/", "./disks/disk2", "./disks/disk3"})
	if sErr != nil {
		panic(sErr)
	}
	db, err := leveldb.Open(store, nil)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100000; i++ {
		key := fmt.Sprintf("Key_%08d", i)
		value := strings.Repeat("x", 1024)
		pErr := db.Put([]byte(key), []byte(value), nil)
		if pErr != nil {
			panic(pErr)
		}
		if i%10000 == 0 {
			fmt.Println(i)
		}
	}
	cErr := db.Close()
	if cErr != nil {
		panic(cErr)
	}
}
