/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"github.com/spf13/cobra"
	"github.com/xuperchain/xupercore/lib/crypto/client"

	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
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
	c.cmd.Flags().StringVar(&c.signType, "signtype", "", "type of signature, support multi/ring(Note: this is a demo feature, do NOT use it in production environment)")
	c.cmd.Flags().StringVar(&c.output, "output", "./sign.out", "Generate signature file for a transaction.")
}

// sign 命令的主入口
func (c *MultisigSignCommand) sign() error {
	data, err := os.ReadFile(c.tx)
	if err != nil {
		return err
	}
	tx := &pb.Transaction{}
	err = proto.Unmarshal(data, tx)
	if err != nil {
		return err
	}

	ak := newAK(c.cli.RootOptions.Keys)
	from, err := ak.keyPair()
	if err != nil {
		return err
	}

	// invalid sign type
	if c.signType != "" {
		return fmt.Errorf("SignType[%s] is not supported", c.signType)
	}

	// sign type: multi
	if c.signType == "multi" {
		return c.signMulti(from, tx)
	}

	// sign type: default
	return c.signDefault(from, tx)
}

// signMulti signs as type multi
func (c *MultisigSignCommand) signMulti(from KeyPair, tx *pb.Transaction) error {
	fmt.Println("Note: this is a demo feature, do NOT use it in production environment.")
	msd, err := loadMultisig(c.tx)
	if err != nil {
		return err
	}

	crypto, err := client.CreateCryptoClientFromJSONPrivateKey([]byte(from.secretKey))
	if err != nil {
		return err
	}
	ecdsaPrivateKey, err := crypto.GetEcdsaPrivateKeyFromJsonStr(from.secretKey)
	if err != nil {
		return err
	}
	digestHash, err := common.MakeTxDigestHash(tx)
	if err != nil {
		return err
	}
	// TODO get partial sign
	ki, idx, err := c.findKFromKList(msd, []byte(from.publicKey))
	if err != nil {
		return err
	}
	si := crypto.GetSiUsingKCRM(ecdsaPrivateKey, ki, msd.C, msd.R, digestHash)
	psd := &PartialSign{
		Si:    si,
		Index: idx,
	}
	jsonContent, err := json.Marshal(psd)
	if err != nil {
		return err
	}
	err = os.WriteFile(c.output, jsonContent, 0755)
	if err != nil {
		return errors.New("WriteFile error")
	}
	fmt.Println(string(jsonContent))
	return nil
}

// signDefault signs as type default
func (c *MultisigSignCommand) signDefault(from KeyPair, tx *pb.Transaction) error {
	crypto, err := client.CreateCryptoClient(c.cli.RootOptions.Crypto)
	if err != nil {
		return err
	}
	signInfo, err := from.SignTx(tx, crypto)
	if err != nil {
		return err
	}
	return c.genSignFile(signInfo)
}

// genSignFile output to file
func (c *MultisigSignCommand) genSignFile(signInfo *pb.SignatureInfo) error {
	signJSON, err := json.MarshalIndent(signInfo, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(signJSON))

	err = os.WriteFile(c.output, signJSON, 0755)
	if err != nil {
		return errors.New("WriteFile error")
	}

	return nil
}

func (c *MultisigSignCommand) findKFromKList(msd *MultisigData, pubJSON []byte) ([]byte, int, error) {
	xcc, err := client.CreateCryptoClientFromJSONPublicKey(pubJSON)
	if err != nil {
		return nil, 0, err
	}
	pubkey, err := xcc.GetEcdsaPublicKeyFromJsonStr(string(pubJSON))
	if err != nil {
		return nil, 0, err
	}
	addr, err := xcc.GetAddressFromPublicKey(pubkey)
	if err != nil {
		return nil, 0, err
	}
	for idx, ki := range msd.KList {
		tmpkey, err := xcc.GetEcdsaPublicKeyFromJsonStr(string(msd.PubKeys[idx]))
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
