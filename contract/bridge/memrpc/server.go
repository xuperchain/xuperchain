package memrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperunion/contract/bridge"
	pb "github.com/xuperchain/xuperunion/contractsdk/go/pb"
)

var (
	// ErrMethodNotFound is returned when the method can not be found
	ErrMethodNotFound = errors.New("syscall method not found")
)

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	messageType = reflect.TypeOf((*proto.Message)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

// Server represents memory RPC server
type Server struct {
	methods  map[string]*reflect.Method
	syscall  *bridge.SyscallService
	vsyscall reflect.Value
}

func isContextType(tp reflect.Type) bool {
	return tp == contextType
}

func isMessageType(tp reflect.Type) bool {
	return tp.Implements(messageType)
}

func isErrorType(tp reflect.Type) bool {
	return tp == errorType
}

func parseMethods(syscall *bridge.SyscallService) map[string]*reflect.Method {
	methods := make(map[string]*reflect.Method)
	v := reflect.TypeOf(syscall)
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Method(i)
		tp := method.Type
		// Method(*bridge.SyscallService, context.Context, Request) (Response, error)
		if tp.NumIn() != 3 || tp.NumOut() != 2 {
			continue
		}
		if !isContextType(tp.In(1)) || !isMessageType(tp.In(2)) {
			continue
		}
		if !isMessageType(tp.Out(0)) || !isErrorType(tp.Out(1)) {
			continue
		}
		methods[method.Name] = &method
	}
	return methods
}

// NewServer instances a new Server
func NewServer(syscall *bridge.SyscallService) *Server {
	return &Server{
		methods:  parseMethods(syscall),
		syscall:  syscall,
		vsyscall: reflect.ValueOf(syscall),
	}
}

type syscallHeaderGetter interface {
	GetHeader() *pb.SyscallHeader
}

// CallMethod runs a single rpc call. requestBuf expected to be a protobuf message
func (s *Server) CallMethod(ctx context.Context, ctxid int64, method string, requestBuf []byte) ([]byte, error) {
	m, ok := s.methods[method]
	if !ok {
		return nil, ErrMethodNotFound
	}
	// m.Type.In(2) 为指针类型，取完Elem()之后就编程非指针类型，再经过New就变成原来的指针类型
	request := reflect.New(m.Type.In(2).Elem())
	reqmsg := request.Interface().(proto.Message)
	err := proto.Unmarshal(requestBuf, reqmsg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal request error:%s", err)
	}
	if headerGetter, ok := request.Interface().(syscallHeaderGetter); ok {
		header := headerGetter.GetHeader()
		if header == nil {
			// 如果Header字段没有赋值会导致后续的处理函数panic，强制赋值一个
			// FIXME:用更具有移植性的方法来赋值
			request.Elem().FieldByName("Header").Set(reflect.ValueOf(new(pb.SyscallHeader)))
			header = headerGetter.GetHeader()
		}
		header.Ctxid = ctxid
	}
	ret := m.Func.Call([]reflect.Value{
		s.vsyscall,
		reflect.ValueOf(ctx),
		request,
	})
	retErr := ret[1].Interface()
	if retErr != nil {
		return nil, retErr.(error)
	}
	response := ret[0].Interface().(proto.Message)
	responseBuf, err := proto.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("marshal response error:%s", err)
	}
	return responseBuf, nil
}
