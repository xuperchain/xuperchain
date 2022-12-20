/*
 * Copyright (c) 2021. Baidu Inc. All Rights Reserved.
 */

package cmd

import "github.com/spf13/cobra"

// NewTDposCommand new tdpos cmd
func NewTDposCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tdpos",
		Short: "Operate a command with tdpos, query-candidates|query-checkResult|query-nominate-records|query-nominee-record|query-vote-records|query-voted-records|status",
	}
	cmd.AddCommand(NewQueryCandidatesCommand(cli))
	cmd.AddCommand(NewQueryCheckResultCommand(cli))
	cmd.AddCommand(NewQueryNominateRecordsCommand(cli))
	cmd.AddCommand(NewQueryNomineeRecordsCommand(cli))
	cmd.AddCommand(NewQueryVoteRecordsCommand(cli))
	cmd.AddCommand(NewQueryVotedRecordsCommand(cli))
	cmd.AddCommand(NewQueryStatusCommand(cli))
	return cmd
}

//func init() {
//	AddCommand(NewTDposCommand)
//}
