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
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/xuperchain/xuperchain/core/contract"
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/global"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/utxo/txhash"
)

// GenModBlockDescCommand modify blockchain data desc file only by regulatory address
type GenModBlockDescCommand struct {
	cli    *Cli
	cmd    *cobra.Command
	output string
}

// NewGenModBlockDescCommand new modify block cmd
func NewGenModBlockDescCommand(cli *Cli) *cobra.Command {
	g := new(GenModBlockDescCommand)
	g.cli = cli
	g.cmd = &cobra.Command{
		Use:   "genModDesc [OPTIONS] txid",
		Short: "Generate modified blockchain data desc: [OPTIONS].",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("expect txid to be modified")
			}
			ctx := context.TODO()
			return g.genDesc(ctx, args[0])
		},
	}
	g.addFlags()
	return g.cmd
}

func (g *GenModBlockDescCommand) genDesc(ctx context.Context, txid string) error {
	// create crypto client
	cryptoClient, err := crypto_client.CreateCryptoClient(g.cli.RootOptions.CryptoType)
	if err != nil {
		return errors.New("Create crypto client error")
	}
	fromScrkey, err := readPrivateKey(g.cli.RootOptions.Keys)
	if err != nil {
		return err
	}
	privateKey, err := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(fromScrkey))
	if err != nil {
		return err
	}

	tx, err := g.queryTx(ctx, txid)
	if err != nil {
		return err
	}
	tx.Desc = []byte("")
	tx.TxOutputsExt = []*pb.TxOutputExt{}
	digestHash, err := txhash.MakeTxDigestHash(tx)
	if err != nil {
		return err
	}
	sign, sErr := cryptoClient.SignECDSA(privateKey, digestHash)
	if sErr != nil {
		return sErr
	}
	signstr := hex.EncodeToString(sign)

	fromPubkey, err := readPublicKey(g.cli.RootOptions.Keys)
	if err != nil {
		return err
	}

	txDesc := &contract.TxDesc{
		Module: "kernel",
		Method: "UpdateBlockChainData",
		Args:   make(map[string]interface{}),
	}
	txDesc.Args["txid"] = txid
	txDesc.Args["publicKey"] = fromPubkey
	txDesc.Args["sign"] = signstr

	signJSON, err := json.MarshalIndent(txDesc, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(signJSON))

	err = ioutil.WriteFile(g.output, signJSON, 0755)
	if err != nil {
		return errors.New("WriteFile error")
	}

	return nil
}

func (g *GenModBlockDescCommand) queryTx(ctx context.Context, txid string) (*pb.Transaction, error) {
	client := g.cli.XchainClient()
	rawTxid, err := hex.DecodeString(txid)
	if err != nil {
		return nil, fmt.Errorf("bad txid:%s", txid)
	}
	txstatus := &pb.TxStatus{
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Bcname: g.cli.RootOptions.Name,
		Txid:   rawTxid,
	}
	reply, err := client.QueryTx(ctx, txstatus)
	if err != nil {
		return nil, err
	}
	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return nil, errors.New(reply.Header.Error.String())
	}
	if reply.Tx == nil {
		return nil, errors.New("tx not found")
	}
	if reply.Status != pb.TransactionStatus_CONFIRM {
		return nil, errors.New("tx is not in block")
	}
	return reply.Tx, nil
}

func (g *GenModBlockDescCommand) addFlags() {
	g.cmd.Flags().StringVar(&g.output, "output", "./modifyBlockChain.desc", "Generate desc for modifing block chain data")
}

func init() {
	AddCommand(NewGenModBlockDescCommand)
}
