package event

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/pb"
)

var _ Topic = (*BlockTopic)(nil)

// BlockTopic handles block events
type BlockTopic struct {
	chainmg ChainManager
}

// NewBlockTopic instances BlockTopic from ChainManager
func NewBlockTopic(chainmg ChainManager) *BlockTopic {
	return &BlockTopic{
		chainmg: chainmg,
	}
}

// NewFilterIterator make a new Iterator base on filter
func (b *BlockTopic) NewFilterIterator(pbfilter *pb.BlockFilter) (Iterator, error) {
	filter, err := newBlockFilter(pbfilter)
	if err != nil {
		return nil, err
	}
	return b.newIterator(filter)
}

// ParseFilter 从指定的bytes buffer反序列化topic过滤器
// 返回的参数会作为入参传递给NewIterator的filter参数
func (b *BlockTopic) ParseFilter(buf []byte) (interface{}, error) {
	pbfilter := new(pb.BlockFilter)
	err := proto.Unmarshal(buf, pbfilter)
	if err != nil {
		return nil, err
	}

	return pbfilter, nil
}

// MarshalEvent encode event payload returns from Iterator.Data()
func (b *BlockTopic) MarshalEvent(x interface{}) ([]byte, error) {
	msg := x.(proto.Message)
	return proto.Marshal(msg)
}

// NewIterator make a new Iterator base on filter
func (b *BlockTopic) NewIterator(ifilter interface{}) (Iterator, error) {
	pbfilter, ok := ifilter.(*pb.BlockFilter)
	if !ok {
		return nil, errors.New("bad filter type for block event")
	}
	filter, err := newBlockFilter(pbfilter)
	if err != nil {
		return nil, err
	}
	return b.newIterator(filter)
}

func (b *BlockTopic) newIterator(filter *blockFilter) (Iterator, error) {

	blockStore, err := b.chainmg.GetBlockStore(filter.GetBcname())
	if err != nil {
		return nil, err
	}

	var startBlockNum, endBlockNum int64
	if filter.GetRange().GetStart() == "" {
		n, err := blockStore.TipBlockHeight()
		if err != nil {
			return nil, err
		}
		startBlockNum = n
	} else {
		n, err := strconv.ParseInt(filter.GetRange().GetStart(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error %s when parse start block number", err)
		}
		startBlockNum = n
	}

	if filter.GetRange().GetEnd() == "" {
		endBlockNum = -1
	} else {
		n, err := strconv.ParseInt(filter.GetRange().GetEnd(), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error %s when parse end block number", err)
		}
		endBlockNum = n
	}

	biter := NewBlockIterator(blockStore, startBlockNum, endBlockNum)
	return &filteredBlockIterator{
		biter:  biter,
		filter: filter,
	}, nil
}
