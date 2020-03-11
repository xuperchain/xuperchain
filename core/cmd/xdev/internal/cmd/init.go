package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/core/cmd/xdev/internal/mkfile"
)

var descTpl = `[package]
name = "main"
`

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

var testTpl = `
var assert = require("assert");

Test("hello", function (t) {
    var contract;
    t.Run("deploy", function (tt) {
        contract = xchain.Deploy({
            name: "hello",
            code: "../hello.wasm",
            lang: "c",
            init_args: {}
        })
    });

    t.Run("invoke", function (tt) {
        resp = contract.Invoke("hello", {});
        assert.equal(resp.Body, "hello world");
    })
})
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
	pkgfile := mkfile.PkgDescFile
	if _, err := os.Stat(pkgfile); err == nil {
		return errors.New("project already initialized")
	}
	err := ioutil.WriteFile(pkgfile, []byte(descTpl), 0644)
	if err != nil {
		return err
	}
	maindir := filepath.Join("src")
	err = os.MkdirAll(maindir, 0755)
	if err != nil {
		return err
	}
	mainfile := filepath.Join(maindir, "main.cc")
	err = ioutil.WriteFile(mainfile, []byte(codeTpl), 0644)
	if err != nil {
		return err
	}

	testdir := filepath.Join("test")
	err = os.MkdirAll(testdir, 0755)
	if err != nil {
		return err
	}
	testfile := filepath.Join(testdir, "hello.test.js")
	err = ioutil.WriteFile(testfile, []byte(testTpl), 0644)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	addCommand(newInitCommand)
}
