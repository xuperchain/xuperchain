package cmd

import "github.com/spf13/cobra"

// UtxoCommand utxo cmd entry
type UtxoCommand struct {
	cli *Cli
	cmd *cobra.Command
}

// NewUtxoCommand init child utxo command
func NewUtxoCommand(cli *Cli) *cobra.Command {
	c := new(UtxoCommand)
	c.cli = cli
	c.cmd = &cobra.Command{
		Use:   "utxo",
		Short: "Operate an utxo: list|merge|split.",
	}
	c.cmd.AddCommand(NewListUtxoCommand(cli))
	c.cmd.AddCommand(NewMergeUtxoCommand(cli))
	c.cmd.AddCommand(NewSplitUtxoCommand(cli))

	return c.cmd
}

func init() {
	AddCommand(NewUtxoCommand)
}
