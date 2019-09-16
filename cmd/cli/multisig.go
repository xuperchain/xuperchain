/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import "github.com/spf13/cobra"

// MultisigCommand Multisig set command
type MultisigCommand struct {
}

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
	cmd.AddCommand(NewMultisigCheckCommand(cli))
	cmd.AddCommand(NewMultisigSignCommand(cli))
	cmd.AddCommand(NewMultisigSendCommand(cli))
	cmd.AddCommand(NewGetComplianceCheckSignCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewMultisigCommand)
}
