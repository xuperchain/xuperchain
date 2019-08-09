package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	osexec "os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"unsafe"

	"github.com/xuperchain/xuperunion/xvm/compile"
	"github.com/xuperchain/xuperunion/xvm/exec"
)

var (
	wasm2cPath    string = "../compile/wabt/build/wasm2c"
	wast2jsonPath string = "../compile/wabt/build/wast2json"
)

type commandType string

var (
	moduleCommand                    commandType = "module"
	actionCommand                    commandType = "action"
	assertReturnCommand              commandType = "assert_return"
	assertInvalidCommand             commandType = "assert_invalid"
	assertTrapCommand                commandType = "assert_trap"
	assertMalformedCommand           commandType = "assert_malformed"
	assertExhaustionCommand          commandType = "assert_exhaustion"
	assertReturnCanonicalNanCommand  commandType = "assert_return_canonical_nan"
	assertReturnArithmeticNanCommand commandType = "assert_return_arithmetic_nan"
	assertUnlinkableCommand          commandType = "assert_unlinkable"
	registerCommand                  commandType = "register"
)

type script struct {
	FileName string    `json:"source_filename"`
	Commands []command `json:"commands"`
}

type command struct {
	Type       commandType `json:"type"`
	Line       int         `json:"line"`
	Filename   string      `json:"filename"`
	Name       string      `json:"name"`
	Action     action      `json:"action"`
	Text       string      `json:"text"`
	ModuleType string      `json:"module_type"`
	Expected   []value     `json:"expected"`
}

type action struct {
	Type     string  `json:"type"`
	Module   string  `json:"module"`
	Field    string  `json:"field"`
	Args     []value `json:"args"`
	Expected []value `json:"expected"`
}

type value struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type module struct {
	filename string
	code     *exec.Code
	context  *exec.Context
}

var resolver = exec.MapResolver(map[string]interface{}{
	"spectest.global_i32": float64(666),
	"spectest.global_i64": float64(666),
	"spectest.global_f32": float64(666.6),
	"spectest.global_f64": float64(666.6),
	"spectest.print": func(ctx *exec.Context) uint32 {
		return 0
	},
	"spectest.print_i32": func(ctx *exec.Context, x uint32) uint32 {
		return 0
	},
	"spectest.print_f32": func(ctx *exec.Context, x uint32) uint32 {
		return 0
	},
	"spectest.print_f64": func(ctx *exec.Context, x uint32) uint32 {
		return 0
	},
	"spectest.print_i32_f32": func(ctx *exec.Context, x, y uint32) uint32 {
		return 0
	},
	"spectest.print_f64_f64": func(ctx *exec.Context, x, y uint32) uint32 {
		return 0
	},
})

func newModule(modulePath string) (*module, error) {
	targetPath := modulePath + ".so"
	err := compile.CompileNativeLibrary(&compile.Config{
		Wasm2cPath: wasm2cPath,
	}, targetPath, modulePath)
	if err != nil {
		return nil, err
	}
	targetPath, _ = filepath.Abs(targetPath)
	code, err := exec.NewCode(targetPath, resolver)
	if err != nil {
		return nil, err
	}
	context, err := exec.NewContext(code, exec.DefaultContextConfig())
	if err != nil {
		return nil, err
	}
	return &module{
		filename: modulePath,
		code:     code,
		context:  context,
	}, nil
}

func (m *module) Release() {
	m.context.Release()
	m.code.Release()
}

type commandRunner struct {
	basedir    string
	testFile   string
	modules    map[string]*module
	lastModule *module
	moduleList []*module
	passed     int
	total      int
}

func newCommandRunner() *commandRunner {
	return &commandRunner{
		modules: make(map[string]*module),
	}
}

func (c *commandRunner) Close() {
	for _, m := range c.moduleList {
		m.Release()
	}
}

func (c *commandRunner) tallyCommand(cmd command, err error) {
	if err == nil {
		c.passed++
	} else {
		fmt.Printf("%s:%d %s\n", c.testFile, cmd.Line, err)
	}
	c.total++
}

func (c *commandRunner) newModule(name string) (*module, error) {
	return newModule(filepath.Join(c.basedir, name))
}

