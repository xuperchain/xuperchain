package memrpc

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/bridge"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
	"github.com/xuperchain/xuperunion/test/util"
)

type serverContext struct {
	*util.XModelContext
	tm     *bridge.ContextManager
	server *Server
}

func (s *serverContext) NewContext() int64 {
	ctx := s.tm.MakeContext()
	ctx.Cache = s.Cache
	ctx.ResourceLimits = contract.MaxLimits
	return ctx.ID
}

func withServerContext(t *testing.T, callback func(s *serverContext)) {
	util.WithXModelContext(t, func(x *util.XModelContext) {
		tm := bridge.NewContextManager()
		syscall := bridge.NewSyscallService(tm, nil)
		s := &serverContext{
			tm:            tm,
			server:        NewServer(syscall),
			XModelContext: x,
		}
		callback(s)
	})

}

func TestMethodNotFound(t *testing.T) {
	withServerContext(t, func(s *serverContext) {
		_, err := s.server.CallMethod(context.TODO(), 0, "missing_method", nil)
		if err == nil {
			t.Fatalf("expect none nil error")
		}
	})
}

func TestCallMethod(t *testing.T) {
	withServerContext(t, func(s *serverContext) {
		ctxid := s.NewContext()
		{
			request := &pb.PutRequest{
				Header: &pb.SyscallHeader{
					Ctxid: ctxid,
				},
				Key:   []byte("k"),
				Value: []byte("v"),
			}
			requestBuf, _ := proto.Marshal(request)
			responseBuf, err := s.server.CallMethod(context.TODO(), ctxid, "PutObject", requestBuf)
			if err != nil {
				t.Fatal(err)
			}
			response := new(pb.PutResponse)
			err = proto.Unmarshal(responseBuf, response)
			if err != nil {
				t.Fatal(err)
			}
		}
		{
			request := &pb.GetRequest{
				Header: &pb.SyscallHeader{
					Ctxid: ctxid,
				},
				Key: []byte("k"),
			}
			requestBuf, _ := proto.Marshal(request)
			responseBuf, err := s.server.CallMethod(context.TODO(), ctxid, "GetObject", requestBuf)
			if err != nil {
				t.Fatal(err)
			}
			response := new(pb.GetResponse)
			err = proto.Unmarshal(responseBuf, response)
			if err != nil {
				t.Fatal(err)
			}
			value := string(response.GetValue())
			if value != "v" {
				t.Fatalf("expect `v` got `%s`", value)
			}
		}
	})
}
