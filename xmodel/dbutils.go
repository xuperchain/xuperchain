package xmodel

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
	xmodel_pb "github.com/xuperchain/xuperunion/xmodel/pb"
)

// KVEngineType KV storage type
const KVEngineType = "default"

// BucketSeperator separator between bucket and raw key
const BucketSeperator = "/"

// DelFlag delete flag
const DelFlag = "\x00"

func isDelFlag(value []byte) bool {
	return bytes.Equal([]byte(DelFlag), value)
}

// MakeRawKey make key with bucket and raw key
func MakeRawKey(bucket string, key []byte) []byte {
	return makeRawKey(bucket, key)
}

func makeRawKey(bucket string, key []byte) []byte {
	k := append([]byte(bucket), []byte(BucketSeperator)...)
	return append(k, key...)
}

func parseRawKey(rawKey []byte) (string, []byte, error) {
	idx := bytes.Index(rawKey, []byte(BucketSeperator))
	if idx < 0 {
		return "", nil, fmt.Errorf("parseRawKey failed, invalid raw key:%s", string(rawKey))
	}
	bucket := string(rawKey[:idx])
	key := rawKey[idx+1 : len(rawKey)]
	return bucket, key, nil
}

func queryUnconfirmTx(txid []byte, table kvdb.Database) (*pb.Transaction, error) {
	pbBuf, findErr := table.Get(txid)
	if findErr != nil {
		return nil, findErr
	}
	tx := &pb.Transaction{}
	pbErr := proto.Unmarshal(pbBuf, tx)
	if pbErr != nil {
		return nil, pbErr
	}
	return tx, nil
}

func saveUnconfirmTx(tx *pb.Transaction, batch kvdb.Batch) error {
	buf, err := proto.Marshal(tx)
	if err != nil {
		return err
	}
	rawKey := append([]byte(pb.UnconfirmedTablePrefix), []byte(tx.Txid)...)
	batch.Put(rawKey, buf)
	return nil
}

func openDB(dbPath string, logger log.Logger) (kvdb.Database, error) {
	// new kvdb instance
	kvParam := &kvdb.KVParameter{
		DBPath:                dbPath,
		KVEngineType:          "default",
		MemCacheSize:          128,
		FileHandlersCacheSize: 512,
		OtherPaths:            []string{},
	}
	baseDB, err := kvdb.NewKVDBInstance(kvParam)
	if err != nil {
		logger.Warn("xmodel::openDB failed to open db", "dbPath", dbPath, "err", err)
		return nil, err
	}
	return baseDB, nil
}

// 快速对写集合排序
type pdSlice []xmodel_pb.PureData

// newPdSlice new a slice instance for PureData
func newPdSlice(vpd []*xmodel_pb.PureData) pdSlice {
	pds := []xmodel_pb.PureData{}
	for _, v := range vpd {
		pds = append(pds, *v)
	}
	return pds
}

// Len length of slice of PureData
func (pds pdSlice) Len() int {
	return len(pds)
}

// Swap swap two pureData elements in a slice
func (pds pdSlice) Swap(i, j int) {
	pds[i], pds[j] = pds[j], pds[i]
}

// Less compare two pureData elements with pureData's key in a slice
func (pds pdSlice) Less(i, j int) bool {
	rawKeyI := string(makeRawKey(pds[i].GetBucket(), pds[i].GetKey()))
	rawKeyJ := string(makeRawKey(pds[j].GetBucket(), pds[j].GetKey()))
	if rawKeyI == rawKeyJ {
		// 注: 正常应该无法走到这个逻辑，因为写集合中的key一定是唯一的
		return string(pds[i].GetValue()) < string(pds[j].GetValue())
	}
	return rawKeyI < rawKeyJ
}

func equal(pd, vpd xmodel_pb.PureData) bool {
	rawKeyI := string(makeRawKey(pd.GetBucket(), pd.GetKey()))
	rawKeyJ := string(makeRawKey(vpd.GetBucket(), vpd.GetKey()))
	if rawKeyI != rawKeyJ {
		return false
	}
	valI := string(pd.GetValue())
	valJ := string(vpd.GetValue())
	return valI == valJ
}

// Equal check if two PureData object equal
func Equal(pd, vpd []*xmodel_pb.PureData) bool {
	if len(pd) != len(vpd) {
		return false
	}
	pds := newPdSlice(pd)
	vpds := newPdSlice(vpd)
	sort.Sort(pds)
	sort.Sort(vpds)
	for i, v := range pds {
		if equal(v, vpds[i]) {
			continue
		}
		return false
	}
	return true
}
