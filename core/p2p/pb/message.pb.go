// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

package xuperp2p

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

type XuperMessage_MessageType int32

const (
	XuperMessage_SENDBLOCK                XuperMessage_MessageType = 0
	XuperMessage_POSTTX                   XuperMessage_MessageType = 1
	XuperMessage_BATCHPOSTTX              XuperMessage_MessageType = 2
	XuperMessage_GET_BLOCK                XuperMessage_MessageType = 3
	XuperMessage_PING                     XuperMessage_MessageType = 4
	XuperMessage_GET_BLOCKCHAINSTATUS     XuperMessage_MessageType = 5
	XuperMessage_GET_BLOCK_RES            XuperMessage_MessageType = 6
	XuperMessage_GET_BLOCKCHAINSTATUS_RES XuperMessage_MessageType = 7
	// 向邻近确认区块是否为最新状态区块
	XuperMessage_CONFIRM_BLOCKCHAINSTATUS     XuperMessage_MessageType = 8
	XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES XuperMessage_MessageType = 9
	XuperMessage_MSG_TYPE_NONE                XuperMessage_MessageType = 10
	// query RPC port information
	XuperMessage_GET_RPC_PORT     XuperMessage_MessageType = 11
	XuperMessage_GET_RPC_PORT_RES XuperMessage_MessageType = 12
	// get authentication information
	XuperMessage_GET_AUTHENTICATION     XuperMessage_MessageType = 13
	XuperMessage_GET_AUTHENTICATION_RES XuperMessage_MessageType = 14
	// chained-bft NEW_VIEW message
	XuperMessage_CHAINED_BFT_NEW_VIEW_MSG XuperMessage_MessageType = 15
	// chained-bft NEW_PROPOSAL message
	XuperMessage_CHAINED_BFT_NEW_PROPOSAL_MSG XuperMessage_MessageType = 16
	// chained-bft vote message
	XuperMessage_CHAINED_BFT_VOTE_MSG XuperMessage_MessageType = 17
	// broadcast new block id to other node
	XuperMessage_NEW_BLOCKID XuperMessage_MessageType = 18
	// new node used to add to network automatic
	XuperMessage_NEW_NODE XuperMessage_MessageType = 19
	// 消息头同步对(GET_HASHES <-> HASHES),
	// 发送方通过GET_HASHES消息询问区间范围内的所有区块哈希信息,
	// 接受方发送HASHES信息, 该消息携带其所知的区间范围内的BlockId列表
	XuperMessage_GET_HASHES XuperMessage_MessageType = 20
	XuperMessage_HASHES     XuperMessage_MessageType = 21
	// 消息对(GET_BLOCKS <-> BLOCKS),
	// 发送方通过GET_BLOCKS消息询问BlockId列表内的所有对应区块信息,
	// 接受方发送BLOCKS信息, 该消息携带具体Block
	XuperMessage_GET_BLOCKS XuperMessage_MessageType = 22
	XuperMessage_BLOCKS     XuperMessage_MessageType = 23
)

var XuperMessage_MessageType_name = map[int32]string{
	0:  "SENDBLOCK",
	1:  "POSTTX",
	2:  "BATCHPOSTTX",
	3:  "GET_BLOCK",
	4:  "PING",
	5:  "GET_BLOCKCHAINSTATUS",
	6:  "GET_BLOCK_RES",
	7:  "GET_BLOCKCHAINSTATUS_RES",
	8:  "CONFIRM_BLOCKCHAINSTATUS",
	9:  "CONFIRM_BLOCKCHAINSTATUS_RES",
	10: "MSG_TYPE_NONE",
	11: "GET_RPC_PORT",
	12: "GET_RPC_PORT_RES",
	13: "GET_AUTHENTICATION",
	14: "GET_AUTHENTICATION_RES",
	15: "CHAINED_BFT_NEW_VIEW_MSG",
	16: "CHAINED_BFT_NEW_PROPOSAL_MSG",
	17: "CHAINED_BFT_VOTE_MSG",
	18: "NEW_BLOCKID",
	19: "NEW_NODE",
	20: "GET_HASHES",
	21: "HASHES",
	22: "GET_BLOCKS",
	23: "BLOCKS",
}

