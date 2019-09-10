package cmd

import (
	"github.com/spf13/cobra"
)

var (
	commandFuncs []func() *cobra.Command

	rootOptions RootOptions
)

type RootOptions struct {
}

func addCommand(cmdFunc func() *cobra.Command) {
	commandFuncs = append(commandFuncs, cmdFunc)
}

func rootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "xc",
		SilenceErrors: false,
		SilenceUsage:  true,
	}
	for _, cmdFunc := range commandFuncs {
		rootCmd.AddCommand(cmdFunc())
	}
	return rootCmd
}

func Main() {
	root := rootCommand()
	root.Execute()

}
