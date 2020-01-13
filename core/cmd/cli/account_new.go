/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/permission/acl/utils"
	"github.com/xuperchain/xuperchain/core/utxo"
)

// AccountNewCommand new account struct
type AccountNewCommand struct {
	cli *Cli
	cmd *cobra.Command

	accountName string
	descfile    string
	fee         string
}

// NewAccountNewCommand new account new cmd
func NewAccountNewCommand(cli *Cli) *cobra.Command {
	t := new(AccountNewCommand)
	t.cli = cli
	t.cmd = &cobra.Command{
		Use:   "new ",
		Short: "Create an account.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return t.newAccount(ctx)
		},
	}
	t.addFlags()

	return t.cmd
}

func (c *AccountNewCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.accountName, "account", "", "Account name for contracts.")
	c.cmd.Flags().StringVar(&c.descfile, "desc", "", "The json config file for creating an account.")
	c.cmd.Flags().StringVar(&c.fee, "fee", "0", "The fee to create an account.")
}

func (c *AccountNewCommand) newAccount(ctx context.Context) error {
	if len(c.descfile) == 0 && len(c.accountName) == 0 {
		return errors.New("account name or desc file required")
	}

	ct := &CommTrans{
		Amount:       "0",
		Fee:          c.fee,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,

		MethodName: "NewAccount",
		Args:       make(map[string][]byte),

		Descfile: c.descfile,
		IsQuick:  false,

		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.CryptoType,
	}

	var err error
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}

	if c.accountName != "" {
		ct.ModuleName = "xkernel"
		ct.Args["account_name"] = []byte(c.accountName)
		simpleACL := `
        {
            "pm": {
                "rule": 1,
                "acceptValue": 1.0
            },
            "aksWeight": {
                "` + ct.To + `": 1.0
            }
        }
        `
		ct.Args["acl"] = []byte(simpleACL)
	}

	err = ct.Transfer(ctx)
	if err != nil {
		return err
	}

	err = c.printRealAccountName(c.accountName)
	if err != nil {
		return err
	}

	return nil
}

func (c *AccountNewCommand) printRealAccountName(name string) error {
	if name != "" {
		fmt.Printf("account name: %s\n", utils.MakeAccountKey(c.cli.RootOptions.Name, name))
		return nil
	}

	desc, err := ioutil.ReadFile(c.descfile)
	if err != nil {
		return err
	}

	preExeParam, err := c.readPreExeParamWithDesc(desc)
	if err != nil {
		return err
	}

	accountName := string(preExeParam.Args["account_name"])
	fmt.Printf("account name: %s\n", utils.MakeAccountKey(c.cli.RootOptions.Name, accountName))
	return nil
}

func (c *AccountNewCommand) readPreExeParamWithDesc(buf []byte) (*pb.InvokeRequest, error) {
	params := new(invokeRequestWraper)
	err := json.Unmarshal(buf, params)
	if err != nil {
		return nil, fmt.Errorf("unmarshal json error:%s", err)
	}
	params.InvokeRequest.Args = make(map[string][]byte)
	for k, v := range params.Args {
		params.InvokeRequest.Args[k] = []byte(v)
	}
	return &params.InvokeRequest, nil
}
