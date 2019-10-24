package xchaincore

import (
	"fmt"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/kv/kvdb"
	"github.com/xuperchain/xuperunion/pb"
)

func (xc *XChainCore) pruneLedger(bcname string, targetBlockid []byte, batch kvdb.Batch) error {
	// get target block
	targetBlock, err := xc.syncTargetBlock(bcname, targetBlockid)
	if err != nil {
		xc.log.Warn("pruneLedger syncTargetBlock error", "err", err)
		return err
	}
	// walk(tipBlockid, targetBlockid)
	walkErr := xc.Utxovm.Walk(targetBlockid)
	if walkErr != nil {
		xc.log.Warn("pruneLedger walk targetBlockid error", "walkErr", walkErr)
		return walkErr
	}
	// select all branchs higher than targetBlock.Height
	targetBlockidStr := fmt.Sprintf("%x", targetBlockid)
	branchHeadArr, branchErr := xc.Ledger.GetTargetRangeBranchInfo(targetBlockidStr, targetBlock.Height)
	if branchErr != nil {
		xc.log.Warn("pruneLedger GetTargetRangeBranchInfo error", "branchErr", branchErr)
		return branchErr
	}
	// 拿到所有的公共节点
	for _, v := range branchHeadArr {
		// get common parent from higher to lower and truncate all of them
		commonParentBlockid, err := xc.Ledger.GetCommonParentBlockid(targetBlockid, []byte(v))
		if err != nil {
			return err
		}
		fmt.Println("common parent blockid", commonParentBlockid)
		err = xc.Ledger.RemoveBlocks([]byte(v), commonParentBlockid, batch)
		if err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound {
			return err
		}
	}

	// update tipBlock and trunkHeight
	// TODO
	return nil
}

func (xc *XChainCore) syncTargetBlock(bcname string, targetBlockid []byte) (*pb.InternalBlock, error) {
	// check if targetBlockid is in branchInfo
	// case1: yes or error happen except not found error
	targetBlock, err := xc.Ledger.QueryBlock(targetBlockid)
	// if query success or error happen except not found error, return directly
	if err == nil || common.NormalizedKVError(err) != common.ErrKVNotFound {
		return nil, err
	}
	// case2: targetBlock not exist in local ledger branch, try to get it from remote nodes
	hd := &global.XContext{Timer: global.NewXTimer()}
	for {
		ib := xc.BroadCastGetBlock(&pb.BlockID{Header: &pb.Header{Logid: global.Glogid()}, Bcname: bcname, Blockid: targetBlockid, NeedContent: true})
		if ib == nil {
			xc.log.Warn("Can't Get a Block", "blockid", global.F(targetBlockid))
			continue
		}
		targetBlock = ib.GetBlock()
	}
	err = xc.SendBlock(
		&pb.Block{
			Header:  global.GHeader(),
			Bcname:  xc.bcname,
			Blockid: targetBlockid,
			Block:   targetBlock}, hd)
	if err != nil {
		xc.log.Warn("syncTargetBlock->SendBlock error", "err", err)
	}
	return targetBlock, err
}
