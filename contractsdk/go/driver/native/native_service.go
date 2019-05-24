package native

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/xuperchain/xuperunion/contractsdk/go/code"
	"github.com/xuperchain/xuperunion/contractsdk/go/pb"
	"google.golang.org/grpc"
)

var (
	_ pb.NativeCodeServer = (*nativeCodeService)(nil)
)

type nativeCodeService struct {
	contract  reflect.Value
	rpcClient pb.SyscallClient
	lastping  time.Time
}

func newNativeCodeService(sockpath string, contract interface{}) *nativeCodeService {
	conn, err := grpc.Dial("unix:"+sockpath, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return &nativeCodeService{
		contract:  reflect.ValueOf(contract),
		rpcClient: pb.NewSyscallClient(conn),
		lastping:  time.Now(),
	}
}

func (s *nativeCodeService) Call(ctx context.Context, request *pb.CallRequest) (*pb.Response, error) {
	nci, err := s.newContext(request.Ctxid, request)
	if err != nil {
		return nil, err
	}
	methodName := request.GetMethod()
	methodv := s.contract.MethodByName(strings.Title(methodName))
	if !methodv.IsValid() {
		return nil, errors.New("bad method " + methodName)
	}
	method, ok := methodv.Interface().(func(code.Context) code.Response)
	if !ok {
		return nil, errors.New("bad method type " + methodName)
	}
	res := method(nci)

	return &pb.Response{
		Status:  int32(res.Status),
		Message: res.Message,
		Body:    res.Body,
	}, nil
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

func (s *nativeCodeService) newContext(ctxid int64, request *pb.CallRequest) (*contextImpl, error) {
	var args map[string]interface{}
	err := json.Unmarshal(request.Args, &args)
	if err != nil {
		return nil, err
	}
	return &contextImpl{
		chainClient: newChainClient(ctxid, s.rpcClient),
		request:     request,
		args:        args,
	}, nil
}
