package xmodel

import (
	xmodel_pb "github.com/xuperchain/xuperchain/core/xmodel/pb"
)

// Iterator iterator interface
type Iterator interface {
	Data() *xmodel_pb.VersionedData
	Next() bool
	First() bool
	Error() error
	Key() []byte
	Release()
}

// XMReader xmodel interface for reader
type XMReader interface {
	//读取一个key的值，返回的value就是有版本的data
	Get(bucket string, key []byte) (*xmodel_pb.VersionedData, error)
	//扫描一个bucket中所有的kv, 调用者可以设置key区间[startKey, endKey)
	Select(bucket string, startKey []byte, endKey []byte) (Iterator, error)
}
