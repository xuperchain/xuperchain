package main

import (
	"fmt"
	"log"

	"github.com/xuperchain/xuperchain/cmd/xchain/cmd"

	"github.com/spf13/cobra"
)

var (
	Version   = ""
	BuildTime = ""
	CommitID  = ""
)

func main() {
	rootCmd, err := NewServiceCommand()
	if err != nil {
		log.Fatalf("start service failed.err:%v", err)
	}

	if err = rootCmd.Execute(); err != nil {
		log.Fatalf("start service failed.err:%v", err)
	}
}

func NewServiceCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:           "xchain <command> [arguments]",
		Short:         "xchain is a blockchain network building service.",
		Long:          "xchain is a blockchain network building service.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example:       "xchain startup --conf /home/rd/xchain/conf/env.yaml",
	}

	// cmd service
	rootCmd.AddCommand(cmd.GetStartupCmd().GetCmd())
	// cmd version
	rootCmd.AddCommand(GetVersionCmd().GetCmd())
	// cmd createchain
	rootCmd.AddCommand(cmd.GetCreateChainCommand().GetCmd())
	// cmd ledgerPrune
	rootCmd.AddCommand(cmd.GetPruneLedgerCommand().GetCmd())
	// cmd offlineQuery
	rootCmd.AddCommand(cmd.GetOfflineQueryCommand().GetCmd())

	return rootCmd, nil
}

type versionCmd struct {
	cmd.BaseCmd
}

func GetVersionCmd() *versionCmd {
	versionCmdIns := new(versionCmd)

	subCmd := &cobra.Command{
		Use:     "version",
		Short:   "View process version information.",
		Example: "xchain version",
		Run: func(cmd *cobra.Command, args []string) {
			versionCmdIns.PrintVersion()
		},
	}
	versionCmdIns.SetCmd(subCmd)

	return versionCmdIns
}

func (t *versionCmd) PrintVersion() {
	fmt.Printf("%s-%s %s\n", Version, CommitID, BuildTime)
}
