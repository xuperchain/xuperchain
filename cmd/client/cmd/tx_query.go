/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"github.com/spf13/cobra"
	"github.com/xuperchain/xupercore/lib/utils"

	"github.com/xuperchain/xuperchain/service/pb"
)

// TxQueryCommand tx query cmd
type TxQueryCommand struct {
	cli *Cli
	cmd *cobra.Command

	pbfile string
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
	t.cmd.Flags().StringVarP(&t.pbfile, "pb", "p", "", "generate pb file")
}

func (t *TxQueryCommand) queryTx(ctx context.Context, txid string) error {
	client := t.cli.XchainClient()
	rawTxid, err := hex.DecodeString(txid)
	if err != nil {
		return fmt.Errorf("bad txid:%s", txid)
	}
	txstatus := &pb.TxStatus{
		Header: &pb.Header{
			Logid: utils.GenLogId(),
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

	if t.pbfile != "" {
		buf, _ := proto.Marshal(reply.Tx)
		err = ioutil.WriteFile(t.pbfile, buf, 0644)
		if err != nil {
			return err
		}
	}
	output, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(output))
	return nil
}
