package relayer

import (
	"fmt"
	"time"

	"google.golang.org/grpc"

	relayerpb "github.com/xuperchain/xuperchain/core/cmd/relayer/pb"
	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/pb"
)

// QueryMetaRegister deliver routine needs to get lastest query block height
// deliver should be lower compared to query routine
type QueryMetaRegister interface {
	GetLastQueryBlockHeight() int64
}

// DeliverBlockCommand parameter of DeliverBlockCommand to be required
// client: in order to communicate with dst chain
// Cfg: config parameter for Deliver Routine
// meta: record the lastest block delived
// AnchorBlockHeight: initial block header to synchronize
// Storage: data source(block header)
// QueryMeta: get latest query block height
type DeliverBlockCommand struct {
	client            pb.XchainClient
	Cfg               ChainConfig
	meta              *relayerpb.DeliverMeta
	AnchorBlockHeight int64
	Storage           *Storage
	QueryMeta         QueryMetaRegister
}

// InitXchainClient initialize client con between relayer and dst chain
// Set MaxMsgSize as 32MB
func (cmd *DeliverBlockCommand) InitXchainClient() error {
	conn, err := grpc.Dial(cmd.Cfg.RPCAddr, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	cmd.client = pb.NewXchainClient(conn)
	return nil
}

// Deliver unified deliver method including DeliverAnchorBlockHeader and DeliverBlockHeader
// In general, check if it's necessary to launch a call of DeliverAnchorBlockHeader
// Depending on the existence of DeliverMeta
// Then, call DeliverBlockHeader to deliver block header except for AnchorBlockHeader
// Totally, AnchorBlockHeader first, and Other Block Headers later.
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

// DeliverAnchorBlockHeader deliver anchor block header
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
		if NormalizedBlockHeaderTxError(err) == ErrBlockHeaderTxOnlyOnce {
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
		if NormalizedBlockHeaderTxError(err) == ErrBlockHeaderTxOnlyOnce {
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

// DeliverBlockHeader deliver block headers except for anchor block header
func (cmd *DeliverBlockCommand) DeliverBlockHeader() error {
	// 持续获取本地区块头数据并转发给目标链的区块头合约
	currHeight := cmd.meta.GetLastDeliverBlockHeight() + 1
	for {
		if currHeight > cmd.QueryMeta.GetLastQueryBlockHeight() {
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
		if NormalizedBlockHeaderTxError(err) == ErrBlockHeaderTxMissingPreHash {
			currHeight--
			continue
		}
		if NormalizedBlockHeaderTxError(err) == ErrBlockHeaderTxExist {
			currHeight++
			continue
		}
		if err != nil {
			fmt.Println("[deliver] preExe failed, err:", err)
			panic("preExe failed, err:" + err.Error())
		}
		// postTx
		txid, sendErr := cmd.SendTx(tx)
		if NormalizedBlockHeaderTxError(sendErr) == ErrBlockHeaderTxMissingPreHash {
			fmt.Println("[deliver] send tx failed, err:", sendErr)
			currHeight--
			continue
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

// LoadDeliverMeta load deliver meta
func (cmd *DeliverBlockCommand) LoadDeliverMeta() (*relayerpb.DeliverMeta, error) {
	return cmd.Storage.LoadDeliverMeta()
}

// UpdateDeliverMeta update deliver meta
func (cmd *DeliverBlockCommand) UpdateDeliverMeta(meta *relayerpb.DeliverMeta) error {
	return cmd.Storage.UpdateDeliverMeta(meta)
}
