// Copyright (c) 2021. Baidu Inc. All Rights Reserved.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xupercore/bcs/ledger/xledger/state/utxo"
	"github.com/xuperchain/xupercore/lib/utils"

	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
	aclUtils "github.com/xuperchain/xupercore/kernel/permission/acl/utils"
)

// SplitUtxoCommand split utxo of ak or account
type SplitUtxoCommand struct {
	cli *Cli
	cmd *cobra.Command
	// account will be split
	account string
	num     int64
	// while splitting an account, it can not be null
	accountPath string
	isGenRawTx  bool
	multiAddrs  string
	output      string
}

// NewSplitUtxoCommand return
func NewSplitUtxoCommand(cli *Cli) *cobra.Command {
	c := new(SplitUtxoCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "split ",
		Short: "Split the utxo of an account or address.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.splitUtxo(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *SplitUtxoCommand) addFlags() {
	c.cmd.Flags().StringVarP(&c.account, "account", "A", "", "The account/address to be split (default ./data/keys/address).")
	c.cmd.Flags().Int64VarP(&c.num, "num", "N", 1, "The number to split.")
	c.cmd.Flags().StringVarP(&c.accountPath, "accountPath", "P", "", "The account path, which is required for an account.")
	c.cmd.Flags().BoolVarP(&c.isGenRawTx, "raw", "m", false, "Is only generate raw tx output.")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "Serialized transaction data file.")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "M", "data/acl/addrs", "MultiAddrs to fill required accounts/addresses.")
}

func (c *SplitUtxoCommand) splitUtxo(_ context.Context) error {
	if c.num <= 0 {
		return errors.New("illegal split utxo num, num > 0 required")
	}
	if aclUtils.IsAK(c.account) && c.accountPath == "" {
		return errors.New("accountPath can not be null because account is an Account name")
	}

	initAk, err := readAddress(c.cli.RootOptions.Keys)
	if c.account == "" {
		c.account = initAk
	}

	if aclUtils.IsAccount(c.account) && c.account != initAk {
		return errors.New("parse account error")
	}

	tx := &pb.Transaction{
		Version:   utxo.TxVersion,
		Coinbase:  false,
		Nonce:     utils.GenNonce(),
		Timestamp: time.Now().UnixNano(),
		Initiator: initAk,
	}

	amount, err := c.getBalanceHelper()
	if err != nil {
		return err
	}
	ct := &CommTrans{
		Amount:       amount,
		FrozenHeight: 0,
		Version:      utxo.TxVersion,
		From:         c.account,
		Args:         make(map[string][]byte),
		IsQuick:      false,
		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
		CryptoType:   c.cli.RootOptions.Crypto,
		MultiAddrs:   c.multiAddrs,
		Output:       c.output,
	}

	totalNeed, ok := big.NewInt(0).SetString(amount, 10)
	if !ok {
		return errors.New("get totalNeed error")
	}

	txInputs, txOutput, err := ct.GenTxInputs(context.Background(), totalNeed)
	tx.TxInputs = txInputs

	txOutputs, err := c.genSplitOutputs(totalNeed)
	if err != nil {
		return err
	}
	if txOutput != nil {
		txOutputs = append(txOutputs, txOutput)
	}
	tx.TxOutputs = txOutputs

	if c.isGenRawTx {
		// 填充需要多重签名的addr
		multiAddrs, err := ct.GenAuthRequire(c.multiAddrs)
		if err != nil {
			return err
		}
		tx.AuthRequire = multiAddrs
	} else {
		tx.AuthRequire, err = genAuthRequire(c.account, c.accountPath)
		if err != nil {
			return errors.New("genAuthRequire error")
		}
	}

	// preExec
	preExeRPCReq := &pb.InvokeRPCRequest{
		Bcname:   c.cli.RootOptions.Name,
		Requests: []*pb.InvokeRequest{},
		Header: &pb.Header{
			Logid: utils.GenLogId(),
		},
		Initiator:   initAk,
		AuthRequire: tx.AuthRequire,
	}
	preExeRes, err := ct.XchainClient.PreExec(context.Background(), preExeRPCReq)
	if err != nil {
		return err
	}
	tx.ContractRequests = preExeRes.GetResponse().GetRequests()
	tx.TxInputsExt = preExeRes.GetResponse().GetInputs()
	tx.TxOutputsExt = preExeRes.GetResponse().GetOutputs()
	if c.isGenRawTx {
		// 直接输出原始交易内容到文件
		return ct.GenTxFile(tx)
	}

	tx.InitiatorSigns, err = ct.genInitSign(tx)
	if err != nil {
		return err
	}
	tx.AuthRequireSigns, err = ct.genAuthRequireSignsFromPath(tx, c.accountPath)
	if err != nil {
		return err
	}

	// calculate tx ID
	tx.Txid, err = common.MakeTxId(tx)
	if err != nil {
		return err
	}
	txID, err := ct.postTx(context.Background(), tx)
	fmt.Println(txID)
	return err
}

func (c *SplitUtxoCommand) getBalanceHelper() (string, error) {
	as := &pb.AddressStatus{}
	as.Address = c.account
	var tokens []*pb.TokenDetail
	token := pb.TokenDetail{Bcname: c.cli.RootOptions.Name}
	tokens = append(tokens, &token)
	as.Bcs = tokens
	r, err := c.cli.XchainClient().GetBalance(context.Background(), as)
	if err != nil {
		return "0", err
	}
	return r.Bcs[0].Balance, nil
}

func (c *SplitUtxoCommand) genSplitOutputs(totalNeed *big.Int) ([]*pb.TxOutput, error) {
	txOutputs := []*pb.TxOutput{}
	amount := big.NewInt(0)
	rest := totalNeed
	if big.NewInt(c.num).Cmp(rest) == 1 {
		return nil, errors.New("illegal split utxo, split utxo <= BALANCE required")
	}
	amount.Div(rest, big.NewInt(c.num))
	output := pb.TxOutput{}
	output.Amount = amount.Bytes()
	output.ToAddr = []byte(c.account)
	for i := int64(1); i < c.num && rest.Cmp(amount) == 1; i++ {
		tmpOutput := output
		txOutputs = append(txOutputs, &tmpOutput)
		rest.Sub(rest, amount)
	}
	output.Amount = rest.Bytes()
	txOutputs = append(txOutputs, &output)
	return txOutputs, nil
}
