package exec

import (
	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

var (
	_ code.Iterator = (*KVIterator)(nil)
)

const MAX_ITERATOR_CAP = 100

type KVIterator struct {
	buf          []*pb.IteratorItem // current buffer of the kv items
	curBuf       *pb.IteratorItem   // pointer of current position
	curIdx       int                // next index
	c            *contractContext   // where we can get the kv items
	err          error
	start, limit []byte
}

// newKVIterator return a code.Iterator
func newKVIterator(c *contractContext, start, limit []byte) code.Iterator {
	return &KVIterator{
		start: start,
		limit: limit,
		c:     c,
	}
}

// load loads the data from xbrigde, called when buf is empty, maintains the curIdx and starter
func (ki *KVIterator) load() {
	//clean the buf at beginning
	ki.buf = ki.buf[0:0]
	req := &pb.IteratorRequest{
		Start:  ki.start,
		Limit:  ki.limit,
		Cap:    MAX_ITERATOR_CAP,
		Header: &ki.c.header,
	}
	resp := new(pb.IteratorResponse)
	if err := ki.c.bridgeCallFunc("NewIterator", req, resp); err != nil {
		ki.err = err
		return
	}
	if len(resp.Items) == 0 {
		ki.start = ki.limit
		return
	}
	ki.curIdx = 0
	ki.buf = resp.Items
	ki.start = resp.Items[len(resp.Items)-1].Key
}

func (ki *KVIterator) Key() []byte {
	return ki.curBuf.Key
}

func (ki *KVIterator) Value() []byte {
	return ki.curBuf.Value
}

func (ki *KVIterator) Next() bool {
	//永远保证有数据
	if ki.curIdx == len(ki.buf) {
		ki.load()
	}
	if len(ki.buf) == 0 || ki.err != nil {
		return false
	}
	ki.curBuf = ki.buf[ki.curIdx]
	ki.curIdx += 1
	return true
}
func (ki *KVIterator) Error() error {
	return ki.err
}

func (ki *KVIterator) Close() {}
