/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/utxo"
)

// MultisigGenCommand multisig generate struct
type MultisigGenCommand struct {
	cli *Cli
	cmd *cobra.Command

	to           string
	amount       string
	descfile     string
	fee          string
	frozenHeight int64
	version      int32
	output       string
	multiAddrs   string
	pubkeys      string
	from         string
	signType     string
	// contract params
	moduleName   string
	contractName string
	methodName   string
	args         string
}

// NewMultisigGenCommand multisig gen init method
func NewMultisigGenCommand(cli *Cli) *cobra.Command {
	c := new(MultisigGenCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "gen",
		Short: "Generate a raw transaction.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			return c.generateTx(ctx)
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *MultisigGenCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.to, "to", "", "Target account/address of transfer.")
	c.cmd.Flags().StringVar(&c.amount, "amount", "0", "Token amount to be transferred.")
	c.cmd.Flags().StringVar(&c.descfile, "desc", "", "Desc file with the format of json for contract.")
	c.cmd.Flags().StringVar(&c.fee, "fee", "", "Fee to run a transaction.")
	c.cmd.Flags().Int64Var(&c.frozenHeight, "frozen", 0, "Frozen height of a transaction.")
	c.cmd.Flags().Int32Var(&c.version, "txversion", utxo.TxVersion, "Tx version.")
	c.cmd.Flags().StringVarP(&c.output, "output", "o", "./tx.out", "Serialized transaction data file.")
	c.cmd.Flags().StringVarP(&c.multiAddrs, "multiAddrs", "A", "data/acl/addrs", "MultiAddrs to fill required accounts/addresses.")
	c.cmd.Flags().StringVarP(&c.pubkeys, "publickeys", "P", "data/acl/pubkeys", "public keys of initiator and auth_require addresses.")
	c.cmd.Flags().StringVar(&c.signType, "signtype", "", "type of signature, support multi/ring")
	c.cmd.Flags().StringVar(&c.from, "from", "", "Initiator of an transaction.")
	c.cmd.Flags().StringVar(&c.moduleName, "module", "", "Contract type: xkernel or wasm or native, native is deprecated.")
	c.cmd.Flags().StringVar(&c.contractName, "contract", "", "Contract name to be called.")
	c.cmd.Flags().StringVar(&c.methodName, "method", "", "Contract method to be called, It has been implemented in the target contract.")
	c.cmd.Flags().StringVar(&c.args, "args", "", "Contract parameters with json format for target contract method.")
}

// gen 命令入口
func (c *MultisigGenCommand) generateTx(ctx context.Context) error {
	var msd *MultisigData
	if c.signType == "multi" {
		// multisig, use XuperSign
		pubkeys, err := c.readPublicKeysFromFile(c.pubkeys)
		if err != nil {
			return err
		}

		msd, err = c.getMultiSignData(pubkeys)
		if err != nil {
			return err
		}
	} else if c.signType != "" {
		fmt.Printf("SignType[%s] is not supported", c.signType)
		return fmt.Errorf("SignType is not supported")
	}

	ct := &CommTrans{
		To:           c.to,
		Amount:       c.amount,
		Descfile:     c.descfile,
		Fee:          c.fee,
		FrozenHeight: c.frozenHeight,
		Version:      c.version,
		From:         c.from,
		ModuleName:   c.moduleName,
		ContractName: c.contractName,
		MethodName:   c.methodName,
		Args:         make(map[string][]byte),
		MultiAddrs:   c.multiAddrs,
		Output:       c.output,
		IsPrint:      true,
		IsQuick:      true,

		ChainName:    c.cli.RootOptions.Name,
		Keys:         c.cli.RootOptions.Keys,
		XchainClient: c.cli.XchainClient(),
	}

	if c.args != "" {
		err := json.Unmarshal([]byte(c.args), &ct.Args)
		if err != nil {
			return err
		}
	}

	if msd != nil {
		jsonData, err := json.Marshal(msd)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(c.output+".ext", jsonData, 0755)
		if err != nil {
			return fmt.Errorf("write file error")
		}
	}

	return ct.GenerateMultisigGenRawTx(ctx)
}

func (c *MultisigGenCommand) readPublicKeysFromFile(filename string) ([][]byte, error) {
	var pubkeys [][]byte
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		pubkeys = append(pubkeys, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Scan file failed, error=%v", err)
		return nil, err
	}
	return pubkeys, nil
}

func (c *MultisigGenCommand) getMultiSignData(pubkeys [][]byte) (*MultisigData, error) {
	var klist [][]byte
	var rlist [][]byte
	var pklist []*ecdsa.PublicKey
	if len(pubkeys) < 2 {
		fmt.Println("the number of public keys for multisig should more than 2")
		return nil, fmt.Errorf("invalid public keys")
	}

	xcc, err := crypto_client.CreateCryptoClientFromJSONPublicKey(pubkeys[0])
	if err != nil {
		return nil, err
	}

	for _, pubkey := range pubkeys {
		// get Ki
		ki, err := xcc.GetRandom32Bytes()
		if err != nil {
			return nil, err
		}
		klist = append(klist, ki)
		pki, err := xcc.GetEcdsaPublicKeyFromJSON(pubkey)
		if err != nil {
			return nil, err
		}
		pklist = append(pklist, pki)
		ri := xcc.GetRiUsingRandomBytes(pki, ki)
		rlist = append(rlist, ri)
	}
	r := xcc.GetRUsingAllRi(pklist[0], rlist)
	spk, err := xcc.GetSharedPublicKeyForPublicKeys(pklist)
	if err != nil {
		return nil, err
	}

	msd := &MultisigData{
		R:       r,
		C:       spk,
		KList:   klist,
		PubKeys: pubkeys,
	}
	return msd, nil
}
