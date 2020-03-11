package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	commandFuncs []func() *cobra.Command

	rootOptions RootOptions
	version     string
)

type RootOptions struct {
}

func addCommand(cmdFunc func() *cobra.Command) {
	commandFuncs = append(commandFuncs, cmdFunc)
}

func rootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "xdev",
		SilenceErrors: false,
		SilenceUsage:  true,
		Version:       version,
	}
	for _, cmdFunc := range commandFuncs {
		rootCmd.AddCommand(cmdFunc())
	}
	return rootCmd
}

func SetVersion(ver, date, commit string) {
	version = fmt.Sprintf("%s-%s %s", ver, commit, date)
}

func Main() {
	root := rootCommand()
	root.Execute()

}