var XuperMessage_MessageType_value = map[string]int32{
	"SENDBLOCK":                    0,
	"POSTTX":                       1,
	"BATCHPOSTTX":                  2,
	"GET_BLOCK":                    3,
	"PING":                         4,
	"GET_BLOCKCHAINSTATUS":         5,
	"GET_BLOCK_RES":                6,
	"GET_BLOCKCHAINSTATUS_RES":     7,
	"CONFIRM_BLOCKCHAINSTATUS":     8,
	"CONFIRM_BLOCKCHAINSTATUS_RES": 9,
	"MSG_TYPE_NONE":                10,
	"GET_RPC_PORT":                 11,
	"GET_RPC_PORT_RES":             12,
	"GET_AUTHENTICATION":           13,
	"GET_AUTHENTICATION_RES":       14,
	"CHAINED_BFT_NEW_VIEW_MSG":     15,
	"CHAINED_BFT_NEW_PROPOSAL_MSG": 16,
	"CHAINED_BFT_VOTE_MSG":         17,
	"NEW_BLOCKID":                  18,
	"NEW_NODE":                     19,
	"GET_HASHES":                   20,
	"HASHES":                       21,
	"GET_BLOCKS":                   22,
	"BLOCKS":                       23,
}

func (x XuperMessage_MessageType) String() string {
	return proto.EnumName(XuperMessage_MessageType_name, int32(x))
}

func (XuperMessage_MessageType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0, 0}
}

type XuperMessage_ErrorType int32

const (
	// success
	XuperMessage_SUCCESS XuperMessage_ErrorType = 0
	XuperMessage_NONE    XuperMessage_ErrorType = 1
	// common error
	XuperMessage_UNKNOW_ERROR             XuperMessage_ErrorType = 2
	XuperMessage_CHECK_SUM_ERROR          XuperMessage_ErrorType = 3
	XuperMessage_UNMARSHAL_MSG_BODY_ERROR XuperMessage_ErrorType = 4
	XuperMessage_CONNECT_REFUSE           XuperMessage_ErrorType = 5
	// block error
	XuperMessage_GET_BLOCKCHAIN_ERROR           XuperMessage_ErrorType = 6
	XuperMessage_BLOCKCHAIN_NOTEXIST            XuperMessage_ErrorType = 7
	XuperMessage_GET_BLOCK_ERROR                XuperMessage_ErrorType = 8
	XuperMessage_CONFIRM_BLOCKCHAINSTATUS_ERROR XuperMessage_ErrorType = 9
	XuperMessage_GET_AUTHENTICATION_ERROR       XuperMessage_ErrorType = 10
	XuperMessage_GET_AUTHENTICATION_NOT_PASS    XuperMessage_ErrorType = 11
)

var XuperMessage_ErrorType_name = map[int32]string{
	0:  "SUCCESS",
	1:  "NONE",
	2:  "UNKNOW_ERROR",
	3:  "CHECK_SUM_ERROR",
	4:  "UNMARSHAL_MSG_BODY_ERROR",
	5:  "CONNECT_REFUSE",
	6:  "GET_BLOCKCHAIN_ERROR",
	7:  "BLOCKCHAIN_NOTEXIST",
	8:  "GET_BLOCK_ERROR",
	9:  "CONFIRM_BLOCKCHAINSTATUS_ERROR",
	10: "GET_AUTHENTICATION_ERROR",
	11: "GET_AUTHENTICATION_NOT_PASS",
}

var XuperMessage_ErrorType_value = map[string]int32{
	"SUCCESS":                        0,
	"NONE":                           1,
	"UNKNOW_ERROR":                   2,
	"CHECK_SUM_ERROR":                3,
	"UNMARSHAL_MSG_BODY_ERROR":       4,
	"CONNECT_REFUSE":                 5,
	"GET_BLOCKCHAIN_ERROR":           6,
	"BLOCKCHAIN_NOTEXIST":            7,
	"GET_BLOCK_ERROR":                8,
	"CONFIRM_BLOCKCHAINSTATUS_ERROR": 9,
	"GET_AUTHENTICATION_ERROR":       10,
	"GET_AUTHENTICATION_NOT_PASS":    11,
}

