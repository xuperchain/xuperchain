package relayer

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	relayerpb "github.com/xuperchain/xuperchain/core/relayer/pb"
)

type QueryBlockCommand struct {
	client  pb.XchainClient
	Cfg     ChainConfig
	Storage *Storage
}

func (cmd *QueryBlockCommand) InitXchainClient() error {
	conn, err := grpc.Dial(cmd.Cfg.RPCAddr, grpc.WithInsecure(), grpc.WithMaxMsgSize(64<<20-1))
	if err != nil {
		return err
	}
	cmd.client = pb.NewXchainClient(conn)
	return nil
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

func (cmd *QueryBlockCommand) LoadQueryMeta() (*relayerpb.QueryMeta, error) {
	return cmd.Storage.LoadQueryMeta()
}

func (cmd *QueryBlockCommand) UpdateQueryMeta(meta *relayerpb.QueryMeta) error {
	return cmd.Storage.UpdateQueryMeta(meta)
}