func (c *commandRunner) Run(basedir string, script *script) {
	c.basedir = basedir
	c.testFile = filepath.Base(script.FileName)
	for _, cmd := range script.Commands {
		// fmt.Printf("run %s:%d %s\n", c.testFile, cmd.Line, cmd.Type)
		switch cmd.Type {
		case moduleCommand:
			c.tallyCommand(cmd, c.onModuleCommand(cmd))
		case actionCommand:
			c.tallyCommand(cmd, c.onActionCommand(cmd))
		case registerCommand:
			continue
		case assertReturnCommand:
			c.tallyCommand(cmd, c.onAssertReturn(cmd))
		case assertInvalidCommand:
			c.tallyCommand(cmd, c.onAssertInvalid(cmd))
		case assertTrapCommand:
			c.tallyCommand(cmd, c.onAssertTrap(cmd))
		case assertMalformedCommand:
			c.tallyCommand(cmd, c.onAssertMalformed(cmd))
		case assertExhaustionCommand:
			c.tallyCommand(cmd, c.onAssertExhaustion(cmd))
		case assertReturnCanonicalNanCommand:
			c.tallyCommand(cmd, c.onAssertReturnCanonicalNan(cmd))
		case assertReturnArithmeticNanCommand:
			c.tallyCommand(cmd, c.onAssertReturnArithmeticNan(cmd))
		case assertUnlinkableCommand:
			c.tallyCommand(cmd, c.onAssertUnlinkableCommand(cmd))
		default:
			c.tallyCommand(cmd, fmt.Errorf("unsupported command:%s", cmd.Type))
		}
	}
}

func parseValues(values []value) []int64 {
	var result []int64
	for _, v := range values {
		n, _ := strconv.ParseUint(v.Value, 10, 64)
		result = append(result, *(*int64)(unsafe.Pointer(&n)))
	}
	return result
}

func (c *commandRunner) runAction(action *action) (int64, error) {
	var m *module
	if action.Module != "" {
		m = c.modules[action.Module]
	} else {
		m = c.lastModule
	}
	if m == nil {
		return 0, errors.New("nil module")
	}
	// fmt.Println("module: ", m.filename)
	switch action.Type {
	case "invoke":
		args := parseValues(action.Args)
		return m.context.Exec(action.Field, args)
	case "get":
		// TODO:
		return 0, nil
	}
	return 0, nil

}
func (c *commandRunner) onModuleCommand(cmd command) error {
	// fmt.Printf("on module:%#v\n", cmd)
	m, err := c.newModule(cmd.Filename)
	if err != nil {
		return err
	}
	c.lastModule = m
	if cmd.Name != "" {
		c.modules[cmd.Name] = m
	}
	c.moduleList = append(c.moduleList, m)
	return nil
}

func (c *commandRunner) onActionCommand(cmd command) error {
	_, err := c.runAction(&cmd.Action)
	return err
}

func (c *commandRunner) onAssertReturn(cmd command) error {
	ret, err := c.runAction(&cmd.Action)
	if err != nil {
		return err
	}
	expect := parseValues(cmd.Expected)
	if len(expect) != 0 && expect[0] != ret {
		return fmt.Errorf("expect %d, got %d", expect[0], ret)
	}
	return nil
}

func (c *commandRunner) expectTrap(cmd command) (exec.Trap, error) {
	_, err := c.runAction(&cmd.Action)
	if err == nil {
		return nil, errors.New("expect error got nil")
	}
	trap, ok := err.(*exec.TrapError)
	if !ok {
		return nil, errors.New("not trap error")
	}
	return trap.Trap, nil
}

func (c *commandRunner) onAssertTrap(cmd command) error {
	_, err := c.expectTrap(cmd)
	return err
}

func (c *commandRunner) onAssertInvalid(cmd command) error {
	module, err := c.newModule(cmd.Filename)
	if err != nil {
		return nil
	}
	module.Release()
	return errors.New("expect error, got nil")
}

func (c *commandRunner) onAssertMalformed(cmd command) error {
	if cmd.ModuleType == "text" {
		return nil
	}
	return c.onAssertInvalid(cmd)
}

