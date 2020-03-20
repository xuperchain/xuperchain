// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tf.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type SyscallHeader struct {
	Ctxid                int64    `protobuf:"varint,1,opt,name=ctxid,proto3" json:"ctxid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SyscallHeader) Reset()         { *m = SyscallHeader{} }
func (m *SyscallHeader) String() string { return proto.CompactTextString(m) }
func (*SyscallHeader) ProtoMessage()    {}
func (*SyscallHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_375fc9137751f710, []int{0}
}

func (m *SyscallHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SyscallHeader.Unmarshal(m, b)
}
func (m *SyscallHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SyscallHeader.Marshal(b, m, deterministic)
}
func (m *SyscallHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyscallHeader.Merge(m, src)
}
func (m *SyscallHeader) XXX_Size() int {
	return xxx_messageInfo_SyscallHeader.Size(m)
}
func (m *SyscallHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_SyscallHeader.DiscardUnknown(m)
}

var xxx_messageInfo_SyscallHeader proto.InternalMessageInfo

func (m *SyscallHeader) GetCtxid() int64 {
	if m != nil {
		return m.Ctxid
	}
	return 0
}

type TrustFunctionCallRequest struct {
	Header               *SyscallHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Method               string         `protobuf:"bytes,2,opt,name=method,proto3" json:"method,omitempty"`
	Args                 string         `protobuf:"bytes,3,opt,name=args,proto3" json:"args,omitempty"`
	Svn                  uint32         `protobuf:"varint,4,opt,name=svn,proto3" json:"svn,omitempty"`
	Address              string         `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *TrustFunctionCallRequest) Reset()         { *m = TrustFunctionCallRequest{} }
func (m *TrustFunctionCallRequest) String() string { return proto.CompactTextString(m) }
func (*TrustFunctionCallRequest) ProtoMessage()    {}
func (*TrustFunctionCallRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_375fc9137751f710, []int{1}
}

func (m *TrustFunctionCallRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TrustFunctionCallRequest.Unmarshal(m, b)
}
func (m *TrustFunctionCallRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TrustFunctionCallRequest.Marshal(b, m, deterministic)
}
func (m *TrustFunctionCallRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TrustFunctionCallRequest.Merge(m, src)
}
func (m *TrustFunctionCallRequest) XXX_Size() int {
	return xxx_messageInfo_TrustFunctionCallRequest.Size(m)
}
func (m *TrustFunctionCallRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TrustFunctionCallRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TrustFunctionCallRequest proto.InternalMessageInfo

func (m *TrustFunctionCallRequest) GetHeader() *SyscallHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *TrustFunctionCallRequest) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *TrustFunctionCallRequest) GetArgs() string {
	if m != nil {
		return m.Args
	}
	return ""
}

func (m *TrustFunctionCallRequest) GetSvn() uint32 {
	if m != nil {
		return m.Svn
	}
	return 0
}

func (m *TrustFunctionCallRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type KVPair struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                string   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVPair) Reset()         { *m = KVPair{} }
func (m *KVPair) String() string { return proto.CompactTextString(m) }
func (*KVPair) ProtoMessage()    {}
func (*KVPair) Descriptor() ([]byte, []int) {
	return fileDescriptor_375fc9137751f710, []int{2}
}

func (m *KVPair) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVPair.Unmarshal(m, b)
}
func (m *KVPair) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVPair.Marshal(b, m, deterministic)
}
func (m *KVPair) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVPair.Merge(m, src)
}
func (m *KVPair) XXX_Size() int {
	return xxx_messageInfo_KVPair.Size(m)
}
func (m *KVPair) XXX_DiscardUnknown() {
	xxx_messageInfo_KVPair.DiscardUnknown(m)
}

var xxx_messageInfo_KVPair proto.InternalMessageInfo

