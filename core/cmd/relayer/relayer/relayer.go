package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	relay "github.com/xuperchain/xuperchain/core/cmd/relayer"
	relayerpb "github.com/xuperchain/xuperchain/core/cmd/relayer/pb"
	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/kv/kvdb"
)

// Relayer parameters for a relayer to be required
// deliverBlockCmd deliver involved
// queryBlockCmd query involved
// storage data involved
// queryMeta query meta involved
type Relayer struct {
	deliverBlockCmd *relay.DeliverBlockCommand
	queryBlockCmd   *relay.QueryBlockCommand
	storage         *relay.Storage
	queryMeta       *relayerpb.QueryMeta
}

// GetLastQueryBlockHeight return latest query block height
// deliver routine needs it
func (relayer *Relayer) GetLastQueryBlockHeight() int64 {
	return relayer.queryMeta.GetLastQueryBlockHeight()
}

// SaveBlockLoop save block header received from source chain
func (relayer *Relayer) SaveBlockLoop() {
	for {
		relayer.saveBlockLoop()
	}
}
func (relayer *Relayer) saveBlockLoop() {
	// 获取下一个需要获取的区块高度
	currHeight := relayer.queryMeta.GetLastQueryBlockHeight() + 1
	fmt.Println("[relayer] the beignning height:", currHeight)
	// 获取截止高度
	endHeight := currHeight + 1
	var err error
	for {
		endHeight, err = relayer.queryBlockCmd.GetLatestBlockHeightFromSrcChain()
		if err != nil {
			fmt.Println("[relayer] get latest block height failed, err:", err)
			time.Sleep(time.Duration(12) * time.Second)
			continue
		}
		break
	}
	fmt.Println("[relayer] latest block height in source chain is ", endHeight, "currHeight is ", currHeight)
	// 持续从原链获取区块头(按照高度)
	for currHeight < endHeight-3 && currHeight >= 0 {
		block, err := relayer.queryBlockCmd.FetchBlockFromSrcChain(currHeight)
		// 如果是not found, 可以等待
		// 如果是原链挂掉了, 需要重启换节点
		if common.NormalizedKVError(err) == common.ErrKVNotFound {
			fmt.Println("[relayer] fetch block from src chain failed, try to fetch agagin", "target height:", currHeight, "err:", err)
			time.Sleep(time.Duration(3) * time.Second)
			continue
		} else if err != nil {
			fmt.Println("[relayer] fetch block from src chain failed, try to fetch agagin", "target height:", currHeight, "err:", err)
			time.Sleep(time.Duration(100) * time.Millisecond)
			currHeight--
			continue
		}
		// 相关字段置空
		block.Transactions = nil
		block.MerkleTree = nil
		// failedTxs has been consideration into blockid
		//block.FailedTxs = nil

		blockBuf, pbErr := proto.Marshal(block)
		// 序列化数据有问题, 重新获取区块
		if pbErr != nil {
			fmt.Println("[relayer] proto.Marshal block failed, try to get again")
			continue
		}
		saveErr := relayer.storage.PutBlockHeader(block.GetBlockid(), blockBuf)
		if saveErr != nil {
			fmt.Println("[relayer] put block failed, err:", saveErr)
			panic("put block into storage error:" + saveErr.Error())
		}
		saveErr = relayer.storage.PutHeightBlockid(currHeight, block.GetBlockid())
		if saveErr != nil {
			fmt.Println("[relayer] put height blockid failed, err", saveErr)
			panic("put height blockid error:" + saveErr.Error())
		}
		// 更新queryMeta
		tmpQueryMeta := &relayerpb.QueryMeta{
			LastQueryBlockHeight: currHeight,
		}
		err = relayer.queryBlockCmd.UpdateQueryMeta(tmpQueryMeta)
		if err != nil {
			panic("update query meta failed, err:" + err.Error())
		}
		relayer.queryMeta = &relayerpb.QueryMeta{
			LastQueryBlockHeight: currHeight,
		}
		currHeight++
	}
}

// InitRelayer 初始化relayer
// 使用一个默认StorageConfig来初始化Storage
func (relayer *Relayer) InitRelayer(cfg *relay.NodeConfig) {
	// init storage
	// 默认配置
	storePath := "./"
	fileName := "xuper"
	defaultStorageConfig := &relay.StorageConfig{
		StorePath: storePath,
		FileName:  fileName,
		KVConfig: &kvdb.KVParameter{
			DBPath:                filepath.Join(storePath, fileName),
			KVEngineType:          "default",
			MemCacheSize:          128,
			FileHandlersCacheSize: 1024,
			OtherPaths:            []string{},
		},
	}
	storage, err := relay.NewStorage(defaultStorageConfig)
	if err != nil {
		fmt.Println("[relayer] new storage error ", err)
		return
	}
	relayer.storage = storage

	// init queryBlockCmd
	// 传递SrcChain配置以及存储实例
	queryBlockCmd := &relay.QueryBlockCommand{
		Cfg:     cfg.Chains.SrcChain,
		Storage: storage,
	}
	relayer.queryBlockCmd = queryBlockCmd
	// 初始化rpc链接
	relayer.queryBlockCmd.InitXchainClient()
	// 查看是否第一次从原链获取区块头
	queryMeta, queryMetaErr := relayer.queryBlockCmd.LoadQueryMeta()
	// 如果为not found, 说明第一次从原链获取区块头
	if common.NormalizedKVError(queryMetaErr) == common.ErrKVNotFound {
		relayer.queryMeta = &relayerpb.QueryMeta{
			LastQueryBlockHeight: cfg.AnchorBlockHeight - 1,
		}
	} else if queryMetaErr != nil {
		panic("load query meta failed, error:" + queryMetaErr.Error())
	}
	relayer.queryMeta = &relayerpb.QueryMeta{
		LastQueryBlockHeight: queryMeta.GetLastQueryBlockHeight(),
	}

	// init deliverBlockCmd
	// 传递DstChain配置参数以及Storage实例以及锚点区块高度
	deliverBlockCmd := &relay.DeliverBlockCommand{
		Cfg:               cfg.Chains.DstChain,
		Storage:           storage,
		AnchorBlockHeight: cfg.AnchorBlockHeight,
		QueryMeta:         relayer,
	}
	relayer.deliverBlockCmd = deliverBlockCmd
	relayer.deliverBlockCmd.InitXchainClient()

	return
}

func main() {
	relayer := &Relayer{}
	// 读取配置文件
	cfg := relay.NewNodeConfig()
	cfg.LoadConfig()
	// 初始化relayer
	relayer.InitRelayer(cfg)
	// 开始从原链获取区块头
	go relayer.SaveBlockLoop()
	// 开始向目标链发送区块头
	go relayer.deliverBlockCmd.Deliver()

	// 注册优雅关停信号，包括ctrl + C 和kill信号
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigc)

	for {
		select {
		case <-sigc:
			fmt.Println("Got terminate, start to shutting down, please wait...")
			return
		}
	}
	return
}
