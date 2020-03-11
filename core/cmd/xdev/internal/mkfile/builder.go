package mkfile

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	StageDir    = "build"
	MainPackage = "main"
	OutFileName = "a.wasm"
)

// Package is the basic compile unit
type Package struct {
	Name string
	Path string
	Deps []*Package
}

// Builder generate makefile
type Builder struct {
	root     string
	entry    *Package
	pkgs     map[string]*Package
	buildDir string
	cxxflags []string
	ldflags  []string
	mkfile   Makefile
	objfiles []string
	srcfiles []string
	depfiles []string
}

// NewBuilder instance a build
func NewBuilder() *Builder {
	return &Builder{
		buildDir: StageDir,
		pkgs:     make(map[string]*Package),
	}
}

// WithCxxFlags set the CXXFLAGS during compiling cxx file
func (b *Builder) WithCxxFlags(flags []string) *Builder {
	b.cxxflags = flags
	return b
}

// WithLDFlags set the LDFLAGS during linking binary
func (b *Builder) WithLDFlags(flags []string) *Builder {
	b.ldflags = flags
	return b
}

func (b *Builder) Parse(entry *Package) error {
	var err error
	b.root, err = filepath.Abs(entry.Path)
	if err != nil {
		return err
	}

	b.entry = entry
	b.pkgs[entry.Name] = entry
	for _, pkg := range entry.Deps {
		b.pkgs[pkg.Name] = pkg
	}

	for _, pkg := range b.pkgs {
		err := b.parsePackage(pkg)
		if err != nil {
			return err
		}
		includePath := b.externalPkgPath(filepath.Join(pkg.Path, "src"))
		b.cxxflags = append(b.cxxflags, "-I"+includePath)
	}
	return nil
}

// GenerateMakeFile generates makefile to w
func (b *Builder) GenerateMakeFile(w io.Writer) error {
	err := b.addPhonyTasks()
	if err != nil {
		return err
	}

	err = b.addObjectFileTask()
	if err != nil {
		return err
	}

	err = b.addBuildEntryTask()
	if err != nil {
		return err
	}

	b.addHeaders([]string{
		"CXXFLAGS ?= " + strings.Join(b.cxxflags, " "),
		"LDFLAGS ?= " + strings.Join(b.ldflags, " "),
	})
	b.addTails([]string{
		"-include " + strings.Join(b.depfiles, " "),
	})

	writer := NewMakeFileWriter(w)
	writer.Write(&b.mkfile)
	return nil
}

func (b *Builder) addHeader(header string) {
	b.mkfile.Headers = append(b.mkfile.Headers, header)
}

func (b *Builder) addHeaders(headers []string) {
	b.mkfile.Headers = append(b.mkfile.Headers, headers...)
}

func (b *Builder) addTails(tails []string) {
	b.mkfile.Tails = append(b.mkfile.Tails, tails...)
}

func (b *Builder) addTask(t *Task) {
	b.mkfile.Tasks = append(b.mkfile.Tasks, *t)
}

func (b *Builder) addPhonyTasks() error {
	b.addTask(&Task{
		Target: ".PHONY",
		Deps:   []string{"all", "build", "clean"},
	})
	b.addTask(&Task{
		Target: "all",
		Deps:   []string{"build"},
	})

	b.addTask(&Task{
		Target: "clean",
		Actions: []string{
			"$(RM) -r build",
		},
	})
	return nil
}

func (b *Builder) addObjectFileTask() error {
	for i := range b.srcfiles {
		src := b.srcfiles[i]
		relsrc := b.externalPkgPath(src)
		obj := b.objfiles[i]
		task := &Task{
			Target: obj,
			Deps:   []string{relsrc},
			Actions: []string{
				`@mkdir -p $(dir $@)`,
				`@echo CC $<`,
				`@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@`,
			},
		}
		b.addTask(task)
	}

	return nil
}

