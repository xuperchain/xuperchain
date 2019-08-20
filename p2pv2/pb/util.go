package xuperp2p

import (
	"hash/crc32"

	"github.com/xuperchain/xuperunion/global"
)

// define message versions
const (
	XuperMsgVersion1 = "1.0.0"
	XuperMsgVersion2 = "2.0.0"
	XuperMsgVersion3 = "3.0.0"
)

// NewXuperMessage create P2P message instance with given params
func NewXuperMessage(version, bcName, lgid string, tp XuperMessage_MessageType, msgInfo []byte, ep XuperMessage_ErrorType) (*XuperMessage, error) {
	msg := &XuperMessage{
		Header: &XuperMessage_MessageHeader{
			Version: version,
			Bcname:  bcName,
			Type:    tp,
		},
		Data: &XuperMessage_MessageData{
			MsgInfo: msgInfo,
		},
	}
	if lgid == "" {
		msg.Header.Logid = global.Glogid()
	} else {
		msg.Header.Logid = lgid
	}
	msg.Header.DataCheckSum = CalDataCheckSum(msg)
	return msg, nil
}

// CalDataCheckSum calculate checksum of message
func CalDataCheckSum(msg *XuperMessage) uint32 {
	return crc32.ChecksumIEEE(msg.GetData().GetMsgInfo())
}

// VerifyDataCheckSum verify the checksum of message
func VerifyDataCheckSum(msg *XuperMessage) bool {
	return crc32.ChecksumIEEE(msg.GetData().GetMsgInfo()) == msg.GetHeader().GetDataCheckSum()
}

// VerifyMsgMatch 用于带返回的请求场景下验证收到的消息是否为预期的消息
func VerifyMsgMatch(msgRaw *XuperMessage, msgNew *XuperMessage, peerID string) bool {
	if msgNew.GetHeader().GetFrom() != peerID {
		return false
	}
	if msgRaw.GetHeader().GetLogid() != msgNew.GetHeader().GetLogid() {
		return false
	}
	switch msgRaw.GetHeader().GetType() {
	case XuperMessage_GET_BLOCK:
		if msgNew.GetHeader().GetType() == XuperMessage_GET_BLOCK_RES {
			return true
		}
		return false
	case XuperMessage_GET_BLOCKCHAINSTATUS:
		if msgNew.GetHeader().GetType() == XuperMessage_GET_BLOCKCHAINSTATUS_RES {
			return true
		}
		return false
	case XuperMessage_CONFIRM_BLOCKCHAINSTATUS:
		if msgNew.GetHeader().GetType() == XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES {
			return true
		}
		return false
	case XuperMessage_GET_AUTHENTICATION:
		if msgNew.GetHeader().GetType() == XuperMessage_GET_AUTHENTICATION_RES {
			return true
		}
		return false
	}

	return true
}

// GetResMsgType get the message type
func GetResMsgType(msgType XuperMessage_MessageType) XuperMessage_MessageType {
	switch msgType {
	case XuperMessage_GET_BLOCK:
		return XuperMessage_GET_BLOCK_RES
	case XuperMessage_GET_BLOCKCHAINSTATUS:
		return XuperMessage_GET_BLOCKCHAINSTATUS_RES
	case XuperMessage_CONFIRM_BLOCKCHAINSTATUS:
		return XuperMessage_CONFIRM_BLOCKCHAINSTATUS_RES
	case XuperMessage_GET_RPC_PORT:
		return XuperMessage_GET_RPC_PORT_RES
	case XuperMessage_GET_AUTHENTICATION:
		return XuperMessage_GET_AUTHENTICATION_RES
	default:
		return XuperMessage_MSG_TYPE_NONE
	}
}
