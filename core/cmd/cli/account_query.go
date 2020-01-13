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
)

// AccountQueryCommand query acl struct
type AccountQueryCommand struct {
	cli     *Cli
	cmd     *cobra.Command
	address string
}

// NewAccountQueryCommand new account query cmd
func NewAccountQueryCommand(cli *Cli) *cobra.Command {
	t := new(AccountQueryCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "query [OPTIONS] account list contains a specific address",
		Short: "query the account list containing a specific address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.queryAccount(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (t *AccountQueryCommand) addFlags() {
	t.cmd.Flags().StringVar(&t.address, "address", "", "address")
}

func (t *AccountQueryCommand) queryAccount(ctx context.Context) error {
	client := t.cli.XchainClient()
	if t.address == "" {
		address, err := readAddress(t.cli.RootOptions.Keys)
		if err != nil {
			return fmt.Errorf("not provide account but read address from file error:%s", err)
		}
		t.address = address
	}
	accountResponse := &pb.AK2AccountRequest{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname:  t.cli.RootOptions.Name,
		Address: t.address,
	}
	reply, err := client.GetAccountByAK(ctx, accountResponse)
	if err != nil {
		return err
	}

	if reply.GetHeader().GetError() != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}

	if reply != nil {
		account := reply.GetAccount()
		output, err := json.MarshalIndent(account, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
	}
	return nil
}
