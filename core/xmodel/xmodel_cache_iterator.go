// Package xmodel CacheIterator is a merged iterator model cache
package xmodel

import (
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	xmodel_pb "github.com/xuperchain/xuperchain/core/xmodel/pb"
)

type dir int

const (
	// 迭代器被释放
	dirReleased dir = iota - 1
	// start of iterator
	dirSOI
	// end of iterator
	dirEOI
	// 正向迭代
	dirForward
)

type setType string

const (
	setTypeNext  = "Next"
	setTypeFirst = "First"
)

// XMCacheIterator 返回XModelCache的迭代器, 需要对inputsCache、outputsCache和model中的iter进行merge
// 当XMCache可以穿透时需要进行3路merge，当XModelCache不可以穿透时需要进行2路merge
// 当3路迭代时从model中取出的key需要存入inputCache
type XMCacheIterator struct {
	mIter     Iterator            // index: 2
	iters     []iterator.Iterator // index:0 是mcOutIter; index:1 是mcInIter
	cmp       comparer.Comparer
	keys      [][]byte
	markedKey map[string]bool
	index     int
	dir       dir
	err       error
	mc        *XMCache
}

// NewXModelCacheIterator new an instance of XModel Cache iterator
func (mc *XMCache) NewXModelCacheIterator(bucket string, startKey []byte, endKey []byte, cmp comparer.Comparer) (*XMCacheIterator, error) {
	rawStartKey := makeRawKey(bucket, startKey)
	rawEndKey := makeRawKey(bucket, endKey)
	var iters []iterator.Iterator
	mcoi := mc.outputsCache.NewIterator(&util.Range{Start: rawStartKey, Limit: rawEndKey})
	iters = append(iters, mcoi)
	mcii := mc.inputsCache.NewIterator(&util.Range{Start: rawStartKey, Limit: rawEndKey})
	iters = append(iters, mcii)
	var mi Iterator
	if mc.isPenetrate {
		var err error
		mi, err = mc.model.Select(bucket, startKey, endKey)
		if err != nil {
			return nil, err
		}
	}
	return &XMCacheIterator{
		mIter:     mi,
		mc:        mc,
		iters:     iters,
		cmp:       cmp,
		keys:      make([][]byte, 3),
		markedKey: make(map[string]bool),
	}, nil
}

// Data get data pointer to VersionedData for XMCacheIterator
func (mci *XMCacheIterator) Data() *xmodel_pb.VersionedData {
	if mci.err != nil || mci.dir == dirReleased {
		return nil
	}
	switch mci.index {
	case 2:
		return mci.mIter.Data()
	case 0, 1:
		return mci.data(mci.iters[mci.index])
	default:
		return nil
	}
}

func (mci *XMCacheIterator) data(iter iterator.Iterator) *xmodel_pb.VersionedData {
	val := iter.Value()

	data := &xmodel_pb.VersionedData{}

	if err := proto.Unmarshal(val, data); err != nil {
		return nil
	}
	return data
}

// Next get next XMCacheIterator
func (mci *XMCacheIterator) Next() bool {
	if mci.dir == dirEOI || mci.err != nil {
		return false
	} else if mci.dir == dirReleased {
		mci.err = iterator.ErrIterReleased
		return false
	}

	switch mci.dir {
	case dirSOI:
		return mci.First()
	}

	if !mci.setMciKeys(mci.index, setTypeNext) {
		return false
	}
	return mci.next()
}

func (mci *XMCacheIterator) next() bool {
	var key []byte
	if mci.dir == dirForward {
		key = mci.keys[mci.index]
	}
	for x, tkey := range mci.keys {
		if tkey != nil && (key == nil || mci.cmp.Compare(tkey, key) < 0) {
			key = tkey
			mci.index = x
		}
	}
	if key == nil {
		mci.dir = dirEOI
		return false
	}

	if mci.markedKey[string(key)] {
		return mci.Next()
	}
	mci.markedKey[string(key)] = true
	mci.dir = dirForward
	return true
}

