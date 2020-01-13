/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/core/pb"
)

// AccountContractsCommand
type AccountContractsCommand struct {
	cli         *Cli
	cmd         *cobra.Command
	accountName string
}

// NewAccountContractsCommand new account contracts cmd
func NewAccountContractsCommand(cli *Cli) *cobra.Command {
	t := new(AccountContractsCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "contracts",
		Short: "query account's contracts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.queryAccountContracts(ctx)
		},
	}
	t.addFlags()
	return t.cmd
}

func (c *AccountContractsCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.accountName, "account", "", "Account name to query contracts.")
}

func (c *AccountContractsCommand) queryAccountContracts(ctx context.Context) error {
	client := c.cli.XchainClient()
	req := &pb.GetAccountContractsRequest{
		Bcname:  c.cli.RootOptions.Name,
		Account: c.accountName,
	}

	res, err := client.GetAccountContracts(ctx, req)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(res.GetContractsStatus(), "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}
