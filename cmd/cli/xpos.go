package main

import (
	"github.com/spf13/cobra"
)

// XPoSCommand cmd for xpos
type XPoSCommand struct {
}

// NewXPoSCommand new xpos cmd
func NewXPoSCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xpos",
		Short: "Operate a command with xpos, lease|cancel|contend|release|slot|xpower",
	}

	cmd.AddCommand(NewXPoSLeaseCommand(cli))
	cmd.AddCommand(NewXPoSCancelLeaseCommand(cli))
	cmd.AddCommand(NewXPoSContendSlotCommand(cli))
	cmd.AddCommand(NewXPoSReleaseSlotCommand(cli))
	cmd.AddCommand(NewXPoSSlotQueryCommand(cli))
	cmd.AddCommand(NewXPoSXPowerQueryCommand(cli))

	return cmd
}

func init() {
	AddCommand(NewXPoSCommand)
}
