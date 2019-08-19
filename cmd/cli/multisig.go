/*
 * Copyright (c) 2019. Baidu Inc. All Rights Reserved.
 */

package main

import "github.com/spf13/cobra"

// MultisigCommand Multisig set command
type MultisigCommand struct {
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
