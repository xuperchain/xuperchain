package cmd

import (
	"github.com/spf13/cobra"
)

type BaseCmd struct {
	// cobra command
	Cmd *cobra.Command
}

func (t *BaseCmd) SetCmd(cmd *cobra.Command) {
	t.Cmd = cmd
}

func (t *BaseCmd) GetCmd() *cobra.Command {
	return t.Cmd
}
