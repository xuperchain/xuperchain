package main

import "os"
import "fmt"
import "strings"
import "github.com/xuperchain/xuperchain/core/ledger"
import "github.com/xuperchain/xuperchain/core/crypto/client"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("./dump_chain /home/xxxx/blochain_work_space")
		os.Exit(1)
	}
	dataPathOthers := []string{}
	if len(os.Args) > 2 {
		dataPathOthers = strings.Split(os.Args[2], ",")
	}
	workspace := os.Args[1]
	lg, err := ledger.NewLedger(workspace, nil, dataPathOthers, "default", client.CryptoTypeDefault)
	if err != nil {
		fmt.Println(err, workspace, dataPathOthers)
	}
	blocks, bErr := lg.Dump()
	if bErr != nil {
		fmt.Println(bErr)
	}
	for height, blks := range blocks {
		fmt.Println("Height: ", height)
		for _, blkInfo := range blks {
			fmt.Printf("  |--%s\n", blkInfo)
		}
	}
}
