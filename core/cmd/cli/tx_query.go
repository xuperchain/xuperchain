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

	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
)

// TxQueryCommand tx query cmd
type TxQueryCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewTxQueryCommand new tx query cmd
func NewTxQueryCommand(cli *Cli) *cobra.Command {
	t := new(TxQueryCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "query txid",
		Short: "query transaction based on txid",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("expect txid")
			}
			ctx := context.TODO()
			return t.queryTx(ctx, args[0])
		},
	}
	t.addFlags()
	return t.cmd
}

func (t *TxQueryCommand) addFlags() {
}

func (t *TxQueryCommand) queryTx(ctx context.Context, txid string) error {
	client := t.cli.XchainClient()
	rawTxid, err := hex.DecodeString(txid)
	if err != nil {
		return fmt.Errorf("bad txid:%s", txid)
	}
	txstatus := &pb.TxStatus{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname: t.cli.RootOptions.Name,
		Txid:   rawTxid,
	}
	reply, err := client.QueryTx(ctx, txstatus)
	if err != nil {
		return err
	}

	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}
	if reply.Tx == nil {
		return errors.New("tx not found")
	}
	tx := FromPBTx(reply.Tx)
	output, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}
