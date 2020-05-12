package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperchain/core/cmd/xdev/internal/jstest"
	"github.com/xuperchain/xuperchain/core/cmd/xdev/internal/jstest/xchain"
)

type testCommand struct {
	cmd       *cobra.Command
	quiet     bool
	runPatten string
}

func newTestCommand() *cobra.Command {
	c := &testCommand{}
	c.cmd = &cobra.Command{
		Use:   "test [contract.test.js]",
		Short: "test perform unit test",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.test(args)
			if err != nil {
				return err
			}
			return nil
		},
	}
	c.addFlags()
	return c.cmd
}

func (c *testCommand) addFlags() {
	c.cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "quiet test output")
	c.cmd.Flags().StringVarP(&c.runPatten, "run", "r", "", "Run only those tests matching the regular expression.")
}

func (c *testCommand) test(args []string) error {
	if len(args) == 0 {
		return c.testPackage()
	}
	wd := filepath.Dir(args[0])
	return c.testFiles(wd, args)
}

func (c *testCommand) testFiles(wd string, files []string) error {
	runner, err := jstest.NewRunner(&jstest.RunOption{
		Quiet:  c.quiet,
		Patten: c.runPatten,
	}, xchain.NewAdapter())

	if err != nil {
		return err
	}

	runner.AddModulePath([]string{filepath.Join(wd, "node_modules")})
	for _, testFile := range files {
		err = runner.AddTestFile(testFile)
		if err != nil {
			return err
		}
	}
	return runner.Run(wd)
}

const (
	testDir = "test"
)

func (c *testCommand) testPackage() error {
	root, err := findPackageRoot()
	if err != nil {
		return err
	}
	err = os.Chdir(root)
	if err != nil {
		return err
	}

	packageTestDir := filepath.Join(root, testDir)
	testFiles, err := filepath.Glob(filepath.Join(packageTestDir, "*.test.js"))
	if err != nil {
		return err
	}

	return c.testFiles(packageTestDir, testFiles)
}

func init() {
	addCommand(newTestCommand)
}
