package cmd

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/xuperchain/xuperunion/contractsdk/xc/internal/mkfile"
)

const (
	projectFile = "WORK_ROOT"
)

var (
	defaultCxxFlags = []string{
		"-std=c++11",
		"-Os",
		"-I/usr/local/include",
		"-Isrc",
		"-Werror=vla",
	}
	defaultLDFlags = []string{
		"-Oz",
		"-s TOTAL_STACK=256KB",
		"-s TOTAL_MEMORY=1MB",
		"-s DETERMINISTIC=1",
		"-s EXTRA_EXPORTED_RUNTIME_METHODS=[\"stackAlloc\"]",
		"-L/usr/local/lib",
		"-lprotobuf-lite",
		"-lpthread",
	}
)

type buildCommand struct {
	cxxFlags []string
	ldflags  []string
	project  *mkfile.Project

	genCompileCommand bool
	makeFileOnly      bool
	cleanBuild        bool
}

func newBuildCommand() *cobra.Command {
	c := &buildCommand{}
	cmd := &cobra.Command{
		Use:   "build",
		Short: "build command builds a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if c.cleanBuild {
				return c.clean()
			}
			return c.build()
		},
	}
	cmd.Flags().BoolVarP(&c.makeFileOnly, "makefile", "m", false, "generate makefile and exits")
	cmd.Flags().BoolVarP(&c.cleanBuild, "clean", "c", false, "clean stage directory")
	cmd.Flags().BoolVarP(&c.genCompileCommand, "compile_command", "p", false, "generate compile_commands.json for IDE")
	return cmd
}

func (c *buildCommand) initProject(root string) error {
	err := os.Chdir(root)
	if err != nil {
		return err
	}
	abspath, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	appname := filepath.Base(abspath) + ".wasm"

	p := mkfile.NewProject().
		WithRoot(root).
		WithCxxFlags(c.cxxFlags).
		WithLDFlags(c.ldflags)

	dirs, err := ioutil.ReadDir("src")
	if err != nil {
		log.Fatal(err)
	}
	for _, dir := range dirs {
		lib := &mkfile.Library{
			Name: dir.Name(),
			Dir:  filepath.Join("src", dir.Name()),
		}
		p.AddLibrary(lib)
	}
	p.AddApplication(&mkfile.Application{Name: appname, LinkAll: true})
	c.project = p
	return nil
}

func (c *buildCommand) findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if wd == "/" {
			return "", errors.New("can't find " + projectFile)
		}
		xcfile := filepath.Join(wd, projectFile)
		if _, err := os.Stat(xcfile); err == nil {
			return wd, nil
		}
		wd = filepath.Dir(wd)
	}
	return os.Getwd()
}

func (c *buildCommand) xchainRoot() (string, error) {
	xroot := os.Getenv("XROOT")
	if xroot == "" {
		return "", errors.New("missing XROOT env")
	}
	if !filepath.IsAbs(xroot) {
		return "", errors.New("XROOT must be abspath")
	}
	return xroot, nil
}

func (c *buildCommand) initCompileFlags(xroot string) error {
	c.cxxFlags = append([]string{}, defaultCxxFlags...)
	c.addCxxFlags("-I" + filepath.Join(xroot, "contractsdk", "cpp"))
	// -lxchain must be in front of -lprotobuf-lite
	xchainLDFlags := []string{"-L" + filepath.Join(xroot, "contractsdk", "cpp", "build"), "-lxchain"}
	c.ldflags = append(xchainLDFlags, defaultLDFlags...)
	exportJsPath := filepath.Join(xroot, "contractsdk", "cpp", "xchain", "exports.js")
	c.ldflags = append(c.ldflags, "--js-library "+exportJsPath)
	return nil
}

func (c *buildCommand) addCxxFlags(flags ...string) {
	c.cxxFlags = append(c.cxxFlags, flags...)
}

func (c *buildCommand) clean() error {
	root, err := c.findProjectRoot()
	if err != nil {
		return err
	}
	stageDir := filepath.Join(root, "build")
	return os.RemoveAll(stageDir)
}

func (c *buildCommand) build() error {
	root, err := c.findProjectRoot()
	if err != nil {
		return err
	}

	xroot, err := c.xchainRoot()
	if err != nil {
		return err
	}
	err = c.initCompileFlags(xroot)
	if err != nil {
		return err
	}

	err = c.initProject(root)
	if err != nil {
		return err
	}

	if c.makeFileOnly {
		return c.project.GenerateMakeFile(os.Stdout)
	}

	if c.genCompileCommand {
		cfile, err := os.Create("compile_commands.json")
		if err != nil {
			return err
		}
		c.project.GenerateCompileCommands(cfile)
		cfile.Close()
	}

	makefile, err := os.Create(".Makefile")
	if err != nil {
		return err
	}
	err = c.project.GenerateMakeFile(makefile)
	if err != nil {
		makefile.Close()
		return err
	}
	makefile.Close()

	cmd := exec.Command("docker",
		"run",
		"-u", strconv.Itoa(os.Getuid()),
		"--rm",
		"-v", root+":/src",
		"-v", xroot+":"+xroot,
		"hub.baidubce.com/xchain/emcc",
		"emmake", "make", "-f", ".Makefile",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func uniq(list []string) []string {
	var result []string
	m := make(map[string]bool)
	for _, str := range list {
		if !m[str] {
			result = append(result, str)
			m[str] = true
		}
	}
	sort.Strings(result)
	return result
}

func init() {
	addCommand(newBuildCommand)
}
