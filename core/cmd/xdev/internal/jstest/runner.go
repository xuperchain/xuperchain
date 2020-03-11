// Package jstest is a test framework using js as test script
package jstest

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/ddliu/motto"
	"github.com/robertkrimen/otto"
	"github.com/xuperchain/xuperchain/core/cmd/xdev/internal/jstest/builtins"
)

// TestCase is the one test case
type TestCase struct {
	Name string
	F    func(*testing.T)
}

// RunOption is the option to run a test script
type RunOption struct {
	Quiet    bool
	Patten   string
	InGoTest bool
}

// DefaultRunOption returns the default RunOption
func DefaultRunOption() *RunOption {
	return &RunOption{
		Quiet:  false,
		Patten: "",
	}
}

// Runner is the runner of test file
type Runner struct {
	Option RunOption

	vm      *motto.Motto
	global  *otto.Object
	tests   []testing.InternalTest
	adapter Adapter
}

// NewRunner instance a Runner
func NewRunner(opt *RunOption, adapter Adapter) (*Runner, error) {
	if adapter == nil {
		adapter = defaultAdapter{}
	}
	if opt == nil {
		opt = DefaultRunOption()
	}

	vm := motto.New()
	r := &Runner{
		Option:  *opt,
		vm:      vm,
		global:  globalObject(vm.Otto),
		adapter: adapter,
	}

	err := r.init()
	if err != nil {
		return nil, err
	}
	adapter.OnSetup(r)
	return r, nil
}

func (r *Runner) init() error {
	err := r.initJSModules()
	if err != nil {
		return err
	}
	r.registerGlobals()
	return nil
}

func (r *Runner) registerGlobals() {
	r.global.Set("_test", r.add)
	for name, v := range builtins.Globals {
		r.global.Set(name, v)
	}
}

func (r *Runner) initGoTestPackage() error {
	var flags []string
	if !r.Option.Quiet {
		flags = append(flags, "-test.v")
	}
	if r.Option.Patten != "" {
		flags = append(flags, "-test.run", r.Option.Patten)
	}
	flag.CommandLine.Parse(flags)
	return nil
}

func (r *Runner) initJSModules() error {
	// load jstest module
	v, err := r.vm.Require("jstest", ".")
	if err != nil {
		return err
	}
	exports := v.Object()
	// export all symbols from jstest module to global
	for _, name := range exports.Keys() {
		value, _ := exports.Get(name)
		r.global.Set(name, value)
	}
	return nil
}

func (r *Runner) add(name string, body func(t *testing.T)) {
	testcase := r.adapter.OnTestCase(r, TestCase{
		Name: name,
		F:    body,
	})

	r.tests = append(r.tests, testing.InternalTest{
		Name: testcase.Name,
		F:    testcase.F,
	})
}

// AddModulePath add path as nodejs module search path
func (r *Runner) AddModulePath(path []string) {
	r.vm.AddPath(path...)
}

// AddTestFile add a js test file to Runner
func (r *Runner) AddTestFile(file string) error {
	_, err := r.vm.Run(file)
	return err
}

// Run run a js test file using file's dir as working directory
func (r *Runner) RunFile(file string) error {
	err := r.AddTestFile(file)
	if err != nil {
		return err
	}
	rundir := filepath.Dir(file)
	return r.Run(rundir)
}

// Run run all tests with rundir as working directory
func (r *Runner) Run(rundir string) error {
	wd, _ := os.Getwd()
	err := os.Chdir(rundir)
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	if r.Option.InGoTest {
		ok := testing.RunTests(testDeps{}.MatchString, r.tests)
		if !ok {
			return errors.New("")
		}
		return nil
	}

	tmain := testing.MainStart(testDeps{}, r.tests, nil, nil)
	err = r.initGoTestPackage()
	if err != nil {
		return err
	}
	ret := tmain.Run()
	if ret != 0 {
		return errors.New("")
	}
	return nil
}

// VM returns the js vm
func (r *Runner) VM() *otto.Otto {
	return r.vm.Otto
}

// GlobalObject returns the global Object in js vm
func (r *Runner) GlobalObject() *otto.Object {
	return r.global
}

// Close release resources by Runner
func (r *Runner) Close() {
	r.adapter.OnTeardown(r)
}

func globalObject(vm *otto.Otto) *otto.Object {
	return vm.Context().This.Object()
}
