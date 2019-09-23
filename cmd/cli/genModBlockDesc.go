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

	"github.com/xuperchain/xuperunion/contract"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/crypto/hash"
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
	digestHash := hash.DoubleSha256([]byte(txid))
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

func (g *GenModBlockDescCommand) addFlags() {
	g.cmd.Flags().StringVar(&g.output, "output", "./modifyBlockChain.desc", "Generate desc for modifing block chain data")
}

func init() {
	AddCommand(NewGenModBlockDescCommand)
}
