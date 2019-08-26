package xmodel

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/xuperchain/xuperunion/pb"
	xmodel_pb "github.com/xuperchain/xuperunion/xmodel/pb"
)

const (
	// DefaultMemDBSize 默认内存db大小
	DefaultMemDBSize = 50000
)

var (
	// ErrHasDel is returned when key was marked as del
	ErrHasDel = errors.New("Key has been mark as del")
	// ErrNotFound is returned when key is not found
	ErrNotFound = errors.New("Key not found")
)

// XMCache data structure for XModel Cache
type XMCache struct {
	// Key: bucket_key; Value: VersionedData
	inputsCache *memdb.DB // bucket -> {k1:v1, k2:v2}
	// Key: bucket_key; Value: PureData
	outputsCache *memdb.DB
	// 是否穿透到model层
	isPenetrate bool
	model       *XModel
}

// NewXModelCache new an instance of XModel Cache
func NewXModelCache(model *XModel, isPenetrate bool) (*XMCache, error) {
	return &XMCache{
		isPenetrate:  isPenetrate,
		model:        model,
		inputsCache:  memdb.New(comparer.DefaultComparer, DefaultMemDBSize),
		outputsCache: memdb.New(comparer.DefaultComparer, DefaultMemDBSize),
	}, nil
}

// Get 读取一个key的值，返回的value就是有版本的data
func (xc *XMCache) Get(bucket string, key []byte) (*xmodel_pb.VersionedData, error) {
	// Level1: get from outputsCache
	data, err := xc.getFromOuputsCache(bucket, key)
	if err != nil && err != memdb.ErrNotFound {
		return nil, err
	}

	if err == nil {
		return data, nil
	}

	// Level2: get and set from inputsCache
	verData, err := xc.getAndSetFromInputsCache(bucket, key)
	if err != nil {
		return nil, err
	}
	if IsEmptyVersionedData(verData) {
		return nil, ErrNotFound
	}
	if isDelFlag(verData.GetPureData().GetValue()) {
		return nil, ErrHasDel
	}
	return verData, nil
}

// Level1 读取，从outputsCache中读取
func (xc *XMCache) getFromOuputsCache(bucket string, key []byte) (*xmodel_pb.VersionedData, error) {
	buKey := makeRawKey(bucket, key)
	val, err := xc.outputsCache.Get(buKey)
	if err != nil {
		return nil, err
	}

	data := &xmodel_pb.VersionedData{}
	if err = proto.Unmarshal(val, data); err != nil {
		return nil, err
	}
	if isDelFlag(data.GetPureData().GetValue()) {
		return nil, ErrHasDel
	}
	return data, nil
}

