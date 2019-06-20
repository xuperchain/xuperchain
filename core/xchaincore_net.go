package xchaincore

import (
	"context"

	"github.com/golang/protobuf/proto"
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
	if xc.IsCoreMiner() {
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
		err = proto.Unmarshal(v.GetData().GetMsgInfo(), block)
		if err != nil {
			xc.log.Warn("BroadCastGetBlock unmarshal error", "error", err)
			continue
		} else {
			if block.Block == nil {
				continue
			}
			return block
		}
	}
	return nil
}