func (c *commandRunner) onAssertExhaustion(cmd command) error {
	trap, err := c.expectTrap(cmd)
	if err != nil {
		return err
	}
	if trap != exec.TrapCallStackExhaustion {
		return errors.New("bad trap")
	}
	return nil
}

func (c *commandRunner) onAssertReturnCanonicalNan(cmd command) error {
	ret, err := c.runAction(&cmd.Action)
	if err != nil {
		return err
	}
	var f64 float64
	switch cmd.Expected[0].Type {
	case "f32":
		f32 := *(*float32)(unsafe.Pointer(&ret))
		f64 = float64(f32)
	case "f64":
		f64 = *(*float64)(unsafe.Pointer(&ret))
	}
	if !math.IsNaN(f64) {
		return errors.New("not NaN")
	}
	return nil
}

func (c *commandRunner) onAssertReturnArithmeticNan(cmd command) error {
	return c.onAssertReturnCanonicalNan(cmd)
}

func (c *commandRunner) onAssertUnlinkableCommand(cmd command) error {
	return nil
}

type runnerConfig struct {
	StageDir string
}

type testRunner struct {
	cfg        *runnerConfig
	stageDir   string
	total      int
	passed     int
	lastTotal  int
	lastPassed int
}

func newTestRunner(cfg *runnerConfig) *testRunner {
	stageDir := cfg.StageDir
	if stageDir == "" {
		var err error
		stageDir, err = ioutil.TempDir("", "xvm-spectest")
		if err != nil {
			panic(err)
		}
	}
	return &testRunner{
		cfg:      cfg,
		stageDir: stageDir,
	}
}

func (t *testRunner) Close() {
	if t.cfg.StageDir == "" {
		os.RemoveAll(t.stageDir)
	}
}

func (t *testRunner) RunJSONScript(name string) {
	basedir := filepath.Dir(name)
	var script script
	buf, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(buf, &script)
	if err != nil {
		log.Fatal(err)
	}

	runner := newCommandRunner()
	runner.Run(basedir, &script)
	runner.Close()
	t.total += runner.total
	t.passed += runner.passed
	t.lastTotal = runner.total
	t.lastPassed = runner.passed
}

func (t *testRunner) RunTest(wastFile string) {
	basename := filepath.Base(wastFile)
	fmt.Printf("\n======== RUN %s\n", basename)
	testdir := filepath.Join(t.stageDir, basename)
	os.MkdirAll(testdir, 0700)
	jsonfile := filepath.Join(testdir, basename+".json")
	cmd := osexec.Command(wast2jsonPath, "-o", jsonfile, wastFile)
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	t.RunJSONScript(jsonfile)

	fmt.Printf("total:%d passed:%d\n", t.lastTotal, t.lastPassed)
	fmt.Print("======== DONE\n")
}

func (t *testRunner) RunTestDir(dir string) {
	freg := regexp.MustCompile(`^name|^linking|^skip-stack`)
	testFiles, err := filepath.Glob(filepath.Join(dir, "*.wast"))
	if err != nil {
		panic(err)
	}
	for _, testFile := range testFiles {
		fileBase := filepath.Base(testFile)
		if freg.MatchString(fileBase) {
			fmt.Printf("skip %s\n", testFile)
			continue
		}
		t.RunTest(testFile)
	}
}

func lookupTools() error {
	return nil
}

var (
	stageDir = flag.String("s", "", "stage dir")
)

func main() {
	flag.Parse()
	err := lookupTools()
	if err != nil {
		log.Fatal(err)
	}
	if flag.NArg() == 0 {
		fmt.Println("usage: ./spectest xxx.wast|directory")
		return
	}
	testTarget := flag.Arg(0)
	stat, err := os.Stat(testTarget)
	if err != nil {
		log.Fatal(err)
	}
	runner := newTestRunner(&runnerConfig{
		StageDir: *stageDir,
	})
	if stat.IsDir() {
		runner.RunTestDir(testTarget)
	} else {
		runner.RunTest(testTarget)
	}
	fmt.Printf("total:%d passed:%d(%0.2f%%)\n", runner.total, runner.passed, 100*float32(runner.passed)/float32(runner.total))
	runner.Close()
}
