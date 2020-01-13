/*
 * Copyright (c) 2019, Baidu.com, Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/pb"
)

// AccountBalanceCommand account balance command
type AccountBalanceCommand struct {
	cli    *Cli
	cmd    *cobra.Command
	frozen bool
}

// NewAccountBalanceCommand new function
func NewAccountBalanceCommand(cli *Cli) *cobra.Command {
	b := new(AccountBalanceCommand)
	b.cli = cli
	b.cmd = &cobra.Command{
		Use:   "balance [account/address]",
		Short: "Query the balance of an account or address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			var account string
			var err error

			if len(args) == 0 {
				account, err = readAddress(cli.RootOptions.Keys)
				if err != nil {
					return fmt.Errorf("not provide account but read address from file error:%s", err)
				}
			} else {
				account = args[0]
			}

			return b.queryBalance(ctx, account)
		},
	}

	b.addFlags()

	return b.cmd
}

func (b *AccountBalanceCommand) addFlags() {
	b.cmd.Flags().BoolVarP(&b.frozen, "frozen", "Z", false, "Get frozen balance.")
}

func (b *AccountBalanceCommand) queryBalance(ctx context.Context, account string) error {
	client := b.cli.XchainClient()
	addrstatus := &pb.AddressStatus{
		Address: account,
		Bcs: []*pb.TokenDetail{
			{Bcname: b.cli.RootOptions.Name},
		},
	}

	fGetBalance := client.GetBalance
	if b.frozen {
		fGetBalance = client.GetFrozenBalance
	}

	reply, err := fGetBalance(ctx, addrstatus)
	if err != nil {
		return err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}
	fmt.Println(reply.Bcs[0].Balance)
	return nil
}
