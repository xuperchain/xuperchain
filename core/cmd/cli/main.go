/*
 * Copyright (c) 2019, Baidu.com, Inc. All Rights Reserved.
 */

package main

import "log"

var (
	buildVersion = ""
	buildDate    = ""
	commitHash   = ""
)

func main() {
	cli := NewCli()
	err := cli.Init()
	if err != nil {
		log.Fatal(err)
	}
	cli.AddCommands(commands)
	cli.Execute()
}
