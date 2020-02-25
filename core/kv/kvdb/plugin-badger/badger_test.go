package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestBadgerIteratorWithPrefixBasic(t *testing.T) {
	otherOpts := map[string]interface{}{}
	badgerDB := &BadgerDatabase{}
	path, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)
	err = badgerDB.Open(path, otherOpts)
	if err != nil {
		t.Error("open error: ", err)
	}

	// test for Put(key, value byte[])
	err = badgerDB.Put([]byte("key1"), []byte("value1"))
	if err != nil {
		t.Error("badgerDB Put failed, err: ", err)
	}

	badgerDB.Put([]byte("key2"), []byte("value2"))
	badgerDB.Put([]byte("key3"), []byte("value3"))
	badgerDB.Put([]byte("key4"), []byte("value4"))
	badgerDB.Put([]byte("key5"), []byte("value5"))
	badgerDB.Put([]byte("key6"), []byte("value6"))
	badgerDB.Put([]byte("key7"), []byte("value7"))

	// test for Get(key byte[])
	value, vErr := badgerDB.Get([]byte("key1"))
	if vErr != nil {
		t.Error("Get failed, error: ", vErr)
	} else {
		t.Log("key: key1 ", "value: ", string(value))
	}

	// test for NewIteratorWithPrefix
	iter := badgerDB.NewIteratorWithPrefix([]byte("key"))

	iter.Next()
	iter.Next()
	t.Log("iter.Curr() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	// test for Prev()
	preErr := iter.Prev()
	if preErr == false {
		t.Error("iter Prev() error:")
	}
	t.Log("iter.Prev() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	iter.Next()
	t.Log("iter.Next() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	iter.Prev()
	t.Log("iter.Prev() key: ", string(iter.Key()), " value: ", string(iter.Value()))

	lastRet := iter.Last()
	if lastRet == true {
		t.Log("iter.Last() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	}
	iter.First()
	t.Log("iter.First() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	defer badgerDB.Close()
	//defer os.RemoveAll(path)
}

func TestBadgerIteratorWithRangeBasic(t *testing.T) {
	otherOpts := map[string]interface{}{}
	badgerDB := &BadgerDatabase{}
	path, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)
	err = badgerDB.Open(path, otherOpts)
	if err != nil {
		t.Error("open error: ", err)
	}

	// test for Put(key, value byte[])
	err = badgerDB.Put([]byte("key1"), []byte("value1"))
	if err != nil {
		t.Error("badgerDB Put failed, err: ", err)
	}

	badgerDB.Put([]byte("a2"), []byte("value2"))
	badgerDB.Put([]byte("b3"), []byte("value3"))
	badgerDB.Put([]byte("c4"), []byte("value4"))
	badgerDB.Put([]byte("d5"), []byte("value5"))
	badgerDB.Put([]byte("e6"), []byte("value6"))
	badgerDB.Put([]byte("f7"), []byte("value7"))

	// test for Get(key byte[])
	value, vErr := badgerDB.Get([]byte("key1"))
	if vErr != nil {
		t.Error("Get failed, error: ", vErr)
	} else {
		t.Log("key: key1 ", "value: ", string(value))
	}

	iter := badgerDB.NewIteratorWithRange([]byte("a"), []byte("d"))

	iter.Next()
	iter.Next()
	t.Log("iter.Curr() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	// test for Prev()
	preErr := iter.Prev()
	if preErr == false {
		t.Error("iter Prev() error:")
	}
	t.Log("iter.Prev() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	iter.Next()
	t.Log("iter.Next() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	iter.Prev()
	t.Log("iter.Prev() key: ", string(iter.Key()), " value: ", string(iter.Value()))

	lastRet := iter.Last()
	if lastRet == true {
		t.Log("iter.Last() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	}
	iter.First()
	t.Log("iter.First() key: ", string(iter.Key()), " value: ", string(iter.Value()))
	defer badgerDB.Close()
	//defer os.RemoveAll(path)
}
