package xchain

import (
	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/xuperchain/xuperchain/core/xmodel"
	xmodel_pb "github.com/xuperchain/xuperchain/core/xmodel/pb"
)

type mockStore struct {
	db *leveldb.DB
}

func newMockStore() *mockStore {
	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		panic(err)
	}
	return &mockStore{
		db: db,
	}
}

func makeRawKey(bucket string, key []byte) []byte {
	buf := make([]byte, 0, len(bucket)+1+len(key))
	buf = append(buf, bucket...)
	buf = append(buf, '/')
	return append(buf, key...)
}

func (m *mockStore) Get(bucket string, key []byte) (*xmodel_pb.VersionedData, error) {
	value, err := m.db.Get(makeRawKey(bucket, key), nil)
	if err != nil {
		return nil, err
	}
	data := new(xmodel_pb.VersionedData)
	err = proto.Unmarshal(value, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *mockStore) Select(bucket string, startKey []byte, endKey []byte) (xmodel.Iterator, error) {
	start, end := makeRawKey(bucket, startKey), makeRawKey(bucket, endKey)
	iter := m.db.NewIterator(&util.Range{
		Start: start,
		Limit: end,
	}, nil)
	return newMockIterator(iter), nil
}

func (m *mockStore) Commit(cache *xmodel.XMCache) error {
	txid := make([]byte, 32)
	rand.Read(txid)

	batch := new(leveldb.Batch)
	_, wset, _ := cache.GetRWSets()
	for i, w := range wset {
		rawKey := makeRawKey(w.GetBucket(), w.GetKey())
		value, _ := proto.Marshal(&xmodel_pb.VersionedData{
			RefTxid:   txid,
			RefOffset: int32(i),
			PureData:  w,
		})
		batch.Put(rawKey, value)
	}

	return m.db.Write(batch, nil)
}

func (m *mockStore) NewCache() *xmodel.XMCache {
	cache, err := xmodel.NewXModelCache(m, nil)
	if err != nil {
		panic(err)
	}
	return cache
}

type mockIterator struct {
	iterator.Iterator

	data xmodel_pb.VersionedData
	err  error
}

func newMockIterator(iter iterator.Iterator) xmodel.Iterator {
	return &mockIterator{
		Iterator: iter,
	}
}

func (m *mockIterator) First() bool {
	if m.err != nil {
		return false
	}
	ok := m.Iterator.First()
	if !ok {
		return false
	}
	err := proto.Unmarshal(m.Iterator.Value(), &m.data)
	if err != nil {
		m.err = err
		return false
	}
	return true
}

func (m *mockIterator) Next() bool {
	if m.err != nil {
		return false
	}
	ok := m.Iterator.Next()
	if !ok {
		return false
	}

	err := proto.Unmarshal(m.Iterator.Value(), &m.data)
	if err != nil {
		m.err = err
		return false
	}
	return true
}

func (m *mockIterator) Error() error {
	if m.err != nil {
		return m.err
	}
	return m.Iterator.Error()
}

func (m *mockIterator) Data() *xmodel_pb.VersionedData {
	return &m.data
}
