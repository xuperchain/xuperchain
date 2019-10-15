package xchaincore

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"

	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
	"github.com/xuperchain/xuperunion/pb"
)

func handleSendBlockMsgFromRemoteNode(msg *xuper_p2p.XuperMessage) (*pb.Block, error) {
	block := &pb.Block{}
	got, err := handleMsgFromRemoteNode(msg)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(got, block)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func handleBatchPostTxMsgFromRemoteNode(msg *xuper_p2p.XuperMessage) (*pb.BatchTxs, error) {
	batchTxs := &pb.BatchTxs{}
	got, err := handleMsgFromRemoteNode(msg)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(got, batchTxs)
	if err != nil {
		return nil, err
	}
	return batchTxs, nil
}

func handlePostTxMsgFromRemoteNode(msg *xuper_p2p.XuperMessage) (*pb.TxStatus, error) {
	txStatus := &pb.TxStatus{}
	got, err := handleMsgFromRemoteNode(msg)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(got, txStatus)
	if err != nil {
		return nil, err
	}
	return txStatus, nil
}

func handleMsgFromRemoteNode(msg *xuper_p2p.XuperMessage) ([]byte, error) {
	originalMsg := msg.GetData().GetMsgInfo()
	var uncompressedMsg []byte
	var decodeErr error
	msgHeader := msg.GetHeader()
	if msgHeader != nil && msgHeader.GetCompressed() {
		uncompressedMsg, decodeErr = snappy.Decode(nil, originalMsg)
		if decodeErr != nil {
			return nil, decodeErr
		}
	} else {
		uncompressedMsg = originalMsg
	}
	return uncompressedMsg, nil
}

func (xm *XChainMG) processMsgToBeBroadcasted(msg []byte, hasCompressed bool) ([]byte, bool) {
	// case1: 信息本来就被压缩了，并且需要压缩转发
	if hasCompressed && xm.enableCompressed {
		return msg, true
	}
	// case2: 信息本来未被压缩，直接转发
	if !hasCompressed && !xm.enableCompressed {
		return msg, false
	}
	// case3: 信息本来被压缩了，需要以未压缩的形式转发
	if hasCompressed && !xm.enableCompressed {
		uncompressedMsg, err := snappy.Decode(nil, msg)
		if err != nil {
			return nil, false
		}
		return uncompressedMsg, false
	}
	// case4: 信息本来未被压缩，需要压缩转发
	if !hasCompressed && xm.enableCompressed {
		got := snappy.Encode(nil, msg)
		return got, true
	}
	return nil, false
}
