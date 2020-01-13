/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

// MultisigSignCommand multisig sign struct
type MultisigSignCommand struct {
	cli *Cli
	cmd *cobra.Command

	tx       string
	output   string
	signType string
}

// NewMultisigSignCommand multisig sign init method
func NewMultisigSignCommand(cli *Cli) *cobra.Command {
	c := new(MultisigSignCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "sign",
		Short: "Generate a signature for the raw transaction with private key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sign()
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *MultisigSignCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.tx, "tx", "./tx.out", "Raw serialized transaction data file")
	c.cmd.Flags().StringVar(&c.signType, "signtype", "", "type of signature, support multi/ring")
	c.cmd.Flags().StringVar(&c.output, "output", "./sign.out", "Generate signature file for a transaction.")
}

// sign 命令的主入口
func (c *MultisigSignCommand) sign() error {
	data, err := ioutil.ReadFile(c.tx)
	if err != nil {
		return err
	}
	tx := &pb.Transaction{}
	err = proto.Unmarshal(data, tx)
	if err != nil {
		return err
	}

	fromPubkey, err := readPublicKey(c.cli.RootOptions.Keys)
	if err != nil {
		return err
	}

	if c.signType == "multi" {
		signData, err := ioutil.ReadFile(c.tx + ".ext")
		if err != nil {
			return err
		}
		msd := &MultisigData{}
		err = json.Unmarshal(signData, msd)
		if err != nil {
			return err
		}
		fromScrkey, err := readPrivateKey(c.cli.RootOptions.Keys)
		if err != nil {
			return err
		}

		xcc, err := crypto_client.CreateCryptoClientFromJSONPrivateKey([]byte(fromScrkey))
		if err != nil {
			return err
		}
		priv, err := xcc.GetEcdsaPrivateKeyFromJSON([]byte(fromScrkey))
		if err != nil {
			return err
		}
		digestHash, err := txhash.MakeTxDigestHash(tx)
		if err != nil {
			return err
		}
		// TODO get partial sign
		ki, idx, err := c.findKfromKlist(msd, []byte(fromPubkey))
		if err != nil {
			return err
		}
		si := xcc.GetSiUsingKCRM(priv, ki, msd.C, msd.R, digestHash)
		psd := &PartialSign{
			Si:    si,
			Index: idx,
		}
		jsonContent, err := json.Marshal(psd)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(c.output, jsonContent, 0755)
		if err != nil {
			return errors.New("WriteFile error")
		}
		fmt.Println(string(jsonContent))
	} else if c.signType != "" {
		return fmt.Errorf("SignType[%s] is not supported", c.signType)
	} else {
		signTx, err := c.genSignTx(tx)
		if err != nil {
			return errors.New("Sign tx error")
		}

		err = c.genSignFile(fromPubkey, signTx)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSignTx use privatekey to get sign
func (c *MultisigSignCommand) genSignTx(tx *pb.Transaction) ([]byte, error) {
	// create crypto client
	cryptoClient, err := crypto_client.CreateCryptoClient(c.cli.RootOptions.CryptoType)
	if err != nil {
		return nil, errors.New("Create crypto client error")
	}
	fromScrkey, err := readPrivateKey(c.cli.RootOptions.Keys)
	if err != nil {
		return nil, err
	}

	signTx, err := txhash.ProcessSignTx(cryptoClient, tx, []byte(fromScrkey))
	if err != nil {
		return nil, err
	}

	return signTx, nil
}

// genSignFile output to file
func (c *MultisigSignCommand) genSignFile(pubkey string, sign []byte) error {
	signInfo := &pb.SignatureInfo{
		PublicKey: pubkey,
		Sign:      sign,
	}

	signJSON, err := json.MarshalIndent(signInfo, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(signJSON))

	err = ioutil.WriteFile(c.output, signJSON, 0755)
	if err != nil {
		return errors.New("WriteFile error")
	}

	return nil
}

func (c *MultisigSignCommand) findKfromKlist(msd *MultisigData, pubJSON []byte) ([]byte, int, error) {
	xcc, err := crypto_client.CreateCryptoClientFromJSONPublicKey(pubJSON)
	if err != nil {
		return nil, 0, err
	}
	pubkey, err := xcc.GetEcdsaPublicKeyFromJSON(pubJSON)
	if err != nil {
		return nil, 0, err
	}
	addr, err := xcc.GetAddressFromPublicKey(pubkey)
	if err != nil {
		return nil, 0, err
	}
	for idx, ki := range msd.KList {
		tmpkey, err := xcc.GetEcdsaPublicKeyFromJSON(msd.PubKeys[idx])
		if err != nil {
			continue
		}
		tmpaddr, err := xcc.GetAddressFromPublicKey(tmpkey)
		if err != nil {
			continue
		}
		if addr == tmpaddr {
			return ki, idx, nil
		}
	}
	return nil, 0, fmt.Errorf("Public key not found in multisig data")
}
