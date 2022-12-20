package cmd

import "github.com/spf13/cobra"

// NewUtxoCommand init child utxo command
func NewUtxoCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "utxo",
		Short: "Operate an utxo: list|merge|split.",
	}
	cmd.AddCommand(NewListUtxoCommand(cli))
	cmd.AddCommand(NewMergeUtxoCommand(cli))
	cmd.AddCommand(NewSplitUtxoCommand(cli))

	return cmd
}

func init() {
	AddCommand(NewUtxoCommand)
}