func (m *KVPair) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *KVPair) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type KVPairs struct {
	Kv                   []*KVPair `protobuf:"bytes,1,rep,name=kv,proto3" json:"kv,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *KVPairs) Reset()         { *m = KVPairs{} }
func (m *KVPairs) String() string { return proto.CompactTextString(m) }
func (*KVPairs) ProtoMessage()    {}
func (*KVPairs) Descriptor() ([]byte, []int) {
	return fileDescriptor_375fc9137751f710, []int{3}
}

func (m *KVPairs) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVPairs.Unmarshal(m, b)
}
func (m *KVPairs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVPairs.Marshal(b, m, deterministic)
}
func (m *KVPairs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVPairs.Merge(m, src)
}
func (m *KVPairs) XXX_Size() int {
	return xxx_messageInfo_KVPairs.Size(m)
}
func (m *KVPairs) XXX_DiscardUnknown() {
	xxx_messageInfo_KVPairs.DiscardUnknown(m)
}

var xxx_messageInfo_KVPairs proto.InternalMessageInfo

func (m *KVPairs) GetKv() []*KVPair {
	if m != nil {
		return m.Kv
	}
	return nil
}

// result of trust call must return a key-value array, key is plain, and value is cipher,
// then be persisted by put_object.
type TrustFunctionCallResponse struct {
	// Types that are valid to be assigned to Results:
	//	*TrustFunctionCallResponse_Plaintext
	//	*TrustFunctionCallResponse_Kvs
	Results              isTrustFunctionCallResponse_Results `protobuf_oneof:"results"`
	XXX_NoUnkeyedLiteral struct{}                            `json:"-"`
	XXX_unrecognized     []byte                              `json:"-"`
	XXX_sizecache        int32                               `json:"-"`
}

func (m *TrustFunctionCallResponse) Reset()         { *m = TrustFunctionCallResponse{} }
func (m *TrustFunctionCallResponse) String() string { return proto.CompactTextString(m) }
func (*TrustFunctionCallResponse) ProtoMessage()    {}
func (*TrustFunctionCallResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_375fc9137751f710, []int{4}
}

func (m *TrustFunctionCallResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TrustFunctionCallResponse.Unmarshal(m, b)
}
func (m *TrustFunctionCallResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TrustFunctionCallResponse.Marshal(b, m, deterministic)
}
func (m *TrustFunctionCallResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TrustFunctionCallResponse.Merge(m, src)
}
func (m *TrustFunctionCallResponse) XXX_Size() int {
	return xxx_messageInfo_TrustFunctionCallResponse.Size(m)
}
func (m *TrustFunctionCallResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TrustFunctionCallResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TrustFunctionCallResponse proto.InternalMessageInfo

type isTrustFunctionCallResponse_Results interface {
	isTrustFunctionCallResponse_Results()
}

type TrustFunctionCallResponse_Plaintext struct {
	Plaintext string `protobuf:"bytes,2,opt,name=plaintext,proto3,oneof"`
}

type TrustFunctionCallResponse_Kvs struct {
	Kvs *KVPairs `protobuf:"bytes,3,opt,name=kvs,proto3,oneof"`
}

func (*TrustFunctionCallResponse_Plaintext) isTrustFunctionCallResponse_Results() {}

func (*TrustFunctionCallResponse_Kvs) isTrustFunctionCallResponse_Results() {}

func (m *TrustFunctionCallResponse) GetResults() isTrustFunctionCallResponse_Results {
	if m != nil {
		return m.Results
	}
	return nil
}

func (m *TrustFunctionCallResponse) GetPlaintext() string {
	if x, ok := m.GetResults().(*TrustFunctionCallResponse_Plaintext); ok {
		return x.Plaintext
	}
	return ""
}

func (m *TrustFunctionCallResponse) GetKvs() *KVPairs {
	if x, ok := m.GetResults().(*TrustFunctionCallResponse_Kvs); ok {
		return x.Kvs
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*TrustFunctionCallResponse) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*TrustFunctionCallResponse_Plaintext)(nil),
		(*TrustFunctionCallResponse_Kvs)(nil),
	}
}

func init() {
	proto.RegisterType((*SyscallHeader)(nil), "SyscallHeader")
	proto.RegisterType((*TrustFunctionCallRequest)(nil), "TrustFunctionCallRequest")
	proto.RegisterType((*KVPair)(nil), "KVPair")
	proto.RegisterType((*KVPairs)(nil), "KVPairs")
	proto.RegisterType((*TrustFunctionCallResponse)(nil), "TrustFunctionCallResponse")
}

func init() { proto.RegisterFile("tf.proto", fileDescriptor_375fc9137751f710) }

var fileDescriptor_375fc9137751f710 = []byte{
	// 324 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x51, 0x5d, 0x4b, 0xeb, 0x40,
	0x14, 0x6c, 0x9a, 0x36, 0x69, 0x4e, 0xe9, 0xe5, 0xb2, 0x5c, 0xae, 0x2b, 0x88, 0x84, 0x88, 0x9a,
	0xa7, 0x44, 0xea, 0x3f, 0xa8, 0x20, 0x01, 0x5f, 0x64, 0x15, 0x1f, 0x7c, 0xdb, 0x26, 0xc7, 0x26,
	0x24, 0x4d, 0xe2, 0x7e, 0x84, 0xf6, 0xaf, 0xf8, 0x6b, 0x25, 0x9b, 0x88, 0x08, 0xbe, 0xcd, 0xcc,
	0xee, 0xcc, 0x39, 0xc3, 0x81, 0x85, 0x7a, 0x8b, 0x5a, 0xd1, 0xa8, 0x26, 0xb8, 0x84, 0xd5, 0xd3,
	0x51, 0xa6, 0xbc, 0xaa, 0x12, 0xe4, 0x19, 0x0a, 0xf2, 0x0f, 0xe6, 0xa9, 0x3a, 0x14, 0x19, 0xb5,
	0x7c, 0x2b, 0xb4, 0xd9, 0x40, 0x82, 0x0f, 0x0b, 0xe8, 0xb3, 0xd0, 0x52, 0xdd, 0xeb, 0x3a, 0x55,
	0x45, 0x53, 0xdf, 0xf1, 0xaa, 0x62, 0xf8, 0xae, 0x51, 0x2a, 0x72, 0x05, 0x4e, 0x6e, 0xcc, 0xc6,
	0xb3, 0x5c, 0xff, 0x89, 0x7e, 0x44, 0xb2, 0xf1, 0x95, 0xfc, 0x07, 0x67, 0x8f, 0x2a, 0x6f, 0x32,
	0x3a, 0xf5, 0xad, 0xd0, 0x63, 0x23, 0x23, 0x04, 0x66, 0x5c, 0xec, 0x24, 0xb5, 0x8d, 0x6a, 0x30,
	0xf9, 0x0b, 0xb6, 0xec, 0x6a, 0x3a, 0xf3, 0xad, 0x70, 0xc5, 0x7a, 0x48, 0x28, 0xb8, 0x3c, 0xcb,
	0x04, 0x4a, 0x49, 0xe7, 0xe6, 0xe3, 0x17, 0x0d, 0x6e, 0xc0, 0x79, 0x78, 0x79, 0xe4, 0x85, 0xe8,
	0x5d, 0x25, 0x1e, 0xcd, 0x1a, 0x1e, 0xeb, 0x61, 0x5f, 0xa7, 0xe3, 0x95, 0xc6, 0x71, 0xe4, 0x40,
	0x82, 0x00, 0xdc, 0xc1, 0x21, 0xc9, 0x09, 0x4c, 0xcb, 0x8e, 0x5a, 0xbe, 0x1d, 0x2e, 0xd7, 0x6e,
	0x34, 0xa8, 0x6c, 0x5a, 0x76, 0x41, 0x06, 0xa7, 0xbf, 0x34, 0x96, 0x6d, 0x53, 0x4b, 0x24, 0xe7,
	0xe0, 0xb5, 0x15, 0x2f, 0x6a, 0x85, 0x07, 0x35, 0x44, 0x27, 0x13, 0xf6, 0x2d, 0x91, 0x33, 0xb0,
	0xcb, 0x6e, 0x68, 0xb4, 0x5c, 0x2f, 0xc6, 0x58, 0x99, 0x4c, 0x58, 0x2f, 0x6f, 0x3c, 0x70, 0x05,
	0x4a, 0x5d, 0x29, 0xb9, 0xb9, 0x4e, 0xec, 0xd7, 0x8b, 0x5d, 0xa1, 0x72, 0xbd, 0x8d, 0xd2, 0x66,
	0x1f, 0x1f, 0x74, 0x8b, 0x22, 0xcd, 0x79, 0x51, 0xc7, 0x69, 0x23, 0x30, 0x56, 0x88, 0xdd, 0x3e,
	0x6e, 0xb7, 0x5b, 0xc7, 0xdc, 0xeb, 0xf6, 0x33, 0x00, 0x00, 0xff, 0xff, 0x32, 0x33, 0x99, 0xc3,
	0xbb, 0x01, 0x00, 0x00,
}
