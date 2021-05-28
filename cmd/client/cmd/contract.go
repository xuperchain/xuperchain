package cmd

import (
	"github.com/spf13/cobra"
)

// ContractCommand contract cmd
type ContractCommand struct {
}

// NewContractCommand new contract cmd
func NewContractCommand(cli *Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "Operate contract command, query",
	}
	cmd.AddCommand(NewContractStatDataQueryCommand(cli))
	return cmd
}

func init() {
	AddCommand(NewContractCommand)
}
