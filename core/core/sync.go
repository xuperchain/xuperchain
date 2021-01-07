package xchaincore

import (
	"context"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/global"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuper_p2p "github.com/xuperchain/xuperchain/core/p2p/pb"
	"github.com/xuperchain/xuperchain/core/pb"
)

const (
	// MaxSyncTimes SyncBlocks 最大次数
	MaxSyncTimes = 5
	// MaxSleepMilSecond ...
	MaxSleepMilSecond = 500
)

// SyncBlocks sync block while start to miner
func (xc *XChainCore) SyncBlocks() {
	hd := &global.XContext{Timer: global.NewXTimer()}
	for i := 0; i < MaxSyncTimes; i++ {
		xc.log.Trace("sync blocks", "blockname", xc.bcname, "try times", i)
		bc, confirm := xc.syncForOnce()
		xc.log.Trace("sync blocks", "bc", bc, "confirm", confirm)
		if bc == nil || bc.GetBlock() == nil {
			time.Sleep(time.Duration(rand.Intn(MaxSleepMilSecond)) * time.Millisecond)
			continue
		}
		if !confirm && i < MaxSyncTimes-1 {
			time.Sleep(time.Duration(rand.Intn(MaxSleepMilSecond)) * time.Millisecond)
			continue
		}
		err := xc.SendBlock(
			&pb.Block{
				Header:  global.GHeader(),
				Bcname:  xc.bcname,
				Blockid: bc.Block.Blockid,
				Block:   bc.Block}, hd)
		if err == nil || err.Error() == ErrBlockExist.Error() {
			break
		}
	}
}

// syncForOnce sync block from peer nodes for one times
func (xc *XChainCore) syncForOnce() (*pb.BCStatus, bool) {
	bcs := &pb.BCStatus{Bcname: xc.bcname}
	bcsBuf, _ := proto.Marshal(bcs)
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, xc.bcname, "", xuper_p2p.XuperMessage_GET_BLOCKCHAINSTATUS, bcsBuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		xc.log.Warn("syncForOnce error", "error", err)
		return nil, false
	}
	filters := []p2p_base.FilterStrategy{p2p_base.NearestBucketStrategy}
	whiteList := xc.groupChain.GetAllowedPeersWithBcname(xc.bcname)
	opts := []p2p_base.MessageOption{
		p2p_base.WithFilters(filters),
		p2p_base.WithBcName(xc.bcname),
		p2p_base.WithWhiteList(whiteList),
	}
	hbcs, err := xc.P2pSvr.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil {
		return nil, false
	}
	hbc := countGetBlockChainStatus(hbcs)
	if hbcs == nil {
		return nil, false
	}
	return hbc, xc.syncConfirm(hbc)
}

// countGetBlockChainStatus 对p2p网络返回的结果进行统计
func countGetBlockChainStatus(hbcs []*xuper_p2p.XuperMessage) *pb.BCStatus {
	p := hbcs[0]
	maxCount := 0
	countHeight := make(map[int64]int)
	for i := 1; i < len(hbcs); i++ {
		bcStatus := &pb.BCStatus{}
		err := proto.Unmarshal(p.GetData().GetMsgInfo(), bcStatus)
		if err != nil {
			continue
		}
		countHeight[bcStatus.GetMeta().GetTrunkHeight()]++
		if countHeight[bcStatus.GetMeta().GetTrunkHeight()] >= maxCount {
			p = hbcs[i]
			maxCount = countHeight[bcStatus.GetMeta().GetTrunkHeight()]
		}
	}
	res := &pb.BCStatus{}
	err := proto.Unmarshal(p.GetData().GetMsgInfo(), res)
	if err != nil {
		return nil
	}
	return res
}

// syncConfirm 向周围节点询问块是否可以被接受
func (xc *XChainCore) syncConfirm(bcs *pb.BCStatus) bool {
	bcsBuf, err := proto.Marshal(bcs)
	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, bcs.GetBcname(), "", xuper_p2p.XuperMessage_CONFIRM_BLOCKCHAINSTATUS, bcsBuf, xuper_p2p.XuperMessage_NONE)
	filters := []p2p_base.FilterStrategy{p2p_base.NearestBucketStrategy}
	whiteList := xc.groupChain.GetAllowedPeersWithBcname(xc.bcname)
	opts := []p2p_base.MessageOption{
		p2p_base.WithFilters(filters),
		p2p_base.WithBcName(xc.bcname),
		p2p_base.WithWhiteList(whiteList),
	}
	res, err := xc.P2pSvr.SendMessageWithResponse(context.Background(), msg, opts...)
	if err != nil {
		return false
	}

	return countConfirmBlockRes(res)
}

// countConfirmBlockRes 对p2p网络返回的确认区块的结果进行统计
func countConfirmBlockRes(res []*xuper_p2p.XuperMessage) bool {
	// 统计邻近节点的返回信息
	agreeCnt := 0
	disAgresCnt := 0
	for i := 0; i < len(res); i++ {
		bts := &pb.BCTipStatus{}
		err := proto.Unmarshal(res[i].GetData().GetMsgInfo(), bts)
		if err != nil {
			continue
		}
		if bts.GetIsTrunkTip() {
			agreeCnt++
		} else {
			disAgresCnt++
		}
	}
	// 支持的节点需要大于反对的节点，并且支持的节点个数需要大于res的1/3
	return agreeCnt >= disAgresCnt && agreeCnt >= len(res)/3
}