// First get the first XMCacheIterator
func (mci *XMCacheIterator) First() bool {
	if mci.err != nil {
		return false
	} else if mci.dir == dirReleased {
		mci.err = iterator.ErrIterReleased
		return false
	}
	if mci.setMciKeys(0, setTypeFirst) && mci.setMciKeys(1, setTypeFirst) && mci.setMciKeys(2, setTypeFirst) {
		mci.dir = dirSOI
		return mci.next()
	}
	return false
}

// Key get key for XMCacheIterator
func (mci *XMCacheIterator) Key() []byte {
	if mci.err != nil || mci.dir == dirReleased {
		return nil
	}
	switch mci.index {
	case 0, 1:
		return mci.iters[mci.index].Key()
	case 2:
		if mci.mc.isPenetrate {
			return mci.mIter.Key()
		}
		return nil
	}
	return nil
}

func (mci *XMCacheIterator) Error() error {
	return mci.err
}

// Release release the XMCacheIterator
func (mci *XMCacheIterator) Release() {
	if mci.dir == dirReleased {
		return
	}
	mci.dir = dirReleased
	if mci.mIter != nil {
		mci.mIter.Release()
	}
	for _, it := range mci.iters {
		it.Release()
	}
	mci.keys = nil
	mci.iters = nil
}

func (mci *XMCacheIterator) setMciKeys(index int, st setType) bool {
	switch index {
	case 0, 1:
		return mci.setMciCiKey(index, st)
	case 2:
		return mci.setMciMiKey(st)
	default:
		return false
	}
}

func (mci *XMCacheIterator) setMciCiKey(index int, st setType) bool {
	mci.keys[index] = nil
	if st == setTypeFirst {
		isFirst := mci.iters[index].First()
		if isFirst {
			for {
				if mci.iters[index].Error() != nil {
					mci.err = mci.iters[index].Error()
					return false
				}
				key := mci.iters[index].Key()
				if mci.mc.isDel(key) {
					if mci.iters[index].Next() {
						continue
					}
					return true
				}
				mci.keys[index] = key
				break
			}
			return true
		}
	} else if st == setTypeNext {
		isNext := mci.iters[index].Next()
		if isNext {
			for {
				if mci.iters[index].Error() != nil {
					mci.err = mci.iters[index].Error()
					return false
				}
				key := mci.iters[index].Key()
				if mci.mc.isDel(key) {
					if mci.iters[index].Next() {
						continue
					}
					return true
				}
				mci.keys[index] = key
				break
			}
			return true
		}
	}
	return true
}

func (mci *XMCacheIterator) setMciMiKey(st setType) bool {
	mci.keys[2] = nil
	if !mci.mc.isPenetrate {
		return true
	}
	if st == setTypeFirst {
		isFirst := mci.mIter.First()
		if isFirst {
			for {
				if mci.mIter.Error() != nil {
					mci.err = mci.mIter.Error()
					return false
				}
				key := mci.mIter.Key()
				if mci.mc.isDel(key) {
					if mci.mIter.Next() {
						continue
					}
					return true
				}
				mci.keys[2] = key
				break
			}
			err := mci.mc.setInputCache(mci.keys[2])
			if err != nil {
				return false
			}
			return true
		}
	} else if st == setTypeNext {
		isNext := mci.mIter.Next()
		if isNext {
			for {
				if mci.mIter.Error() != nil {
					mci.err = mci.mIter.Error()
					return false
				}
				key := mci.mIter.Key()
				if mci.mc.isDel(key) {
					if mci.mIter.Next() {
						continue
					}
					return true
				}
				mci.keys[2] = key
				break
			}
			err := mci.mc.setInputCache(mci.keys[2])
			if err != nil {
				return false
			}
			return true
		}
	}
	return true
}
