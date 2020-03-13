package relayer

import (
	//"context"
	//"encoding/hex"
	//"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	//"github.com/golang/protobuf/proto"

	//"github.com/xuperchain/xuperchain/core/contract"
	//"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	//"github.com/xuperchain/xuperchain/core/utxo"
	//"github.com/xuperchain/xuperchain/core/utxo/txhash"
	"github.com/xuperchain/xuperchain/core/common"
	relayerpb "github.com/xuperchain/xuperchain/core/relayer/pb"
)

type QueryMetaRegister interface {
	GetLastQueryBlockHeight() int64
}

type DeliverBlockCommand struct {
	client            pb.XchainClient
	Cfg               ChainConfig
	meta              *relayerpb.DeliverMeta
	AnchorBlockHeight int64
	Storage           *Storage
	QueryMeta         QueryMetaRegister
}

func (cmd *DeliverBlockCommand) InitXchainClient() error {
	conn, err := grpc.Dial(cmd.Cfg.RPCAddr, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	cmd.client = pb.NewXchainClient(conn)
	return nil
}

func (cmd *DeliverBlockCommand) Deliver() error {
	// 是否有必要调用DeliverAnchorBlockHeader()
	meta, err := cmd.LoadDeliverMeta()
	if common.NormalizedKVError(err) == common.ErrKVNotFound {
		cmd.DeliverAnchorBlockHeader()
	} else if err != nil {
		panic("get deliver meta failed, err:" + err.Error())
	} else {
		cmd.meta = &relayerpb.DeliverMeta{
			LastDeliverBlockHeight: meta.GetLastDeliverBlockHeight(),
		}
	}
	cmd.DeliverBlockHeader()
	return nil
}

func (cmd *DeliverBlockCommand) DeliverAnchorBlockHeader() error {
	for {
		blockBuf, err := cmd.Storage.GetBlockHeaderByHeight(cmd.AnchorBlockHeight)
		// 本地还没有存储AnchorBlock, 等待
		if common.NormalizedKVError(err) == common.ErrKVNotFound {
			time.Sleep(time.Duration(1) * time.Second)
			continue
		} else if err != nil {
			panic("get anchor block header failed, err:" + err.Error())
		}

		args := make(map[string][]byte)
		args["blockHeader"] = blockBuf
		// set preExe parameter
		moduleName := cmd.Cfg.ContractConfig.ModuleName
		contractName := cmd.Cfg.ContractConfig.ContractName
		methodName := cmd.Cfg.ContractConfig.AnchorMethod
		tx, err := cmd.PreExe(moduleName, contractName, methodName, args)
		// 如果之前已经调用过initAnchorBlockHeader，则跳过
		if err != nil && strings.Contains(err.Error(), "only once") {
			tmpDeliverMeta := &relayerpb.DeliverMeta{
				LastDeliverBlockHeight: cmd.AnchorBlockHeight,
			}
			updateErr := cmd.UpdateDeliverMeta(tmpDeliverMeta)
			if updateErr != nil {
				panic("update deliver meta failed, err:" + updateErr.Error())
			}
			cmd.meta = &relayerpb.DeliverMeta{
				LastDeliverBlockHeight: cmd.AnchorBlockHeight,
			}
			return err
		} else if err != nil {
			fmt.Println("[deliver] preExe for synchronzing anchor block failed, err:", err)
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		txid, err := cmd.SendTx(tx)
		if err != nil && strings.Contains(err.Error(), "only once") {
			tmpDeliverMeta := &relayerpb.DeliverMeta{
				LastDeliverBlockHeight: cmd.AnchorBlockHeight,
			}
			updateErr := cmd.UpdateDeliverMeta(tmpDeliverMeta)
			if updateErr != nil {
				panic("update deliver meta failed, err:" + updateErr.Error())
			}
			cmd.meta = &relayerpb.DeliverMeta{
				LastDeliverBlockHeight: cmd.AnchorBlockHeight,
			}
			return err
		} else if err != nil {
			fmt.Println("[deliver] preExe for synchronzing anchor block failed, err:", err)
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		fmt.Println("[deliver] txid:", txid)
		// 更新DeliverMeta
		tmpDeliverMeta := &relayerpb.DeliverMeta{
			LastDeliverBlockHeight: cmd.AnchorBlockHeight,
		}
		updateErr := cmd.UpdateDeliverMeta(tmpDeliverMeta)
		if updateErr != nil {
			panic("update deliver meta failed, err:" + updateErr.Error())
		}
		cmd.meta = &relayerpb.DeliverMeta{
			LastDeliverBlockHeight: cmd.AnchorBlockHeight,
		}

		break
	}

	return nil
}

func (cmd *DeliverBlockCommand) DeliverBlockHeader() error {
	// 持续获取本地区块头数据并转发给目标链的区块头合约
	currHeight := cmd.meta.GetLastDeliverBlockHeight() + 1
	for {
		if currHeight+3 >= cmd.QueryMeta.GetLastQueryBlockHeight() {
			// 需要等待几个区块再Deliver
			fmt.Println("[deliver] should wait for seconds to deliver again", "currHeight:", currHeight, " deliverBlockHeight:", cmd.QueryMeta.GetLastQueryBlockHeight())
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		// 按照高度获取本地区块头
		blockBuf, getErr := cmd.Storage.GetBlockHeaderByHeight(currHeight)
		// 本地没有该区块头, 稍等一会再重试
		if common.NormalizedKVError(getErr) == common.ErrKVNotFound {
			fmt.Println("[deliver] block is not found in the storage, err:", getErr)
			time.Sleep(time.Duration(1) * time.Second)
			continue
		} else if getErr != nil {
			panic("gey block failed, err:" + getErr.Error())
		}
		// prepare para for pre-exe
		args := make(map[string][]byte)
		args["blockHeader"] = blockBuf
		// set preExe parameter
		moduleName := cmd.Cfg.ContractConfig.ModuleName
		contractName := cmd.Cfg.ContractConfig.ContractName
		methodName := cmd.Cfg.ContractConfig.UpdateMethod
		tx, err := cmd.PreExe(moduleName, contractName, methodName, args)
		if err != nil {
			if strings.Contains(err.Error(), "missing preHash") {
				currHeight--
				continue
			}
			if strings.Contains(err.Error(), "existed already") {
				currHeight++
				continue
			}
		}
		// postTx
		txid, sendErr := cmd.SendTx(tx)
		if sendErr != nil {
			fmt.Println("[deliver] send tx failed, err:", sendErr)
			if strings.Contains(sendErr.Error(), "missing preHash") {
				currHeight--
				continue
			}
		}
		fmt.Println("[deliver] txid:", txid)
		// 更新deliverMeta
		tmpDeliverMeta := &relayerpb.DeliverMeta{
			LastDeliverBlockHeight: currHeight,
		}
		if err = cmd.UpdateDeliverMeta(tmpDeliverMeta); err != nil {
			panic("update deliver meta failed, err:" + err.Error())
		}
		currHeight++
	}

	return nil
}

func (cmd *DeliverBlockCommand) LoadDeliverMeta() (*relayerpb.DeliverMeta, error) {
	return cmd.Storage.LoadDeliverMeta()
}

func (cmd *DeliverBlockCommand) UpdateDeliverMeta(meta *relayerpb.DeliverMeta) error {
	return cmd.Storage.UpdateDeliverMeta(meta)
}
