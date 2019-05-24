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

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
)

// BlockCommand query block
type BlockCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewBlockCommand new block cmd
func NewBlockCommand(cli *Cli) *cobra.Command {
	b := new(BlockCommand)
	b.cli = cli
	b.cmd = &cobra.Command{
		Use:   "block [OPTIONS] blockid",
		Short: "Operate a block: [OPTIONS].",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("expect blockid")
			}
			ctx := context.TODO()
			return b.queryBlock(ctx, args[0])
		},
	}
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

func init() {
	AddCommand(NewBlockCommand)
}
