/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import "github.com/spf13/cobra"

// ACLCommand acl cmd struct
type ACLCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewACLCommand new acl cmd
func NewACLCommand(cli *Cli) *cobra.Command {
	c := new(ACLCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "acl",
		Short: "Operate an access control list(ACL): query.",
	}
	c.cmd.AddCommand(NewACLQueryCommand(cli))
	return c.cmd
}

func init() {
	AddCommand(NewACLCommand)
}