func (b *Builder) addBuildEntryTask() error {
	var buildTaskDep string

	if b.entry.Name == MainPackage {
		wasmTask := &Task{
			Target: b.buildPath(OutFileName),
			Deps:   b.objfiles,
			Actions: []string{
				`@echo LD a.wasm`,
				fmt.Sprintf(`@$(CXX) -o $@ $^ $(LDFLAGS)`),
			},
		}
		b.addTask(wasmTask)
		buildTaskDep = wasmTask.Target
	} else {
		libTask := &Task{
			Target: b.buildPath("liba.a"),
			Deps:   b.objfiles,
			Actions: []string{
				"@$(AR) -rc $@ $^",
				"@$(RANLIB) $@",
			},
		}
		b.addTask(libTask)
		buildTaskDep = libTask.Target
	}

	b.addTask(&Task{
		Target: "build",
		Deps:   []string{buildTaskDep},
	})
	return nil
}

func (b *Builder) buildPath(fpath string) string {
	return filepath.Join(b.buildDir, fpath)
}

func (b *Builder) buildLibpath(name string) string {
	return b.buildPath("lib" + name + ".a")
}

// 如果path是entry package的子目录，则返回相对目录
// 否则返回原目录
// 否则在编译容器里面找不到对应的目录
func (b *Builder) externalPkgPath(path string) string {
	if strings.HasPrefix(path, b.root) {
		return b.relpath(path)
	}
	return path
}

func (b *Builder) relpath(p string) string {
	path, _ := filepath.Rel(b.root, p)
	return path
}

func (b *Builder) objectFilePath(src string) string {
	return b.buildPath(objectFileName(src))
}

func (b *Builder) parsePackage(pkg *Package) error {
	srcs, objects := b.pkgObjectFiles(pkg)
	for i, obj := range objects {
		b.addDepFile(obj[:len(obj)-1] + "d")
		b.addObjFile(obj)
		b.addSrcFile(srcs[i])
	}

	return nil
}

func (b *Builder) pkgSourceFiles(pkg *Package) []string {
	var files []string
	srcdir := filepath.Join(pkg.Path, "src")
	filepath.Walk(srcdir, func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return nil
		}
		ext := filepath.Ext(path)
		switch ext {
		case ".cc", ".cpp", ".c":
			files = append(files, path)
		}
		return nil
	})
	return files
}

func (b *Builder) pkgObjectFiles(pkg *Package) ([]string, []string) {
	var objects []string
	srcs := b.pkgSourceFiles(pkg)
	for _, src := range srcs {
		object := b.objectFilePath(src)
		objects = append(objects, object)
	}

	return srcs, objects
}

func (b *Builder) addDepFile(name string) {
	b.depfiles = append(b.depfiles, name)
}

func (b *Builder) addObjFile(name string) {
	b.objfiles = append(b.objfiles, name)
}

func (b *Builder) addSrcFile(name string) {
	b.srcfiles = append(b.srcfiles, name)
}

// GenerateCompileCommands generates compile_commands.json to io.Writer
func (b *Builder) GenerateCompileCommands(w io.Writer) error {
	commands, err := b.newLibraryCompileCommand()
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(output)
	return err
}

type compileCommand struct {
	Directory string `json:"directory"`
	Command   string `json:"command"`
	File      string `json:"file"`
}

func (b *Builder) newLibraryCompileCommand() ([]*compileCommand, error) {
	var commands []*compileCommand
	cxxflags := strings.Join(b.cxxflags, " ")
	for _, src := range b.srcfiles {
		obj := b.objectFilePath(src)
		buildCommand := fmt.Sprintf("g++ %s -c -o%s %s", cxxflags, obj, src)
		command := &compileCommand{
			Directory: b.root,
			File:      src,
			Command:   buildCommand,
		}
		commands = append(commands, command)
	}
	return commands, nil
}

func objectFileName(srcfile string) string {
	h := fnv.New64a()
	h.Write([]byte(srcfile))
	return fmt.Sprintf("%d.o", h.Sum64())
}
