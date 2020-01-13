/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

// BlockCommand query block
type BlockCommand struct {
	cli      *Cli
	cmd      *cobra.Command
	byHeight bool
}

// NewBlockCommand new block cmd
func NewBlockCommand(cli *Cli) *cobra.Command {
	b := new(BlockCommand)
	b.cli = cli
	b.cmd = &cobra.Command{
		Use:   "block [OPTIONS] blockid or height",
		Short: "Operate a block: [OPTIONS].",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("expect blockid")
			}
			ctx := context.TODO()
			if b.byHeight {
				height, err := strconv.Atoi(args[0])
				if err != nil {
					return err
				}
				return b.queryBlockByHeight(ctx, int64(height))
			}
			return b.queryBlock(ctx, args[0])
		},
	}
	b.addFlags()
	return b.cmd
}

func (b *BlockCommand) queryBlock(ctx context.Context, blockid string) error {
	client := b.cli.XchainClient()
	rawBlockid, err := hex.DecodeString(blockid)
	if err != nil {
		return err
	}
	blockIDPB := &pb.BlockID{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname:      b.cli.RootOptions.Name,
		Blockid:     rawBlockid,
		NeedContent: true,
	}
	block, err := client.GetBlock(ctx, blockIDPB)
	if err != nil {
		return err
	}
	if block.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(block.Header.Error.String())
	}
	if block.Block == nil {
		return errors.New("block not found")
	}
	iblock := FromInternalBlockPB(block.Block)
	output, err := json.MarshalIndent(iblock, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

func (b *BlockCommand) queryBlockByHeight(ctx context.Context, height int64) error {
	client := b.cli.XchainClient()
	blockHeightPB := &pb.BlockHeight{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname: b.cli.RootOptions.Name,
		Height: height,
	}
	block, err := client.GetBlockByHeight(ctx, blockHeightPB)
	if err != nil {
		return err
	}
	if block.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(block.Header.Error.String())
	}
	if block.Block == nil {
		return errors.New("block not found")
	}
	iblock := FromInternalBlockPB(block.Block)
	output, err := json.MarshalIndent(iblock, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}

func (b *BlockCommand) addFlags() {
	b.cmd.Flags().BoolVarP(&b.byHeight, "byHeight", "N", false, "Get block by height.")
}

func init() {
	AddCommand(NewBlockCommand)
}
