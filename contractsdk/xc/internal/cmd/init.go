package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var codeTpl = `#include "xchain/xchain.h"

struct Hello : public xchain::Contract {};

DEFINE_METHOD(Hello, initialize) {
    xchain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(Hello, hello) {
    xchain::Context* ctx = self.context();
    ctx->ok("hello world");
}
`

type initCommand struct {
}

func newInitCommand() *cobra.Command {
	c := &initCommand{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init initializes a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			var root string
			if len(args) == 1 {
				root = args[0]
			}
			return c.init(root)
		},
	}
	return cmd
}

func (c *initCommand) init(root string) error {
	if root != "" {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
		os.Chdir(root)
	}
	xcfile := projectFile
	if _, err := os.Stat(xcfile); err == nil {
		return errors.New("project already initialized")
	}
	err := ioutil.WriteFile(xcfile, nil, 0644)
	if err != nil {
		return err
	}
	maindir := filepath.Join("src", "main")
	err = os.MkdirAll(maindir, 0755)
	if err != nil {
		return err
	}
	mainfile := filepath.Join(maindir, "main.cc")
	err = ioutil.WriteFile(mainfile, []byte(codeTpl), 0644)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	addCommand(newInitCommand)
}
