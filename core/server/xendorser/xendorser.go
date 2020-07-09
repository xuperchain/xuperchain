package xendorser

import (
	context "golang.org/x/net/context"

	"github.com/xuperchain/xuperchain/core/pb"
)

// XEndorser is the interface for endorser service
// Endorser protocol provide standard interface for endorser operations.
// In many cases, a full node could be used for Endorser service.
// For example, an endorser service can provide delegated computing and
// compliance check for transactions, and set an endorser fee for each request.
// endorser service provider could decide how much fee needed for each operation.
type XEndorser interface {
	Init(confPath string, params map[string]interface{}) error
	EndorserCall(ctx context.Context, req *pb.EndorserRequest) (*pb.EndorserResponse, error)
}