func (x XuperMessage_ErrorType) String() string {
	return proto.EnumName(XuperMessage_ErrorType_name, int32(x))
}

func (XuperMessage_ErrorType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0, 1}
}

// XuperMessage is the message of Xuper p2p server
type XuperMessage struct {
	Header               *XuperMessage_MessageHeader `protobuf:"bytes,1,opt,name=Header,proto3" json:"Header,omitempty"`
	Data                 *XuperMessage_MessageData   `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                    `json:"-"`
	XXX_unrecognized     []byte                      `json:"-"`
	XXX_sizecache        int32                       `json:"-"`
}

func (m *XuperMessage) Reset()         { *m = XuperMessage{} }
func (m *XuperMessage) String() string { return proto.CompactTextString(m) }
func (*XuperMessage) ProtoMessage()    {}
func (*XuperMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0}
}

func (m *XuperMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_XuperMessage.Unmarshal(m, b)
}
func (m *XuperMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_XuperMessage.Marshal(b, m, deterministic)
}
func (m *XuperMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_XuperMessage.Merge(m, src)
}
func (m *XuperMessage) XXX_Size() int {
	return xxx_messageInfo_XuperMessage.Size(m)
}
func (m *XuperMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_XuperMessage.DiscardUnknown(m)
}

var xxx_messageInfo_XuperMessage proto.InternalMessageInfo

func (m *XuperMessage) GetHeader() *XuperMessage_MessageHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *XuperMessage) GetData() *XuperMessage_MessageData {
	if m != nil {
		return m.Data
	}
	return nil
}

// MessageHeader is the message header of Xuper p2p server
type XuperMessage_MessageHeader struct {
	Version string `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	// dataCheckSum is the message data checksum, it can be used check where the message have been received
	Logid                string                   `protobuf:"bytes,2,opt,name=logid,proto3" json:"logid,omitempty"`
	From                 string                   `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	Bcname               string                   `protobuf:"bytes,4,opt,name=bcname,proto3" json:"bcname,omitempty"`
	Type                 XuperMessage_MessageType `protobuf:"varint,5,opt,name=type,proto3,enum=xuperp2p.XuperMessage_MessageType" json:"type,omitempty"`
	DataCheckSum         uint32                   `protobuf:"varint,6,opt,name=dataCheckSum,proto3" json:"dataCheckSum,omitempty"`
	ErrorType            XuperMessage_ErrorType   `protobuf:"varint,7,opt,name=errorType,proto3,enum=xuperp2p.XuperMessage_ErrorType" json:"errorType,omitempty"`
	EnableCompress       bool                     `protobuf:"varint,8,opt,name=enableCompress,proto3" json:"enableCompress,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *XuperMessage_MessageHeader) Reset()         { *m = XuperMessage_MessageHeader{} }
func (m *XuperMessage_MessageHeader) String() string { return proto.CompactTextString(m) }
func (*XuperMessage_MessageHeader) ProtoMessage()    {}
func (*XuperMessage_MessageHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0, 0}
}

func (m *XuperMessage_MessageHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_XuperMessage_MessageHeader.Unmarshal(m, b)
}
func (m *XuperMessage_MessageHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_XuperMessage_MessageHeader.Marshal(b, m, deterministic)
}
func (m *XuperMessage_MessageHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_XuperMessage_MessageHeader.Merge(m, src)
}
func (m *XuperMessage_MessageHeader) XXX_Size() int {
	return xxx_messageInfo_XuperMessage_MessageHeader.Size(m)
}
func (m *XuperMessage_MessageHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_XuperMessage_MessageHeader.DiscardUnknown(m)
}

var xxx_messageInfo_XuperMessage_MessageHeader proto.InternalMessageInfo

func (m *XuperMessage_MessageHeader) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *XuperMessage_MessageHeader) GetLogid() string {
	if m != nil {
		return m.Logid
	}
	return ""
}

func (m *XuperMessage_MessageHeader) GetFrom() string {
	if m != nil {
		return m.From
	}
	return ""
}

func (m *XuperMessage_MessageHeader) GetBcname() string {
	if m != nil {
		return m.Bcname
	}
	return ""
}

