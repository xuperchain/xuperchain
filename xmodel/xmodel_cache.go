package xmodel

import (
	"bytes"
	"errors"
	"math/big"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/xuperchain/xuperunion/pb"
	xmodel_pb "github.com/xuperchain/xuperunion/xmodel/pb"
)

const (
	// DefaultMemDBSize 默认内存db大小
	DefaultMemDBSize = 32
)

var (
	// ErrHasDel is returned when key was marked as del
	ErrHasDel = errors.New("Key has been mark as del")
	// ErrNotFound is returned when key is not found
	ErrNotFound = errors.New("Key not found")
)

var (
	contractUtxoInputKey  = []byte("ContractUtxo.Inputs")
	contractUtxoOutputKey = []byte("ContractUtxo.Outputs")
)

// UtxoVM manages utxos
type UtxoVM interface {
	SelectUtxos(fromAddr string, fromPubKey string, totalNeed *big.Int, needLock, excludeUnconfirmed bool) ([]*pb.TxInput, [][]byte, *big.Int, error)
}

// XMCache data structure for XModel Cache
type XMCache struct {
	// Key: bucket_key; Value: VersionedData
	inputsCache *memdb.DB // bucket -> {k1:v1, k2:v2}
	// Key: bucket_key; Value: PureData
	outputsCache *memdb.DB
	// 是否穿透到model层
	isPenetrate bool
	model       *XModel
	utxoCache   *UtxoCache
}

// NewXModelCache new an instance of XModel Cache
func NewXModelCache(model *XModel, utxovm UtxoVM) (*XMCache, error) {
	return &XMCache{
		isPenetrate:  true,
		model:        model,
		inputsCache:  memdb.New(comparer.DefaultComparer, DefaultMemDBSize),
		outputsCache: memdb.New(comparer.DefaultComparer, DefaultMemDBSize),
		utxoCache:    NewUtxoCache(utxovm),
	}, nil
}

func NewXModelCacheWithInputs(vdatas []*xmodel_pb.VersionedData, utxoInputs []*pb.TxInput) *XMCache {
	xc := &XMCache{
		isPenetrate:  false,
		inputsCache:  memdb.New(comparer.DefaultComparer, DefaultMemDBSize),
		outputsCache: memdb.New(comparer.DefaultComparer, DefaultMemDBSize),
	}
	for _, vd := range vdatas {
		bucket := vd.GetPureData().GetBucket()
		key := vd.GetPureData().GetKey()
		rawKey := makeRawKey(bucket, key)
		valBuf, _ := proto.Marshal(vd)
		xc.inputsCache.Put(rawKey, valBuf)
	}
	xc.utxoCache = NewUtxoCacheWithInputs(utxoInputs)
	return xc
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
	if bucket != TransientBucket {
		// put 前先强制get一下
		xc.Get(bucket, key)
	}
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

// Transfer transfer tokens using utxo
func (xc *XMCache) Transfer(from, to string, amount *big.Int) error {
	return xc.utxoCache.Transfer(from, to, amount)
}

// GetUtxoRWSets returns the inputs and outputs of utxo
func (xc *XMCache) GetUtxoRWSets() ([]*pb.TxInput, []*pb.TxOutput) {
	return xc.utxoCache.GetRWSets()
}

func (xc *XMCache) PutUtxos(inputs []*pb.TxInput, outputs []*pb.TxOutput) error {
	var in, out []byte
	var err error
	if len(inputs) != 0 {
		in, err = marshalMessages(inputs)
		if err != nil {
			return err
		}
	}
	if len(outputs) != 0 {
		out, err = marshalMessages(outputs)
		if err != nil {
			return err
		}
	}
	if in != nil {
		err = xc.Put(TransientBucket, contractUtxoInputKey, in)
		if err != nil {
			return err
		}
	}
	if out != nil {
		err = xc.Put(TransientBucket, contractUtxoOutputKey, out)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseContractUtxo parse contract utxo inputs from tx write sets
func ParseContractUtxoInputs(tx *pb.Transaction) ([]*pb.TxInput, error) {
	var (
		utxoInputs []*pb.TxInput
		extInput   []byte
	)
	for _, out := range tx.GetTxOutputsExt() {
		if out.GetBucket() != TransientBucket {
			continue
		}
		if bytes.Equal(out.GetKey(), contractUtxoInputKey) {
			extInput = out.GetValue()
		}
	}
	if extInput != nil {
		err := unmsarshalMessages(extInput, &utxoInputs)
		if err != nil {
			return nil, err
		}
	}
	return utxoInputs, nil
}

// ParseContractUtxo parse contract utxos from tx write sets
func ParseContractUtxo(tx *pb.Transaction) ([]*pb.TxInput, []*pb.TxOutput, error) {
	var (
		utxoInputs  []*pb.TxInput
		utxoOutputs []*pb.TxOutput
		extInput    []byte
		extOutput   []byte
	)
	for _, out := range tx.GetTxOutputsExt() {
		if out.GetBucket() != TransientBucket {
			continue
		}
		if bytes.Equal(out.GetKey(), contractUtxoInputKey) {
			extInput = out.GetValue()
		}
		if bytes.Equal(out.GetKey(), contractUtxoOutputKey) {
			extOutput = out.GetValue()
		}
	}
	if extInput != nil {
		err := unmsarshalMessages(extInput, &utxoInputs)
		if err != nil {
			return nil, nil, err
		}
	}
	if extOutput != nil {
		err := unmsarshalMessages(extOutput, &utxoOutputs)
		if err != nil {
			return nil, nil, err
		}
	}
	return utxoInputs, utxoOutputs, nil
}

func makeInputsMap(txInputs []*pb.TxInput) map[string]bool {
	res := map[string]bool{}
	if len(txInputs) == 0 {
		return nil
	}
	for _, v := range txInputs {
		utxoKey := string(v.GetRefTxid()) + strconv.FormatInt(int64(v.GetRefOffset()), 10)
		res[utxoKey] = true
	}
	return res
}

func isSubOutputs(contractOutputs, txOutputs []*pb.TxOutput) bool {
	markedOutput := map[string]int{}
	for _, v := range txOutputs {
		key := string(v.GetAmount()) + string(v.GetToAddr())
		markedOutput[key]++
	}

	for _, v := range contractOutputs {
		key := string(v.GetAmount()) + string(v.GetToAddr())
		if val, ok := markedOutput[key]; !ok {
			return false
		} else if val < 1 {
			return false
		} else {
			markedOutput[key] = val - 1
		}
	}
	return true
}

// IsContractUtxoEffective check if contract utxo in tx utxo
func IsContractUtxoEffective(contractTxInputs []*pb.TxInput, contractTxOutputs []*pb.TxOutput, tx *pb.Transaction) bool {
	if len(contractTxInputs) > len(tx.GetTxInputs()) || len(contractTxOutputs) > len(tx.GetTxOutputs()) {
		return false
	}

	contractTxInputsMap := makeInputsMap(contractTxInputs)
	txInputsMap := makeInputsMap(tx.GetTxInputs())
	for k := range contractTxInputsMap {
		if !(txInputsMap[k]) {
			return false
		}
	}

	if !isSubOutputs(contractTxOutputs, tx.GetTxOutputs()) {
		return false
	}
	return true
}
