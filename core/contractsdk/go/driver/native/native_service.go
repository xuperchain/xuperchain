package native

import (
	"context"
	"net/url"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/exec"
	pb "github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
	pbrpc "github.com/xuperchain/xuperchain/core/contractsdk/go/pbrpc"
	"google.golang.org/grpc"
)

var (
	_ pbrpc.NativeCodeServer = (*nativeCodeService)(nil)
)

type nativeCodeService struct {
	contract  code.Contract
	rpcClient *grpc.ClientConn
	lastping  time.Time
}

func newNativeCodeService(chainAddr string, contract code.Contract) *nativeCodeService {
	uri, err := url.Parse(chainAddr)
	if err != nil {
		panic(err)
	}
	switch uri.Scheme {
	case "tcp":
	default:
		panic("unsupported protocol " + uri.Scheme)
	}
	conn, err := grpc.Dial(uri.Host, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return &nativeCodeService{
		contract:  contract,
		rpcClient: conn,
		lastping:  time.Now(),
	}
}

func (s *nativeCodeService) bridgeCall(method string, request proto.Message, response proto.Message) error {
	// NOTE sync with contract.proto's package name
	fullmethod := "/xchain.contract.svc.Syscall/" + method
	return s.rpcClient.Invoke(context.Background(), fullmethod, request, response)
}

func (s *nativeCodeService) Call(ctx context.Context, request *pb.NativeCallRequest) (*pb.NativeCallResponse, error) {
	exec.RunContract(request.GetCtxid(), s.contract, s.bridgeCall)
	return new(pb.NativeCallResponse), nil
}

func (s *nativeCodeService) Ping(ctx context.Context, request *pb.PingRequest) (*pb.PingResponse, error) {
	s.lastping = time.Now()
	return &pb.PingResponse{}, nil
}

func (s *nativeCodeService) LastpingTime() time.Time {
	return s.lastping
}

func (s *nativeCodeService) Close() error {
	return nil
}