func (m *XuperMessage_MessageHeader) GetType() XuperMessage_MessageType {
	if m != nil {
		return m.Type
	}
	return XuperMessage_SENDBLOCK
}

func (m *XuperMessage_MessageHeader) GetDataCheckSum() uint32 {
	if m != nil {
		return m.DataCheckSum
	}
	return 0
}

func (m *XuperMessage_MessageHeader) GetErrorType() XuperMessage_ErrorType {
	if m != nil {
		return m.ErrorType
	}
	return XuperMessage_SUCCESS
}

func (m *XuperMessage_MessageHeader) GetEnableCompress() bool {
	if m != nil {
		return m.EnableCompress
	}
	return false
}

// MessageData is the message data of Xuper p2p server
type XuperMessage_MessageData struct {
	// msgInfo is the message infomation, use protobuf coding style
	MsgInfo              []byte   `protobuf:"bytes,3,opt,name=msgInfo,proto3" json:"msgInfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *XuperMessage_MessageData) Reset()         { *m = XuperMessage_MessageData{} }
func (m *XuperMessage_MessageData) String() string { return proto.CompactTextString(m) }
func (*XuperMessage_MessageData) ProtoMessage()    {}
func (*XuperMessage_MessageData) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0, 1}
}

func (m *XuperMessage_MessageData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_XuperMessage_MessageData.Unmarshal(m, b)
}
func (m *XuperMessage_MessageData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_XuperMessage_MessageData.Marshal(b, m, deterministic)
}
func (m *XuperMessage_MessageData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_XuperMessage_MessageData.Merge(m, src)
}
func (m *XuperMessage_MessageData) XXX_Size() int {
	return xxx_messageInfo_XuperMessage_MessageData.Size(m)
}
func (m *XuperMessage_MessageData) XXX_DiscardUnknown() {
	xxx_messageInfo_XuperMessage_MessageData.DiscardUnknown(m)
}

var xxx_messageInfo_XuperMessage_MessageData proto.InternalMessageInfo

func (m *XuperMessage_MessageData) GetMsgInfo() []byte {
	if m != nil {
		return m.MsgInfo
	}
	return nil
}

func init() {
	proto.RegisterEnum("xuperp2p.XuperMessage_MessageType", XuperMessage_MessageType_name, XuperMessage_MessageType_value)
	proto.RegisterEnum("xuperp2p.XuperMessage_ErrorType", XuperMessage_ErrorType_name, XuperMessage_ErrorType_value)
	proto.RegisterType((*XuperMessage)(nil), "xuperp2p.XuperMessage")
	proto.RegisterType((*XuperMessage_MessageHeader)(nil), "xuperp2p.XuperMessage.MessageHeader")
	proto.RegisterType((*XuperMessage_MessageData)(nil), "xuperp2p.XuperMessage.MessageData")
}

func init() { proto.RegisterFile("message.proto", fileDescriptor_33c57e4bae7b9afd) }

var fileDescriptor_33c57e4bae7b9afd = []byte{
	// 685 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x94, 0xdb, 0x6e, 0xf2, 0x46,
	0x10, 0xc7, 0xc3, 0x19, 0x86, 0x43, 0x36, 0x13, 0x4a, 0xac, 0x34, 0x6a, 0x11, 0xaa, 0x5a, 0xae,
	0xb8, 0x48, 0xa5, 0x5e, 0x55, 0x95, 0x8c, 0xd9, 0x60, 0x2b, 0x61, 0xd7, 0xda, 0x5d, 0x72, 0xb8,
	0xb2, 0x9c, 0xc4, 0x49, 0xa3, 0x06, 0x8c, 0x0c, 0xa9, 0x9a, 0x37, 0xe8, 0x93, 0xf4, 0xb6, 0x4f,
	0x58, 0xa9, 0xda, 0xb5, 0x21, 0xe4, 0xf4, 0x7d, 0x57, 0xec, 0xfc, 0xff, 0xbf, 0xf1, 0xcc, 0xce,
	0x8e, 0x80, 0xe6, 0x2c, 0x5a, 0x2e, 0xc3, 0xfb, 0x68, 0xb0, 0x48, 0xe2, 0x55, 0x8c, 0xd5, 0xbf,
	0x9e, 0x16, 0x51, 0xb2, 0x38, 0x5e, 0xf4, 0xfe, 0x06, 0x68, 0x5c, 0xea, 0x60, 0x92, 0x02, 0xf8,
	0x2b, 0x94, 0xdd, 0x28, 0xbc, 0x8d, 0x12, 0x2b, 0xd7, 0xcd, 0xf5, 0xeb, 0xc7, 0x3f, 0x0c, 0xd6,
	0xec, 0x60, 0x9b, 0x1b, 0x64, 0xbf, 0x29, 0x2b, 0xb2, 0x1c, 0xfc, 0x05, 0x8a, 0xa3, 0x70, 0x15,
	0x5a, 0x79, 0x93, 0xdb, 0xfb, 0x72, 0xae, 0x26, 0x85, 0xe1, 0x0f, 0xff, 0xcd, 0x43, 0xf3, 0xd5,
	0x17, 0xd1, 0x82, 0xca, 0x9f, 0x51, 0xb2, 0x7c, 0x88, 0xe7, 0xa6, 0x91, 0x9a, 0x58, 0x87, 0xd8,
	0x86, 0xd2, 0x63, 0x7c, 0xff, 0x70, 0x6b, 0x8a, 0xd4, 0x44, 0x1a, 0x20, 0x42, 0xf1, 0x2e, 0x89,
	0x67, 0x56, 0xc1, 0x88, 0xe6, 0x8c, 0x1d, 0x28, 0x5f, 0xdf, 0xcc, 0xc3, 0x59, 0x64, 0x15, 0x8d,
	0x9a, 0x45, 0xba, 0xcb, 0xd5, 0xf3, 0x22, 0xb2, 0x4a, 0xdd, 0x5c, 0xbf, 0xf5, 0xb5, 0x2e, 0xd5,
	0xf3, 0x22, 0x12, 0x86, 0xc7, 0x1e, 0x34, 0x6e, 0xc3, 0x55, 0xe8, 0xfc, 0x1e, 0xdd, 0xfc, 0x21,
	0x9f, 0x66, 0x56, 0xb9, 0x9b, 0xeb, 0x37, 0xc5, 0x2b, 0x0d, 0x7f, 0x83, 0x5a, 0x94, 0x24, 0x71,
	0xa2, 0xd3, 0xac, 0x8a, 0x29, 0xd0, 0xfd, 0xa4, 0x00, 0x5d, 0x73, 0xe2, 0x25, 0x05, 0x7f, 0x84,
	0x56, 0x34, 0x0f, 0xaf, 0x1f, 0x23, 0x27, 0x9e, 0x2d, 0x92, 0x68, 0xb9, 0xb4, 0xaa, 0xdd, 0x5c,
	0xbf, 0x2a, 0xde, 0xa8, 0x87, 0x3f, 0x41, 0x7d, 0x6b, 0x8c, 0x7a, 0x5c, 0xb3, 0xe5, 0xbd, 0x37,
	0xbf, 0x8b, 0xcd, 0x04, 0x1a, 0x62, 0x1d, 0xf6, 0xfe, 0x2b, 0x6c, 0x48, 0x53, 0xa0, 0x09, 0x35,
	0x49, 0xd9, 0x68, 0x78, 0xc6, 0x9d, 0x53, 0xb2, 0x83, 0x00, 0x65, 0x9f, 0x4b, 0xa5, 0x2e, 0x49,
	0x0e, 0x77, 0xa1, 0x3e, 0xb4, 0x95, 0xe3, 0x66, 0x42, 0x5e, 0xb3, 0x63, 0xaa, 0x82, 0x94, 0x2d,
	0x60, 0x15, 0x8a, 0xbe, 0xc7, 0xc6, 0xa4, 0x88, 0x16, 0xb4, 0x37, 0x86, 0xe3, 0xda, 0x1e, 0x93,
	0xca, 0x56, 0x53, 0x49, 0x4a, 0xb8, 0x07, 0xcd, 0x8d, 0x13, 0x08, 0x2a, 0x49, 0x19, 0x8f, 0xc0,
	0xfa, 0x08, 0x36, 0x6e, 0x45, 0xbb, 0x0e, 0x67, 0x27, 0x9e, 0x98, 0xbc, 0xff, 0x5c, 0x15, 0xbb,
	0x70, 0xf4, 0x99, 0x6b, 0xf2, 0x6b, 0xba, 0xe0, 0x44, 0x8e, 0x03, 0x75, 0xe5, 0xd3, 0x80, 0x71,
	0x46, 0x09, 0x20, 0x81, 0x86, 0x2e, 0x28, 0x7c, 0x27, 0xf0, 0xb9, 0x50, 0xa4, 0x8e, 0x6d, 0x20,
	0xdb, 0x8a, 0x49, 0x6d, 0x60, 0x07, 0x50, 0xab, 0xf6, 0x54, 0xb9, 0x94, 0x29, 0xcf, 0xb1, 0x95,
	0xc7, 0x19, 0x69, 0xe2, 0x21, 0x74, 0xde, 0xeb, 0x26, 0xa7, 0x65, 0xda, 0xd5, 0x3d, 0xd0, 0x51,
	0x30, 0x3c, 0x51, 0x01, 0xa3, 0x17, 0xc1, 0xb9, 0x47, 0x2f, 0x82, 0x89, 0x1c, 0x93, 0x5d, 0xd3,
	0xee, 0x1b, 0xd7, 0x17, 0xdc, 0xe7, 0xd2, 0x3e, 0x33, 0x04, 0xd1, 0x93, 0xdb, 0x26, 0xce, 0xb9,
	0xa2, 0xc6, 0xd9, 0xd3, 0xd3, 0xd7, 0xbc, 0xb9, 0xa6, 0x37, 0x22, 0x88, 0x0d, 0xa8, 0x6a, 0x81,
	0xf1, 0x11, 0x25, 0xfb, 0xd8, 0x02, 0xd0, 0x4d, 0xb9, 0xb6, 0x74, 0xa9, 0x24, 0x6d, 0xfd, 0x70,
	0xd9, 0xf9, 0x9b, 0xb5, 0x67, 0x52, 0x25, 0xe9, 0x68, 0x2f, 0x3b, 0x1f, 0xf4, 0xfe, 0xc9, 0x43,
	0x6d, 0xb3, 0x69, 0x58, 0x87, 0x8a, 0x9c, 0x3a, 0x0e, 0x95, 0x92, 0xec, 0xe8, 0xf7, 0x34, 0x13,
	0xcb, 0xe9, 0x89, 0x4d, 0xd9, 0x29, 0xe3, 0x17, 0x01, 0x15, 0x82, 0x0b, 0x92, 0xc7, 0x7d, 0xd8,
	0x75, 0x5c, 0xea, 0x9c, 0x06, 0x72, 0x3a, 0xc9, 0xc4, 0x82, 0xbe, 0xfc, 0x94, 0x4d, 0x6c, 0x21,
	0xdd, 0xf4, 0x3e, 0xc1, 0x90, 0x8f, 0xae, 0x32, 0xb7, 0x88, 0x08, 0x2d, 0x87, 0x33, 0x46, 0x1d,
	0x3d, 0xdf, 0x93, 0xa9, 0xa4, 0xa4, 0xf4, 0x7e, 0x51, 0x32, 0xba, 0x8c, 0x07, 0xb0, 0xbf, 0xa5,
	0x32, 0xae, 0xe8, 0xa5, 0x27, 0x15, 0xa9, 0xe8, 0xca, 0x2f, 0x1b, 0x94, 0xd2, 0x55, 0xec, 0xc1,
	0x77, 0x9f, 0xee, 0x41, 0xca, 0xd4, 0xd6, 0x7b, 0xf6, 0xe6, 0xd9, 0x52, 0x17, 0xf0, 0x7b, 0xf8,
	0xf6, 0x03, 0x97, 0x71, 0x15, 0xf8, 0xb6, 0x94, 0xa4, 0x7e, 0x5d, 0x36, 0xff, 0x8d, 0x3f, 0xff,
	0x1f, 0x00, 0x00, 0xff, 0xff, 0x1b, 0xae, 0xfc, 0x85, 0x2c, 0x05, 0x00, 0x00,
}
