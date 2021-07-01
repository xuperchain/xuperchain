package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
)

// AccountCommand account cmd entrance
type ParachainCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewAccountCommand new account cmd
func NewParachainCommand(cli *Cli) *cobra.Command {
	c := new(AccountCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "parachain",
		Short: "parachain module",
	}
	c.cmd.AddCommand(NewParachainInvokeCommand(cli))
	return c.cmd
}

func init() {
	AddCommand(NewParachainCommand)
}

// 本文件封装了和平行链有关的client调用接口, 具体格式为:
// xchain-cli parachain invoke 当前parachain kernel调用
//   --method 标识平行链方法，即调用的目标kernerl方法
//   --desc 标识输入参数，json格式

type ParachainInvokeCommand struct {
	cli *Cli
	cmd *cobra.Command

	module   string
	chain    string
	bucket   string
	method   string
	descfile string
	account  string
	fee      string
}

// NewConsensusCommand new consensus cmd
func NewParachainInvokeCommand(cli *Cli) *cobra.Command {
	c := new(ParachainInvokeCommand)
	c.cli = cli
	c.module = ModuleName
	c.cmd = &cobra.Command{
		Use:   "invoke",
		Short: "invoke parachain method",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.invoke(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *ParachainInvokeCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.descfile, "desc", "", "The json config file for parachain.")
	c.cmd.Flags().StringVarP(&c.method, "method", "", "", "kernel method name")
	c.cmd.Flags().StringVarP(&c.account, "account", "", "", "account name")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "fee of one tx")
}

func (c *ParachainInvokeCommand) invoke(ctx context.Context) error {
	ct := &CommTrans{
		Version:      utxo.TxVersion,
		Amount:       "0",
		From:         c.account,
		ModuleName:   c.module,
		ContractName: "$parachain",
		MethodName:   c.method,
		Args:         make(map[string][]byte),
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
		Fee:          c.fee,
	}
	args := map[string]interface{}{}
	if c.descfile == "" {
		return fmt.Errorf("parachain needs desc file.\n")
	}
	desc, err := ioutil.ReadFile(c.descfile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(desc, &args)
	if err != nil {
		return err
	}
	ct.Args, err = convertToXuper3Args(args)
	if err != nil {
		return err
	}
	ct.To, err = readAddress(ct.Keys)
	if err != nil {
		return err
	}
	if c.account == "" {
		initAk, _ := readAddress(c.cli.RootOptions.Keys)
		c.account = initAk
	}
	// 若为getGroup仅走合约预执行
	if c.method == "getGroup" {
		_, _, err := ct.GenPreExeRes(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	return ct.Transfer(ctx)
}
