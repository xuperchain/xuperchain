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
