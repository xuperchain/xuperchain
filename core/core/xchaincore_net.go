package xchaincore

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
	"github.com/xuperchain/xuperunion/pb"
)

// BroadCastGetBlock get block from p2p network nodes
func (xc *XChainCore) BroadCastGetBlock(bid *pb.BlockID) *pb.Block {
	msgbuf, err := proto.Marshal(bid)
	if err != nil {
		xc.log.Warn("BroadCastGetBlock Marshal msg error", "error", err)
		return nil
	}
	msg, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion2, bid.GetBcname(), "", xuper_p2p.XuperMessage_GET_BLOCK, msgbuf, xuper_p2p.XuperMessage_NONE)
	filters := []p2pv2.FilterStrategy{p2pv2.NearestBucketStrategy}
	if xc.NeedCoreConnection() {
		filters = append(filters, p2pv2.CorePeersStrategy)
	}
	opts := []p2pv2.MessageOption{
		p2pv2.WithFilters(filters),
		p2pv2.WithBcName(xc.bcname),
	}
	res, err := xc.P2pv2.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil || len(res) < 1 {
		return nil
	}

	for _, v := range res {
		if v.GetHeader().GetErrorType() != xuper_p2p.XuperMessage_SUCCESS {
			continue
		}

		block := &pb.Block{}
		blockBuf, err := xuper_p2p.Uncompress(v)
		if blockBuf == nil || err != nil {
			xc.log.Warn("BroadCastGetBlock xuper_p2p Uncompress error", "error", err)
			continue
		}
		err = proto.Unmarshal(blockBuf, block)
		if block == nil || block.GetBlock() == nil || err != nil {
			xc.log.Warn("BroadCastGetBlock unmarshal error", "error", err)
			continue
		}
		return block
	}
	return nil
}

// getBlockFromPeer get Block from given peer
func (xc *XChainCore) getBlockFromPeer(ctx context.Context, blockid []byte, remotePid string) (*pb.Block, error) {
	bid := &pb.BlockID{
		Blockid:     blockid,
		Bcname:      xc.bcname,
		NeedContent: true,
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
	}
	msgbuf, err := proto.Marshal(bid)
	if err != nil {
		xc.log.Warn("getBlockFromPeer Marshal msg error", "error", err)
		return nil, err
	}
	// send GET_BLOCK message to the remote peer
	msg, _ := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion2, bid.GetBcname(), "",
		xuper_p2p.XuperMessage_GET_BLOCK, msgbuf, xuper_p2p.XuperMessage_NONE)
	opts := []p2pv2.MessageOption{
		p2pv2.WithBcName(xc.bcname),
		p2pv2.WithTargetPeerIDs([]string{remotePid}),
	}
	res, err := xc.P2pv2.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil || len(res) < 1 {
		xc.log.Warn("getBlockFromPeer get error or empty response", "error", err, "msglen", len(res))
		return nil, errors.New("get block failed")
	}

	// get the block data in result
	for _, v := range res {
		if v.GetHeader().GetErrorType() != xuper_p2p.XuperMessage_SUCCESS {
			continue
		}

		block := &pb.Block{}
		blockBuf, err := xuper_p2p.Uncompress(v)
		if blockBuf == nil || err != nil {
			xc.log.Warn("getBlockFromPeer xuper_p2p Uncompress error", "error", err)
			continue
		}
		err = proto.Unmarshal(blockBuf, block)
		if block == nil || block.GetBlock() == nil || err != nil {
			xc.log.Warn("getBlockFromPeer unmarshal error", "error", err)
			continue
		}
		block.Blockid = blockid
		return block, nil
	}
	return nil, errors.New("get block failed, no block data")
}

// handleNewBlockID handle signal of New_BlockID
func (xc *XChainCore) handleNewBlockID(ctx context.Context, blockid []byte, remotePid string) (*pb.Block, error) {
	if len(blockid) == 0 || remotePid == "" {
		xc.log.Warn("handleNewBlockID: blockid or remotePid cannot be nil", "remotePid", remotePid)
		return nil, errors.New("block/remotePid is nil")
	}

	// dup check in message cache
	bidPretty := hex.EncodeToString(blockid)
	_, exist := xc.msgCache.Get(bidPretty)
	if exist {
		// this block id is processing or processed, ignore it
		xc.log.Info("Received block id but it's in cache, ignore it", "blockid", bidPretty, "peerid", remotePid)
		return nil, nil
	}

	// dup check in ledger
	if xc.Ledger.ExistBlock(blockid) {
		// this block id is exist in ledger, ignore it
		xc.log.Info("Received block id but it's in ledger, ignore it", "blockid", bidPretty, "peerid", remotePid)
		return nil, nil
	}

	// new block found, start processing
	if _, existNow := xc.msgCache.Get(bidPretty); !existNow {
		// double check cache then add cache
		xc.msgCache.Add(bidPretty, remotePid)
		block, err := xc.getBlockFromPeer(ctx, blockid, remotePid)
		if err != nil {
			xc.log.Warn("Received block id but get block failed",
				"blockid", bidPretty, "peerid", remotePid, "error", err)
			// remove block id since block process failed
			xc.msgCache.Del(bidPretty)
			return nil, err
		}
		return block, nil
	}
	xc.log.Trace("Received block id but it might be processed before, igore it",
		"blockid", bidPretty, "peerid", remotePid)
	return nil, nil
}
