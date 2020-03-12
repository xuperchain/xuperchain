package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/xuperchain/xuperchain/core/common"
	relay "github.com/xuperchain/xuperchain/core/relayer"
	relayerpb "github.com/xuperchain/xuperchain/core/relayer/pb"
)

type Relayer struct {
	deliverBlockCmd *relay.DeliverBlockCommand
	queryBlockCmd   *relay.QueryBlockCommand
	storage         *relay.Storage
	deliverMeta     *relayerpb.DeliverMeta
}

func (relayer *Relayer) LoadDeliverMeta() (string, int64, error) {
	metaBuf, findErr := relayer.storage.LoadDeliverMeta()
	// 之前已经同步过区块
	if findErr == nil {
		meta := &relayerpb.DeliverMeta{}
		err := proto.Unmarshal(metaBuf, meta)
		if err != nil {
			return "", -1, err
		}
		return meta.GetCurrHash(), meta.GetHeight(), nil
	} else if common.NormalizedKVError(findErr) == common.ErrKVNotFound {
		// 第一次
		// step1: 读配置
		// step2: 0
		return "", 0, common.ErrKVNotFound
	}

	return "", -1, findErr
}

func (relayer *Relayer) InitRelayer(cfg *relay.NodeConfig) {
	// init queryBlockCmd
	queryBlockCmd := &relay.QueryBlockCommand{
		Cfg: cfg.Chains.SrcChain,
	}
	relayer.queryBlockCmd = queryBlockCmd
	relayer.queryBlockCmd.InitXchainClient()
	deliverBlockCmd := &relay.DeliverBlockCommand{
		Cfg: cfg.Chains.DstChain,
	}
	relayer.deliverBlockCmd = deliverBlockCmd
	relayer.deliverBlockCmd.InitXchainClient()
	// init storage
	storage, err := relay.NewStorage()
	if err != nil {
		fmt.Println("new storage error")
		return
	}
	relayer.storage = storage
	currHash, height, err := relayer.LoadDeliverMeta()
	// 还没有注入过LedgerMeta, 说明还没有注入「锚点区块头」
	if err == common.ErrKVNotFound {
		anchorBlock, err := relayer.queryBlockCmd.FetchBlockFromSrcChain(cfg.AnchorBlockHeight)
		if err != nil {
			fmt.Println("fetch block from src chain failed, try to fetch agagin")
			panic("fetch block from src chain failed")
		}
		// 相关字段置空
		anchorBlock.Transactions = nil
		anchorBlock.MerkleTree = nil
		anchorBlock.FailedTxs = nil

		anchorBlockHeaderBuf, pbErr := proto.Marshal(anchorBlock)
		if pbErr != nil {
			fmt.Println("proto.Marshal anchor block failed, try to fetch again")
			panic("proto.Marshal anchor block failed")
		}
		deliverErr := relayer.deliverBlockCmd.DeliverAnchorBlockHeader(anchorBlockHeaderBuf)
		if deliverErr != nil {
			fmt.Println("DeliverAnchorBlockHeader failed, err:", deliverErr)
			panic("DeliverAnchorBlockHeader failed")
		}
		tmpDeliverMeta := &relayerpb.DeliverMeta{
			CurrHash: fmt.Sprintf("%x", anchorBlock.GetBlockid()),
			Height:   anchorBlock.GetHeight(),
		}
		tmpDeliverMetaBuf, err := proto.Marshal(tmpDeliverMeta)
		err = relayer.storage.UpdateDeliverMeta(tmpDeliverMetaBuf)
		if err != nil {
			panic("should not be here")
		}
		relayer.deliverMeta = &relayerpb.DeliverMeta{
			CurrHash: tmpDeliverMeta.GetCurrHash(),
			Height:   tmpDeliverMeta.GetHeight(),
		}
		return
	} else if err != nil {
		panic("should not be here")
	}
	relayer.deliverMeta = &relayerpb.DeliverMeta{
		CurrHash: currHash,
		Height:   height,
	}
}

func (relayer *Relayer) SaveBlockLoop() {
	currHeight := relayer.deliverMeta.GetHeight() + 1
	fmt.Println("the beginning height:", currHeight)
	for {
		block, err := relayer.queryBlockCmd.FetchBlockFromSrcChain(currHeight)
		if err != nil {
			fmt.Println("fetch block from src chain failed, try to fetch agagin", "target height:", currHeight)
			time.Sleep(time.Duration(100) * time.Millisecond)
			currHeight--
			continue
		}
		// 相关字段置空
		block.Transactions = nil
		block.MerkleTree = nil
		block.FailedTxs = nil

		blockBuf, pbErr := proto.Marshal(block)
		if pbErr != nil {
			fmt.Println("proto.Marshal block failed, try to fetch again")
			continue
		}
		saveErr := relayer.storage.Put(block.GetBlockid(), blockBuf)
		if saveErr != nil {
			fmt.Println("put block failed, try to save again")
			continue
		}
		deliverErr := relayer.deliverBlockCmd.DeliverBlockHeader(blockBuf)
		if deliverErr != nil {
			if strings.Contains(deliverErr.Error(), "missing preHash") {
				fmt.Println("Deliver Block Header error:", deliverErr)
				currHeight--
				continue
			}
			fmt.Println("Deliver Block Header error:", deliverErr)
		}
		tmpDeliverMeta := &relayerpb.DeliverMeta{
			CurrHash: fmt.Sprintf("%x", block.GetBlockid()),
			Height:   block.GetHeight(),
		}
		tmpDeliverMetaBuf, err := proto.Marshal(tmpDeliverMeta)
		err = relayer.storage.UpdateDeliverMeta(tmpDeliverMetaBuf)
		if err != nil {
			panic("should not be here")
		}
		relayer.deliverMeta = &relayerpb.DeliverMeta{
			CurrHash: tmpDeliverMeta.GetCurrHash(),
			Height:   tmpDeliverMeta.GetHeight(),
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
		currHeight++
	}
}

func main() {
	relayer := &Relayer{}
	cfg := relay.NewNodeConfig()
	cfg.LoadConfig()
	relayer.InitRelayer(cfg)
	relayer.SaveBlockLoop()
}
