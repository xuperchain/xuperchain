package code

import (
	"math/big"

	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

// Context is the context in which the contract runs
type Context interface {
	Args() map[string][]byte
	Caller() string
	Initiator() string
	AuthRequire() []string

	PutObject(key []byte, value []byte) error
	GetObject(key []byte) ([]byte, error)
	DeleteObject(key []byte) error
	NewIterator(start, limit []byte) Iterator

	QueryTx(txid string) (*pb.TxStatus, error)
	QueryBlock(blockid string) (*pb.Block, error)
	Transfer(to string, amount *big.Int) error
	Call(module, contract, method string, args map[string][]byte) (*Response, error)
}

// Iterator iterates over key/value pairs in key order
type Iterator interface {
	Key() []byte
	Value() []byte
	Next() bool
	Error() error
	// Iterator 必须在使用完毕后关闭
	Close()
}

// PrefixRange returns key range that satisfy the given prefix
func PrefixRange(prefix []byte) ([]byte, []byte) {
	var limit []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			limit = make([]byte, i+1)
			copy(limit, prefix)
			limit[i] = c + 1
			break
		}
	}
	return prefix, limit
}
