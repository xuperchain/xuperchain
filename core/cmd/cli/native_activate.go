/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo"
)

// NativeActivateCommand native activate cmd
type NativeActivateCommand struct {
	cli *Cli
	cmd *cobra.Command

	version              string
	minVotePercent       float64
	stopVoteHeightOffset int64
	stopVoteHeight       int64
	triggerHeight        int64
}

// NewNativeActivateCommand new native activate cmd
func NewNativeActivateCommand(cli *Cli) *cobra.Command {
	c := new(NativeActivateCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "activate",
		Short: "[Deprecated] Activate a native contract and make the contract ready.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			pluginName := args[0]
			return c.run(ctx, pluginName, "activate")
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *NativeActivateCommand) addFlags() {
	c.cmd.Flags().Float64Var(&c.minVotePercent, "vote-percent", 51, "proposal min vote percent")
	c.cmd.Flags().Int64Var(&c.stopVoteHeightOffset, "vote-height-offset", 20, "proposal stop vote height offset from current height")
	c.cmd.MarkFlagRequired("vote-height-offset")
	c.cmd.Flags().Int64Var(&c.stopVoteHeight, "vote-height", 0, "proposal stop vote height")
	c.cmd.Flags().Int64Var(&c.triggerHeight, "trigger-height", 0, "proposal trigger height if zero will use vote-height+1")

	//设置为必填选项
	c.cmd.Flags().StringVar(&c.version, "version", "", "specific version to be activated")
	c.cmd.MarkFlagRequired("version")
}

func getCurrentHeight(ctx context.Context, bcname string, client pb.XchainClient) (int64, error) {
	req := &pb.CommonIn{
		Header: global.GHeader(),
	}
	reply, err := client.GetSystemStatus(ctx, req)
	if err != nil {
		return 0, err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return 0, errors.New(reply.Header.Error.String())
	}
	for _, chain := range reply.GetSystemsStatus().GetBcsStatus() {
		if chain.GetBcname() == bcname {
			return chain.GetMeta().GetTrunkHeight(), nil
		}
	}

	return 0, fmt.Errorf("block chain not found:%s", bcname)
}

func (c *NativeActivateCommand) run(ctx context.Context, pluginName string, action string) error {
	client := c.cli.XchainClient()
	if c.stopVoteHeight == 0 {
		currentHeight, err := getCurrentHeight(ctx, c.cli.RootOptions.Name, client)
		if err != nil {
			return err
		}
		c.stopVoteHeight = currentHeight + c.stopVoteHeightOffset
	}
	if c.triggerHeight == 0 {
		c.triggerHeight = c.stopVoteHeight + 1
	}
	contractDesc := ContractDesc{
		Module: "proposal",
		Method: "Propose",
		Args: map[string]interface{}{
			"min_vote_percent": c.minVotePercent,
			"stop_vote_height": c.stopVoteHeight,
		},
		Trigger: TriggerDesc{
			Module: "native",
			Method: action,
			Args: map[string]interface{}{
				"pluginName": pluginName,
				"version":    c.version,
			},
			Height: c.triggerHeight,
		},
	}
	myaddress, err := readAddress(c.cli.RootOptions.Keys)
	if err != nil {
		return err
	}
	desc, _ := json.Marshal(contractDesc)
	opt := TransferOptions{
		BlockchainName: c.cli.RootOptions.Name,
		KeyPath:        c.cli.RootOptions.Keys,
		CryptoType:     c.cli.RootOptions.CryptoType,
		To:             myaddress,
		Amount:         "0",
		Desc:           desc,
		Version:        utxo.TxVersion,
	}
	txid, err := c.cli.Transfer(ctx, &opt)
	if err != nil {
		return err
	}
	fmt.Printf("txid:%s\n", txid)
	fmt.Printf("stop-vote-height:%d\n", c.stopVoteHeight)
	fmt.Printf("trigger-vote-height:%d\n", c.triggerHeight)
	return nil
}
