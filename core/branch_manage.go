package xchaincore

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
)

func (xc *XChainCore) pruneLedger(targetBlockid []byte) error {
	// get target block
	targetBlock, err := xc.syncTargetBlock(targetBlockid)
	if err != nil {
		xc.log.Warn("pruneLedger syncTargetBlock error", "err", err, "blockid", string(targetBlockid))
		return err
	}
	fmt.Println("------------- step1 syncTargetBlock succeed")
	// utxo 主干切换
	walkErr := xc.Utxovm.Walk(targetBlockid)
	if walkErr != nil {
		xc.log.Warn("pruneLedger walk targetBlockid error", "walkErr", walkErr)
		return walkErr
	}
	fmt.Println("------------- step2 utxo walk succeed")
	// ledger 主干切换
	batch := xc.Ledger.GetLDB().NewBatch()
	_, splitErr := xc.Ledger.HandleFork(xc.Ledger.GetMeta().TipBlockid, targetBlockid, batch)
	if splitErr != nil {
		return splitErr
	}
	fmt.Println("------------- step3 handle fork succeed")
	// ledger主干切换的扫尾工作
	newMeta := proto.Clone(xc.Ledger.GetMeta()).(*pb.LedgerMeta)
	newMeta.TrunkHeight = targetBlock.Height
	newMeta.TipBlockid = targetBlock.Blockid
	metaBuf, pbErr := proto.Marshal(newMeta)
	if pbErr != nil {
		return pbErr
	}
	batch.Put([]byte(pb.MetaTablePrefix), metaBuf)
	// 剪掉所有无效分支
	// step1: 获取所有无效分支
	branchHeadArr, branchErr := xc.Ledger.GetBranchInfo(targetBlockid, targetBlock.Height)
	if branchErr != nil {
		xc.log.Warn("pruneLedger GetTargetRangeBranchInfo error", "branchErr", branchErr)
		return branchErr
	}
	fmt.Println("------------- step4 get branch info succeed")
	// step2: 将无效分支剪掉
	for _, v := range branchHeadArr {
		// get common parent from higher to lower and truncate all of them
		commonParentBlockid, err := xc.Ledger.GetCommonParentBlockid(targetBlockid, []byte(v))
		if err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound {
			return err
		}
		err = xc.Ledger.RemoveBlocks([]byte(v), commonParentBlockid, batch)
		if err != nil && common.NormalizedKVError(err) != common.ErrKVNotFound {
			return err
		}
		fmt.Println("--------- before step5", fmt.Sprintf("%x", v))
	}
	fmt.Println("------------- step5 remove blocks succeed")
	kvErr := batch.Write()
	if kvErr != nil {
		return kvErr
	}
	xc.Ledger.SetMeta(newMeta)
	return nil
}

func (xc *XChainCore) syncTargetBlock(targetBlockid []byte) (*pb.InternalBlock, error) {
	// check if targetBlockid is in branchInfo
	// case1: yes or error happen except not found error
	targetBlock, err := xc.Ledger.QueryBlock(targetBlockid)
	// if query success or error happen except not found error, return
	if err == nil || common.NormalizedKVError(err) != common.ErrKVNotFound {
		return targetBlock, err
	}
	// case2: targetBlock not exist in local ledger branch, try to get it from remote nodes
	hd := &global.XContext{Timer: global.NewXTimer()}
	for {
		ib := xc.BroadCastGetBlock(&pb.BlockID{Header: &pb.Header{Logid: global.Glogid()}, Bcname: xc.bcname, Blockid: targetBlockid, NeedContent: true})
		if ib == nil {
			xc.log.Warn("Can't Get a Block", "blockid", global.F(targetBlockid))
			continue
		}
		targetBlock = ib.GetBlock()
		break
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
	return nil, nil
}
