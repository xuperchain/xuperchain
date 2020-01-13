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

// ACLQueryCommand query acl struct
type ACLQueryCommand struct {
	cli          *Cli
	cmd          *cobra.Command
	accountName  string
	contractName string
	methodName   string
}

// NewACLQueryCommand new acl query cmd
func NewACLQueryCommand(cli *Cli) *cobra.Command {
	t := new(ACLQueryCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "query [OPTIONS] account contract method",
		Short: "query an access control list(ACL) for an account or contract method.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.queryACL(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (t *ACLQueryCommand) addFlags() {
	t.cmd.Flags().StringVar(&t.accountName, "account", "", "contract account name")
	t.cmd.Flags().StringVar(&t.contractName, "contract", "", "contract name")
	t.cmd.Flags().StringVar(&t.methodName, "method", "", "method name")
}

func (t *ACLQueryCommand) queryACL(ctx context.Context) error {
	client := t.cli.XchainClient()
	aclStatus := &pb.AclStatus{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname:       t.cli.RootOptions.Name,
		AccountName:  t.accountName,
		ContractName: t.contractName,
		MethodName:   t.methodName,
	}
	if len(t.accountName) == 0 && len(t.contractName) == 0 {
		return errors.New("param error")
	}
	reply, err := client.QueryACL(ctx, aclStatus)
	if err != nil {
		return err
	}

	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return errors.New(reply.Header.Error.String())
	}

	if reply != nil {
		acl := reply.GetAcl()
		output, err := json.MarshalIndent(acl, "", "  ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(output))
		if string(output) != "{}" && output != nil {
			confirmed := reply.GetConfirmed()
			if !confirmed {
				fmt.Println("unconfirmed")
			} else {
				fmt.Println("confirmed")
			}
		}
	}
	return nil
}
