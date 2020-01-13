package main

import "github.com/xuperchain/xuperchain/core/contractsdk/xc/internal/cmd"

var (
	buildVersion = ""
	buildDate    = ""
	commitHash   = ""
)

func main() {
	cmd.SetVersion(buildVersion, buildDate, commitHash)
	cmd.Main()
}
