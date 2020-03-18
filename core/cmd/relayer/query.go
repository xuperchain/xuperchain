package relayer

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"

	relayerpb "github.com/xuperchain/xuperchain/core/cmd/relayer/pb"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

// QueryBlockCommand parameter of QueryBlockCommand to be required
// client: a tool to communicate with source chain
// Cfg: chain config for realyer
// Storage: a place to store block data received from source chain
type QueryBlockCommand struct {
	client  pb.XchainClient
	Cfg     ChainConfig
	Storage *Storage
}

// InitXchainClient initialize the communication client
// Set MaxMsgSize as 32MB
func (cmd *QueryBlockCommand) InitXchainClient() error {
	conn, err := grpc.Dial(cmd.Cfg.RPCAddr, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	cmd.client = pb.NewXchainClient(conn)
	return nil
}

func (cmd *QueryBlockCommand) GetLatestBlockHeightFromSrcChain() (int64, error) {
	fmt.Println("[query] prepare to query latest block height from source chain")
	bcStatusPB := &pb.BCStatus{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname: cmd.Cfg.Bcname,
	}
	bcStatus, err := cmd.client.GetBlockChainStatus(context.TODO(), bcStatusPB)
	if err != nil {
		return 0, err
	}
	if bcStatus.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return 0, errors.New(bcStatus.Header.Error.String())
	}
	if bcStatus.GetMeta() == nil {
		return 0, errors.New("can't get ledger meta")
	}
	fmt.Println("[query] the latest block height in source chain is ", bcStatus.GetMeta().GetTrunkHeight())
	return bcStatus.GetMeta().GetTrunkHeight(), nil
}

// 从原链获取指定高度区块
func (cmd *QueryBlockCommand) QueryBlockByHeight(height int64) (*pb.InternalBlock, error) {
	fmt.Println("[query] prepare to query block from source chain")
	blockHeightPB := &pb.BlockHeight{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname: cmd.Cfg.Bcname,
		Height: height,
	}
	block, err := cmd.client.GetBlockByHeight(context.TODO(), blockHeightPB)
	if err != nil {
		return nil, err
	}
	if block.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, errors.New(block.Header.Error.String())
	}
	if block.Block == nil {
		return nil, errors.New("block not found")
	}
	fmt.Println("[query] query block by height success")
	return block.Block, nil
}

func (cmd *QueryBlockCommand) FetchBlockFromSrcChain(height int64) (*pb.InternalBlock, error) {
	block, err := cmd.QueryBlockByHeight(height)
	if err != nil {
		fmt.Println("[query] QueryBlockByHeight error:", err)
		return nil, err
	}

	return block, nil
}

// LoadQueryMeta load query meta
func (cmd *QueryBlockCommand) LoadQueryMeta() (*relayerpb.QueryMeta, error) {
	return cmd.Storage.LoadQueryMeta()
}

// UpdateQueryMeta update query meta
func (cmd *QueryBlockCommand) UpdateQueryMeta(meta *relayerpb.QueryMeta) error {
	return cmd.Storage.UpdateQueryMeta(meta)
}