// Level2 读取，从inputsCache中读取, 读取不到的情况下，如果isPenetrate为true，会更深一层次从model里读取，并且会将内容填充到readSets中
func (xc *XMCache) getAndSetFromInputsCache(bucket string, key []byte) (*xmodel_pb.VersionedData, error) {
	buKey := makeRawKey(bucket, key)
	valBuf, err := xc.inputsCache.Get(buKey)
	if err != nil && err != memdb.ErrNotFound {
		return nil, err
	}

	if err == memdb.ErrNotFound {
		if !xc.isPenetrate {
			return nil, err
		}
		err := xc.setInputCache(buKey)
		if err != nil {
			return nil, err
		}
	}
	valBuf, err = xc.inputsCache.Get(buKey)
	data := &xmodel_pb.VersionedData{}
	if err = proto.Unmarshal(valBuf, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (xc *XMCache) setInputCache(rawKey []byte) error {
	if val, _ := xc.inputsCache.Get(rawKey); val != nil {
		return nil
	}
	bucket, key, err := parseRawKey(rawKey)
	if err != nil {
		return err
	}
	val, err := xc.model.Get(bucket, key)
	if err != nil {
		return err
	}
	valBuf, _ := proto.Marshal(val)
	return xc.inputsCache.Put(rawKey, valBuf)
}

// Put put a pair of <key, value> into XModel Cache
func (xc *XMCache) Put(bucket string, key []byte, value []byte) error {
	buKey := makeRawKey(bucket, key)
	_, err := xc.getFromOuputsCache(bucket, key)
	if err != nil && err != memdb.ErrNotFound && err != ErrHasDel {
		return err
	}

	val := &xmodel_pb.VersionedData{
		PureData: &xmodel_pb.PureData{
			Key:    key,
			Value:  value,
			Bucket: bucket,
		},
	}
	valBuf, err := proto.Marshal(val)
	if err != nil {
		return err
	}
	// put 前先强制get一下
	xc.Get(bucket, key)
	return xc.outputsCache.Put(buKey, valBuf)
}

// Del delete one key from outPutCache, marked its value as `DelFlag`
func (xc *XMCache) Del(bucket string, key []byte) error {
	return xc.Put(bucket, key, []byte(DelFlag))
}

// Select select all kv from a bucket, can set key range, left closed, right opend
// When xc.isPenetrate equals true, three-way merge, When xc.isPenetrate equals false, two-way merge
func (xc *XMCache) Select(bucket string, startKey []byte, endKey []byte) (Iterator, error) {
	return xc.NewXModelCacheIterator(bucket, startKey, endKey, comparer.DefaultComparer)
}

// QueryTx query transaction from xmodel
func (xc *XMCache) QueryTx(txid []byte) (*pb.Transaction, bool, error) {
	return xc.model.QueryTx(txid)
}

// QueryBlock query block from xmodel
func (xc *XMCache) QueryBlock(blockid []byte) (*pb.InternalBlock, error) {
	return xc.model.QueryBlock(blockid)
}

// GetRWSets get read/write sets
func (xc *XMCache) GetRWSets() ([]*xmodel_pb.VersionedData, []*xmodel_pb.PureData, error) {
	readSets, err := xc.getReadSets()
	if err != nil {
		return nil, nil, err
	}
	writeSets, err := xc.getWriteSets()
	if err != nil {
		return nil, nil, err
	}
	return readSets, writeSets, nil
}

func (xc *XMCache) getReadSets() ([]*xmodel_pb.VersionedData, error) {
	var readSets []*xmodel_pb.VersionedData
	iter := xc.inputsCache.NewIterator(&util.Range{Start: nil, Limit: nil})
	defer iter.Release()
	for iter.Next() {
		val := iter.Value()
		vd := &xmodel_pb.VersionedData{}
		err := proto.Unmarshal(val, vd)
		if err != nil {
			return nil, err
		}
		readSets = append(readSets, vd)
	}
	return readSets, nil
}

func (xc *XMCache) getWriteSets() ([]*xmodel_pb.PureData, error) {
	var writeSets []*xmodel_pb.PureData
	iter := xc.outputsCache.NewIterator(&util.Range{Start: nil, Limit: nil})
	defer iter.Release()
	for iter.Next() {
		val := iter.Value()
		vd := &xmodel_pb.VersionedData{}
		err := proto.Unmarshal(val, vd)
		if err != nil {
			return nil, err
		}
		writeSets = append(writeSets, vd.GetPureData())
	}
	return writeSets, nil
}

// isDel 确认key在XModelCache中是否被删除
func (xc *XMCache) isDel(rawKey []byte) bool {
	val, err := xc.outputsCache.Get(rawKey)
	if err == memdb.ErrNotFound {
		return false
	}
	data := &xmodel_pb.VersionedData{}
	err = proto.Unmarshal(val, data)
	if err != nil {
		return false
	}
	return isDelFlag(data.GetPureData().GetValue())
}

// fill 填充XModelCache, 当某个bucket, key已经存在的时候，会覆盖之前的值
func (xc *XMCache) fill(vd *xmodel_pb.VersionedData) error {
	bucket := vd.GetPureData().GetBucket()
	key := vd.GetPureData().GetKey()
	rawKey := makeRawKey(bucket, key)
	valBuf, _ := proto.Marshal(vd)
	return xc.inputsCache.Put(rawKey, valBuf)
}

// GetBcname 返回bcname
func (xc *XMCache) GetBcname() string {
	return xc.model.bcname
}
