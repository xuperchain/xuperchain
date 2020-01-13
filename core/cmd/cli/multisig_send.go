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
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"

	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/global"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/utxo/txhash"
)

// MultisigSendCommand multisig send struct
type MultisigSendCommand struct {
	cli *Cli
	cmd *cobra.Command

	tx       string
	signType string
}

// NewMultisigSendCommand multisig gen init method
func NewMultisigSendCommand(cli *Cli) *cobra.Command {
	c := new(MultisigSendCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "send",
		Short: "Post a raw transaction along with multi-signatures.",
		Long: `./xchain-cli multisig --tx ./tx.out arg1 [arg2] --signtype [multi/ring]
If signtype is empty:
	arg1: Initiator signature array, separated with commas; 
	arg2: AuthRequire signature array, separated with commas.
If signtype is "multi":
    arg1: The signature array, separated with commas.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.TODO()
			if c.signType == "multi" {
				return c.sendXuper(ctx, args[0])
			} else if c.signType != "" {
				return fmt.Errorf("SignType[%s] is not supported", c.signType)
			}
			if len(args) < 2 {
				return fmt.Errorf("Args error, need at least two arguments but got %d", len(args))
			}
			return c.send(ctx, args[0], args[1])
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *MultisigSendCommand) addFlags() {
	c.cmd.Flags().StringVar(&c.tx, "tx", "./tx.out", "Serialized transaction data file")
	c.cmd.Flags().StringVar(&c.signType, "signtype", "", "type of signature, support multi/ring")
}

// send 命令的主入口
func (c *MultisigSendCommand) send(ctx context.Context, initPath string, authPath string) error {
	data, err := ioutil.ReadFile(c.tx)
	if err != nil {
		return errors.New("Fail to open serialized transaction data file")
	}
	tx := &pb.Transaction{}
	err = proto.Unmarshal(data, tx)
	if err != nil {
		return errors.New("Fail to Unmarshal proto")
	}

	signs, err := c.getSigns(initPath)
	if err != nil {
		return err
	}
	tx.InitiatorSigns = signs

	signAuths, err := c.getSigns(authPath)
	if err != nil {
		return err
	}
	tx.AuthRequireSigns = signAuths

	tx.Txid, err = txhash.MakeTransactionID(tx)
	if err != nil {
		return errors.New("MakeTxDigesthash txid error")
	}

	txid, err := c.sendTx(ctx, tx)
	if err != nil {
		return err
	}
	fmt.Printf("Tx id: %s\n", txid)

	return nil
}

// sendXuper process XuperSign
func (c *MultisigSendCommand) sendXuper(ctx context.Context, signs string) error {
	data, err := ioutil.ReadFile(c.tx)
	if err != nil {
		return errors.New("Fail to open serialized transaction data file")
	}
	tx := &pb.Transaction{}
	err = proto.Unmarshal(data, tx)
	if err != nil {
		return errors.New("Fail to Unmarshal proto")
	}

	signData, err := ioutil.ReadFile(c.tx + ".ext")
	if err != nil {
		return err
	}
	msd := &MultisigData{}
	err = json.Unmarshal(signData, msd)
	if err != nil {
		return fmt.Errorf("Unmarshal MultisigData failed, err=%v", err)
	}
	needLen := len(msd.KList)
	if needLen <= 1 {
		return fmt.Errorf("multisig need at least two parties, but got %d", needLen)
	}
	slist := make([][]byte, needLen)
	signSlice := strings.Split(signs, ",")
	if len(signSlice) != needLen {
		return fmt.Errorf("sign file is not equal to multisig public keys, need[%d] but got[%d]",
			needLen, len(signSlice))
	}
	for _, signfile := range signSlice {
		sign, err := ioutil.ReadFile(signfile)
		if err != nil {
			return errors.New("Failed to open sign file")
		}
		psi := &PartialSign{}
		err = json.Unmarshal([]byte(sign), psi)
		if err != nil {
			return fmt.Errorf("Unmarshal PartialSign failed, err=%v", err)
		}
		if psi.Index > needLen-1 || psi.Index < 0 {
			return fmt.Errorf("partial signature data is invalid")
		}
		slist[psi.Index] = psi.Si
	}
	xcc, err := crypto_client.CreateCryptoClientFromJSONPublicKey(msd.PubKeys[0])
	if err != nil {
		return fmt.Errorf("create crypto client failed, err=%v", err)
	}
	s := xcc.GetSUsingAllSi(slist)
	finalsign, err := xcc.GenerateMultiSignSignature(s, msd.R)
	if err != nil {
		return fmt.Errorf("GenerateMultiSignSignature failed, err=%v", err)
	}
	tx.XuperSign = &pb.XuperSignature{
		PublicKeys: msd.PubKeys,
		Signature:  finalsign,
	}

	tx.Txid, err = txhash.MakeTransactionID(tx)
	if err != nil {
		return errors.New("MakeTxDigesthash txid error")
	}

	txid, err := c.sendTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("sendTx failed, err=%v", err)
	}
	fmt.Printf("Tx id: %s\n", txid)

	return nil
}

// getSigns 读文件，填充pb.SignatureInfo
func (c *MultisigSendCommand) getSigns(path string) ([]*pb.SignatureInfo, error) {
	signs := []*pb.SignatureInfo{}
	for _, file := range strings.Split(path, ",") {
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, errors.New("Failed to open sign file")
		}

		sign := &pb.SignatureInfo{}
		err = json.Unmarshal(buf, sign)
		if err != nil {
			return nil, errors.New("Failed to json unmarshal sign file")
		}

		signs = append(signs, sign)
	}

	return signs, nil
}

func (c *MultisigSendCommand) sendTx(ctx context.Context, tx *pb.Transaction) (string, error) {
	txStatus := &pb.TxStatus{
		Bcname: c.cli.RootOptions.Name,
		Status: pb.TransactionStatus_UNCONFIRM,
		Tx:     tx,
		Header: &pb.Header{
			Logid: global.Glogid(),
		},
		Txid: tx.Txid,
	}

	//reply, err := c.cli.XchainClient().Send(ctx, txStatus)
	reply, err := c.cli.XchainClient().PostTx(ctx, txStatus)
	if err != nil {
		return "", err
	}

	if reply.Header.Error != pb.XChainErrorEnum_SUCCESS {
		return "", fmt.Errorf("Failed to post tx:%s, logid:%s", reply.Header.Error.String(), reply.Header.Logid)
	}

	return hex.EncodeToString(txStatus.Txid), nil
}
