/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

// MultisigData generated multisig data
type MultisigData struct {
	R       []byte   // common random value
	C       []byte   // common public key
	KList   [][]byte // random value list
	PubKeys [][]byte // all public keys
}

// PartialSign partial sign is single Si sign for multisig
type PartialSign struct {
	Si    []byte
	Index int
}

// NewMultisigCommand MultisigCommand init method
func NewMultisigCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisig",
		Short: "Operate a command with multisign: check|gen|send|sign|get.",
	}
	cmd.AddCommand(NewMultisigGenCommand(cli))
	cmd.AddCommand(NewGetComplianceCheckSignCommand(cli))
	cmd.AddCommand(NewMultisigCheckCommand(cli))
	cmd.AddCommand(NewMultisigSignCommand(cli))
	cmd.AddCommand(NewMultisigSendCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewMultisigCommand)
}

// loadMultisig loads multisig data from file
// Params:
// 	txFile: tx file path
func loadMultisig(txFile string) (*MultisigData, error) {
	signData, err := ioutil.ReadFile(txFile + ".ext")
	if err != nil {
		return nil, err
	}

	msd := &MultisigData{}
	err = json.Unmarshal(signData, msd)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal MultisigData failed, err=%v", err)
	}
	return msd, err
}
