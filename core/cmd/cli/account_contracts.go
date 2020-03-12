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
	address     string
	verbose     bool
}

// NewAccountContractsCommand new account contracts cmd
func NewAccountContractsCommand(cli *Cli) *cobra.Command {
	t := new(AccountContractsCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "contracts",
		Short: "query address/account's contracts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.queryContracts(ctx)
		},
	}
	t.addFlags()
	return t.cmd
}

func (c *AccountContractsCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.accountName, "account", "", "Account name to query contracts.")
	c.cmd.Flags().StringVar(&c.address, "address", "", "address to query contracts.")
	c.cmd.Flags().BoolVar(&c.verbose, "verbose", false, "verbose info will include detailed contract info for address query.")
}

func (c *AccountContractsCommand) queryContracts(ctx context.Context) error {
	if c.accountName != "" {
		err := c.queryAccountContracts(ctx)
		if err != nil {
			fmt.Println("query failed, err=", err.Error())
		}
	} else if c.address != "" {
		err := c.queryAddressContracts(ctx)
		if err != nil {
			fmt.Println("query failed, err=", err.Error())
		}
	} else {
		fmt.Println("this query must use '--address' or '--account' option")
	}
	return nil
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

func (c *AccountContractsCommand) queryAddressContracts(ctx context.Context) error {
	client := c.cli.XchainClient()
	req := &pb.AddressContractsRequest{
		Bcname:      c.cli.RootOptions.Name,
		Address:     c.address,
		NeedContent: c.verbose,
	}

	res, err := client.GetAddressContracts(ctx, req)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(res.GetContracts(), "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(output))
	return nil
}
