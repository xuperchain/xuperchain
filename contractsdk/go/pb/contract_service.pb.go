// Code generated by protoc-gen-go. DO NOT EDIT.
// source: contract_service.proto

package xchain_contract_sdk

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("contract_service.proto", fileDescriptor_e663a77702825514) }

var fileDescriptor_e663a77702825514 = []byte{
	// 376 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x94, 0xc1, 0x4e, 0xea, 0x40,
	0x14, 0x86, 0x73, 0xef, 0x25, 0x97, 0xcb, 0x81, 0xdc, 0xc5, 0x98, 0xb8, 0x20, 0x31, 0xa2, 0x22,
	0xe8, 0xa6, 0x0b, 0x7c, 0x02, 0xc1, 0x04, 0x8d, 0x49, 0x41, 0xc1, 0x90, 0x90, 0x18, 0x53, 0x86,
	0x23, 0x56, 0x9a, 0x0e, 0xce, 0x9c, 0x22, 0xbc, 0x96, 0x6f, 0xe2, 0x1b, 0x19, 0xca, 0x0c, 0xb0,
	0x98, 0xb6, 0x6e, 0x5c, 0xb6, 0xff, 0xf7, 0x7f, 0x3d, 0x39, 0x99, 0x29, 0xec, 0x73, 0x11, 0x92,
	0xf4, 0x38, 0x3d, 0x29, 0x94, 0x73, 0x9f, 0xa3, 0x33, 0x93, 0x82, 0x04, 0xdb, 0x5b, 0xf0, 0x17,
	0xcf, 0x0f, 0x1d, 0x13, 0x3b, 0x6a, 0x3c, 0x2d, 0xff, 0xdf, 0x3c, 0xc5, 0x50, 0xe3, 0xe3, 0x17,
	0x80, 0xeb, 0x91, 0x3f, 0xc7, 0x96, 0x18, 0x23, 0x1b, 0x40, 0xae, 0xe5, 0x05, 0x01, 0xab, 0x39,
	0x96, 0xb2, 0xa3, 0x41, 0x2f, 0x08, 0xee, 0xf1, 0x2d, 0x42, 0x45, 0xe5, 0x7a, 0x26, 0xa7, 0x66,
	0x22, 0x54, 0xc8, 0x6e, 0x21, 0xd7, 0xf5, 0xc3, 0x09, 0xab, 0x58, 0x0b, 0xab, 0xc8, 0x28, 0x8f,
	0x52, 0x88, 0xb5, 0xac, 0xf1, 0x99, 0x87, 0x7c, 0x6f, 0xa9, 0xf8, 0x6a, 0x52, 0x17, 0x0a, 0xdd,
	0x88, 0x3a, 0xa3, 0x57, 0xe4, 0xc4, 0x0e, 0xed, 0xdd, 0x88, 0x8c, 0xbc, 0x92, 0x0c, 0xe8, 0x41,
	0x5d, 0x28, 0xb4, 0x31, 0xdd, 0xd7, 0xc6, 0x0c, 0x5f, 0x0c, 0x68, 0xdf, 0x00, 0x4a, 0x57, 0x18,
	0x20, 0xa1, 0x56, 0x1e, 0x5b, 0x1b, 0x6b, 0xc4, 0x58, 0x4f, 0x52, 0x19, 0x2d, 0x1e, 0x42, 0xd1,
	0xc5, 0xf7, 0x1b, 0x42, 0xe9, 0x91, 0x90, 0xac, 0x6a, 0xed, 0x98, 0xd8, 0x98, 0x4f, 0x33, 0x28,
	0xed, 0xee, 0x43, 0xfe, 0x2e, 0x42, 0xb9, 0xec, 0x2f, 0x98, 0x7d, 0x16, 0x9d, 0x1a, 0x6d, 0x35,
	0x1d, 0xd2, 0xd6, 0x47, 0x80, 0xf8, 0x55, 0x33, 0x10, 0x7c, 0x9a, 0x70, 0xc4, 0xb6, 0x40, 0xfa,
	0x11, 0xdb, 0xe5, 0x36, 0x9b, 0xfe, 0xd7, 0x97, 0x5e, 0xa8, 0x9e, 0x31, 0x69, 0x1b, 0x26, 0x4e,
	0xdf, 0xc6, 0x96, 0xd2, 0x62, 0x0e, 0xa5, 0x96, 0x06, 0xe2, 0xcb, 0x71, 0x66, 0xad, 0xed, 0x22,
	0xe6, 0x03, 0xe7, 0xdf, 0x20, 0x7f, 0xe0, 0x82, 0xb0, 0x07, 0x28, 0xb6, 0x31, 0xf6, 0x5f, 0xca,
	0x89, 0x62, 0xf5, 0xa4, 0x53, 0x6a, 0x08, 0xa3, 0x3e, 0xb0, 0xcf, 0x6b, 0x3c, 0x43, 0x28, 0xf4,
	0x90, 0x3a, 0x11, 0xcd, 0x22, 0x62, 0xf6, 0xe5, 0x6d, 0x72, 0xa3, 0xac, 0x65, 0x61, 0xeb, 0x91,
	0x9b, 0xbf, 0xaf, 0xff, 0x8c, 0xfe, 0xc6, 0xff, 0xa4, 0x8b, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x35, 0xa4, 0x9a, 0xcf, 0xd2, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NativeCodeClient is the client API for NativeCode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NativeCodeClient interface {
	Call(ctx context.Context, in *NativeCallRequest, opts ...grpc.CallOption) (*NativeCallResponse, error)
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
}

type nativeCodeClient struct {
	cc *grpc.ClientConn
}

func NewNativeCodeClient(cc *grpc.ClientConn) NativeCodeClient {
	return &nativeCodeClient{cc}
}

func (c *nativeCodeClient) Call(ctx context.Context, in *NativeCallRequest, opts ...grpc.CallOption) (*NativeCallResponse, error) {
	out := new(NativeCallResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.NativeCode/Call", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nativeCodeClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.NativeCode/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NativeCodeServer is the server API for NativeCode service.
type NativeCodeServer interface {
	Call(context.Context, *NativeCallRequest) (*NativeCallResponse, error)
	Ping(context.Context, *PingRequest) (*PingResponse, error)
}

// UnimplementedNativeCodeServer can be embedded to have forward compatible implementations.
type UnimplementedNativeCodeServer struct {
}

func (*UnimplementedNativeCodeServer) Call(ctx context.Context, req *NativeCallRequest) (*NativeCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Call not implemented")
}
func (*UnimplementedNativeCodeServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func RegisterNativeCodeServer(s *grpc.Server, srv NativeCodeServer) {
	s.RegisterService(&_NativeCode_serviceDesc, srv)
}

func _NativeCode_Call_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NativeCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NativeCodeServer).Call(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.NativeCode/Call",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NativeCodeServer).Call(ctx, req.(*NativeCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NativeCode_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NativeCodeServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.NativeCode/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NativeCodeServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NativeCode_serviceDesc = grpc.ServiceDesc{
	ServiceName: "xchain.contract.sdk.NativeCode",
	HandlerType: (*NativeCodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Call",
			Handler:    _NativeCode_Call_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _NativeCode_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "contract_service.proto",
}

// SyscallClient is the client API for Syscall service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SyscallClient interface {
	// KV service
	PutObject(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error)
	GetObject(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	DeleteObject(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	NewIterator(ctx context.Context, in *IteratorRequest, opts ...grpc.CallOption) (*IteratorResponse, error)
	// Chain service
	QueryTx(ctx context.Context, in *QueryTxRequest, opts ...grpc.CallOption) (*QueryTxResponse, error)
	QueryBlock(ctx context.Context, in *QueryBlockRequest, opts ...grpc.CallOption) (*QueryBlockResponse, error)
	Transfer(ctx context.Context, in *TransferRequest, opts ...grpc.CallOption) (*TransferResponse, error)
	ContractCall(ctx context.Context, in *ContractCallRequest, opts ...grpc.CallOption) (*ContractCallResponse, error)
	// Heartbeat
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	GetCallArgs(ctx context.Context, in *GetCallArgsRequest, opts ...grpc.CallOption) (*CallArgs, error)
	SetOutput(ctx context.Context, in *SetOutputRequest, opts ...grpc.CallOption) (*SetOutputResponse, error)
}

type syscallClient struct {
	cc *grpc.ClientConn
}

func NewSyscallClient(cc *grpc.ClientConn) SyscallClient {
	return &syscallClient{cc}
}

func (c *syscallClient) PutObject(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error) {
	out := new(PutResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/PutObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) GetObject(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/GetObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) DeleteObject(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/DeleteObject", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) NewIterator(ctx context.Context, in *IteratorRequest, opts ...grpc.CallOption) (*IteratorResponse, error) {
	out := new(IteratorResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/NewIterator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) QueryTx(ctx context.Context, in *QueryTxRequest, opts ...grpc.CallOption) (*QueryTxResponse, error) {
	out := new(QueryTxResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/QueryTx", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) QueryBlock(ctx context.Context, in *QueryBlockRequest, opts ...grpc.CallOption) (*QueryBlockResponse, error) {
	out := new(QueryBlockResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/QueryBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) Transfer(ctx context.Context, in *TransferRequest, opts ...grpc.CallOption) (*TransferResponse, error) {
	out := new(TransferResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/Transfer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) ContractCall(ctx context.Context, in *ContractCallRequest, opts ...grpc.CallOption) (*ContractCallResponse, error) {
	out := new(ContractCallResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/ContractCall", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) GetCallArgs(ctx context.Context, in *GetCallArgsRequest, opts ...grpc.CallOption) (*CallArgs, error) {
	out := new(CallArgs)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/GetCallArgs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syscallClient) SetOutput(ctx context.Context, in *SetOutputRequest, opts ...grpc.CallOption) (*SetOutputResponse, error) {
	out := new(SetOutputResponse)
	err := c.cc.Invoke(ctx, "/xchain.contract.sdk.Syscall/SetOutput", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SyscallServer is the server API for Syscall service.
type SyscallServer interface {
	// KV service
	PutObject(context.Context, *PutRequest) (*PutResponse, error)
	GetObject(context.Context, *GetRequest) (*GetResponse, error)
	DeleteObject(context.Context, *DeleteRequest) (*DeleteResponse, error)
	NewIterator(context.Context, *IteratorRequest) (*IteratorResponse, error)
	// Chain service
	QueryTx(context.Context, *QueryTxRequest) (*QueryTxResponse, error)
	QueryBlock(context.Context, *QueryBlockRequest) (*QueryBlockResponse, error)
	Transfer(context.Context, *TransferRequest) (*TransferResponse, error)
	ContractCall(context.Context, *ContractCallRequest) (*ContractCallResponse, error)
	// Heartbeat
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	GetCallArgs(context.Context, *GetCallArgsRequest) (*CallArgs, error)
	SetOutput(context.Context, *SetOutputRequest) (*SetOutputResponse, error)
}

// UnimplementedSyscallServer can be embedded to have forward compatible implementations.
type UnimplementedSyscallServer struct {
}

func (*UnimplementedSyscallServer) PutObject(ctx context.Context, req *PutRequest) (*PutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutObject not implemented")
}
func (*UnimplementedSyscallServer) GetObject(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (*UnimplementedSyscallServer) DeleteObject(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteObject not implemented")
}
func (*UnimplementedSyscallServer) NewIterator(ctx context.Context, req *IteratorRequest) (*IteratorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewIterator not implemented")
}
func (*UnimplementedSyscallServer) QueryTx(ctx context.Context, req *QueryTxRequest) (*QueryTxResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTx not implemented")
}
func (*UnimplementedSyscallServer) QueryBlock(ctx context.Context, req *QueryBlockRequest) (*QueryBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryBlock not implemented")
}
func (*UnimplementedSyscallServer) Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transfer not implemented")
}
func (*UnimplementedSyscallServer) ContractCall(ctx context.Context, req *ContractCallRequest) (*ContractCallResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ContractCall not implemented")
}
func (*UnimplementedSyscallServer) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedSyscallServer) GetCallArgs(ctx context.Context, req *GetCallArgsRequest) (*CallArgs, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCallArgs not implemented")
}
func (*UnimplementedSyscallServer) SetOutput(ctx context.Context, req *SetOutputRequest) (*SetOutputResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetOutput not implemented")
}

func RegisterSyscallServer(s *grpc.Server, srv SyscallServer) {
	s.RegisterService(&_Syscall_serviceDesc, srv)
}

func _Syscall_PutObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).PutObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/PutObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).PutObject(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_GetObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).GetObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/GetObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).GetObject(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_DeleteObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).DeleteObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/DeleteObject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).DeleteObject(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_NewIterator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IteratorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).NewIterator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/NewIterator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).NewIterator(ctx, req.(*IteratorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_QueryTx_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTxRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).QueryTx(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/QueryTx",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).QueryTx(ctx, req.(*QueryTxRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_QueryBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).QueryBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/QueryBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).QueryBlock(ctx, req.(*QueryBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_Transfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransferRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).Transfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/Transfer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).Transfer(ctx, req.(*TransferRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_ContractCall_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContractCallRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).ContractCall(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/ContractCall",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).ContractCall(ctx, req.(*ContractCallRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_GetCallArgs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCallArgsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).GetCallArgs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/GetCallArgs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).GetCallArgs(ctx, req.(*GetCallArgsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Syscall_SetOutput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetOutputRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyscallServer).SetOutput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xchain.contract.sdk.Syscall/SetOutput",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyscallServer).SetOutput(ctx, req.(*SetOutputRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Syscall_serviceDesc = grpc.ServiceDesc{
	ServiceName: "xchain.contract.sdk.Syscall",
	HandlerType: (*SyscallServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PutObject",
			Handler:    _Syscall_PutObject_Handler,
		},
		{
			MethodName: "GetObject",
			Handler:    _Syscall_GetObject_Handler,
		},
		{
			MethodName: "DeleteObject",
			Handler:    _Syscall_DeleteObject_Handler,
		},
		{
			MethodName: "NewIterator",
			Handler:    _Syscall_NewIterator_Handler,
		},
		{
			MethodName: "QueryTx",
			Handler:    _Syscall_QueryTx_Handler,
		},
		{
			MethodName: "QueryBlock",
			Handler:    _Syscall_QueryBlock_Handler,
		},
		{
			MethodName: "Transfer",
			Handler:    _Syscall_Transfer_Handler,
		},
		{
			MethodName: "ContractCall",
			Handler:    _Syscall_ContractCall_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _Syscall_Ping_Handler,
		},
		{
			MethodName: "GetCallArgs",
			Handler:    _Syscall_GetCallArgs_Handler,
		},
		{
			MethodName: "SetOutput",
			Handler:    _Syscall_SetOutput_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "contract_service.proto",
}
