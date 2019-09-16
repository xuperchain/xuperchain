package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuperchain/xuperunion/xvm/compile"
	"github.com/xuperchain/xuperunion/xvm/debug"
	"github.com/xuperchain/xuperunion/xvm/exec"
	"github.com/xuperchain/xuperunion/xvm/runtime/emscripten"
	gowasm "github.com/xuperchain/xuperunion/xvm/runtime/go"
	"github.com/xuperchain/xuperunion/xvm/runtime/wasi"
)

var (
	centry      = flag.String("entry", "main", "entry function")
	compileOnly = flag.Bool("c", false, "only compile wasm file")
	environ     = flag.String("e", "c", "environ, c or go")
)

func replaceExt(name, ext string) string {
	dir, file := filepath.Split(name)
	idx := strings.Index(file, ".")
	if idx == -1 {
		file = file + ext
	} else {
		file = file[:idx] + ext
	}
	return filepath.Join(dir, file)
}

func compileLibrary(wasmpath string) (string, error) {
	tmpdir, err := ioutil.TempDir("", "xvm-exec-test")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpdir)
	cfg := &compile.Config{
		Wasm2cPath: "wasm2c",
		OptLevel:   0,
	}
	libpath := replaceExt(wasmpath, ".so")
	err = compile.CompileNativeLibrary(cfg, libpath, wasmpath)
	if err != nil {
		return "", err
	}
	return libpath, nil
}

func prepareArgs(mem []byte, args []string, envs []string) (int, int) {
	argc := len(args)
	offset := 4 << 10
	strdup := func(s string) int {
		copy(mem[offset:], s+"\x00")
		ptr := offset
		offset += len(s) + (8 - len(s)%8)
		return ptr
	}
	var argvAddr []int
	for _, arg := range args {
		argvAddr = append(argvAddr, strdup(arg))
	}

	argvAddr = append(argvAddr, len(envs))
	for _, env := range envs {
		argvAddr = append(argvAddr, strdup(env))
	}

	argv := offset
	buf := bytes.NewBuffer(mem[offset:offset])
	for _, addr := range argvAddr {
		if *environ == "go" {
			binary.Write(buf, binary.LittleEndian, uint64(addr))
		} else {
			binary.Write(buf, binary.LittleEndian, uint32(addr))
		}
	}
	return argc, argv
}

func run(modulePath string, args []string) error {
	fullepath, err := filepath.Abs(modulePath)
	if err != nil {
		return err
	}
	resolver := exec.NewMultiResolver(resolver, gowasm.NewResolver(), emscripten.NewResolver(), wasi.NewResolver())
	code, err := exec.NewCode(fullepath, resolver)
	if err != nil {
		return err
	}
	defer code.Release()

	ctx, err := exec.NewContext(code, exec.DefaultContextConfig())
	if err != nil {
		return err
	}
	if *environ == "go" {
		gowasm.RegisterRuntime(ctx)
	}
	defer ctx.Release()
	debug.SetWriter(ctx, os.Stderr)
	var entry string
	switch *environ {
	case "c":
		entry = *centry
	case "go":
		entry = "run"
	}

	var argc, argv int
	if ctx.Memory() != nil {
		argc, argv = prepareArgs(ctx.Memory(), args, nil)
	}
	ret, err := ctx.Exec(entry, []int64{int64(argc), int64(argv)})
	fmt.Println("gas: ", ctx.GasUsed())
	fmt.Println("ret: ", ret)
	return err
}

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	ext := filepath.Ext(filename)
	var target string
	var err error
	switch ext {
	case ".wasm":
		target, err = compileLibrary(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		if *compileOnly {
			return
		}
	case ".so":
		target = filename
	default:
		log.Fatalf("bad file ext:%s", ext)
	}

	err = run(target, flag.Args()[0:])
	if err != nil {
		log.Fatal(err)
	}
}
