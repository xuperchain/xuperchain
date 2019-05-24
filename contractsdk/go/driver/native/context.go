package native

import (
	"github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

type contextImpl struct {
	*chainClient

	request *pb.CallRequest
	args    map[string]interface{}
}

func (n *contextImpl) Args() map[string]interface{} {
	return n.args
}

func (n *contextImpl) TxID() []byte {
	return n.request.Txid
}

func (n *contextImpl) Caller() string {
	return n.request.Caller
}
