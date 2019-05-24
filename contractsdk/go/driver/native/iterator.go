package native

import "github.com/xuperchain/xuperunion/contractsdk/go/pb"

type arrayIterator struct {
	pos   int
	items []*pb.IteratorItem
	err   error
}

func newArrayIterator(items []*pb.IteratorItem) *arrayIterator {
	return &arrayIterator{
		pos:   -1,
		items: items,
	}
}

func newErrorArrayIterator(err error) *arrayIterator {
	return &arrayIterator{
		err: err,
	}
}

func (it *arrayIterator) Key() []byte {
	return it.items[it.pos].Key
}

func (it *arrayIterator) Value() []byte {
	return it.items[it.pos].Value
}

func (it *arrayIterator) Error() error {
	return it.err
}

func (it *arrayIterator) Next() bool {
	if it.err != nil {
		return false
	}
	defer func() { it.pos++ }()
	return it.pos+1 < len(it.items)
}

func (it *arrayIterator) Close() {
}
