/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import "github.com/spf13/cobra"

// NewACLCommand new acl cmd
func NewACLCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acl",
		Short: "Operate an access control list(ACL): query.",
	}
	cmd.AddCommand(NewACLQueryCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewACLCommand)
}
