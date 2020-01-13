package mkfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Library represents a C++ library module
type Library struct {
	Name string
	Dir  string
}

// Application represents a C++ binary module
type Application struct {
	Name    string
	LinkAll bool
	Libs    []string
}

// Project represents a C++ project
type Project struct {
	root     string
	stageDir string
	cxxflags []string
	ldflags  []string
	apps     map[string]*Application
	libs     map[string]*Library
	mkfile   Makefile
	depfiles []string
}

// NewProject instance a new Project
func NewProject() *Project {
	return &Project{
		apps:     make(map[string]*Application),
		libs:     make(map[string]*Library),
		stageDir: "build",
	}
}

// WithRoot set the root directory of project
func (p *Project) WithRoot(root string) *Project {
	p.root = root
	return p
}

// WithCxxFlags set the CXXFLAGS during compiling cxx file
func (p *Project) WithCxxFlags(flags []string) *Project {
	p.cxxflags = flags
	return p
}

// WithLDFlags set the LDFLAGS during linking binary
func (p *Project) WithLDFlags(flags []string) *Project {
	p.ldflags = flags
	return p
}

// AddLibrary add a new library module to project
func (p *Project) AddLibrary(lib *Library) *Project {
	p.libs[lib.Name] = lib
	return p
}

// AddApplication add a new application module to project
func (p *Project) AddApplication(app *Application) *Project {
	p.apps[app.Name] = app
	return p
}

// GenerateMakeFile generates makefile to w
func (p *Project) GenerateMakeFile(w io.Writer) error {
	err := p.addPhonyTasks()
	if err != nil {
		return err
	}

	err = p.addAutoCxxTask()
	if err != nil {
		return err
	}

	for _, lib := range p.libs {
		err = p.addLibraryTask(lib)
		if err != nil {
			return err
		}
	}

	for _, app := range p.apps {
		err = p.addApplicationTask(app)
		if err != nil {
			return err
		}
	}

	p.addHeaders([]string{
		"CXXFLAGS ?= " + strings.Join(p.cxxflags, " "),
		"LDFLAGS ?= " + strings.Join(p.ldflags, " "),
	})
	p.addTails([]string{
		"-include " + strings.Join(p.depfiles, " "),
	})

	writer := NewMakeFileWriter(w)
	writer.Write(&p.mkfile)
	return nil
}

func (p *Project) addHeader(header string) {
	p.mkfile.Headers = append(p.mkfile.Headers, header)
}

func (p *Project) addHeaders(headers []string) {
	p.mkfile.Headers = append(p.mkfile.Headers, headers...)
}

func (p *Project) addTails(tails []string) {
	p.mkfile.Tails = append(p.mkfile.Tails, tails...)
}

func (p *Project) addTask(t *Task) {
	p.mkfile.Tasks = append(p.mkfile.Tasks, *t)
}

func (p *Project) addPhonyTasks() error {
	p.addTask(&Task{
		Target: ".PHONY",
		Deps:   []string{"all", "clean"},
	})
	var allTargets []string
	for _, app := range p.apps {
		allTargets = append(allTargets, app.Name)
	}
	p.addTask(&Task{
		Target: "all",
		Deps:   allTargets,
	})
	p.addTask(&Task{
		Target: "clean",
		Actions: []string{
			"$(RM) -r build",
		},
	})
	return nil
}

func (p *Project) addAutoCxxTask() error {
	target := p.stagePath("%.cc.o")
	task := &Task{
		Target: target,
		Deps:   []string{"%.cc"},
		Actions: []string{
			`@mkdir -p $(dir $@)`,
			`@echo CC $<`,
			`@$(CXX) $(CXXFLAGS) -MMD -MP -c $< -o $@`,
		},
	}
	p.addTask(task)
	return nil
}

func (p *Project) addLibraryTask(lib *Library) error {
	task, err := p.newLibraryTask(lib)
	if err != nil {
		return err
	}
	p.addTask(task)
	return nil
}

func (p *Project) newLibraryTask(lib *Library) (*Task, error) {
	objects := p.libObjectFiles(lib)
	for _, obj := range objects {
		p.addDepFile(obj[:len(obj)-1] + "d")
	}

	return &Task{
		Target: p.stageLibpath(lib.Name),
		Deps:   objects,
		Actions: []string{
			"@$(AR) -rc $@ $^",
			"@$(RANLIB) $@",
		},
	}, nil
}

func (p *Project) addApplicationTask(app *Application) error {
	task, err := p.newApplicationTask(app)
	if err != nil {
		return err
	}
	p.addTask(task)
	p.addTask(&Task{
		Target: app.Name,
		Deps:   []string{task.Target},
	})
	return nil
}

func (p *Project) newApplicationTask(app *Application) (*Task, error) {
	var libs []*Library
	if app.LinkAll {
		for _, lib := range p.libs {
			libs = append(libs, lib)
		}
	} else {
		for _, libname := range app.Libs {
			lib, ok := p.libs[libname]
			if !ok {
				return nil, fmt.Errorf("missing library %s", libname)
			}
			libs = append(libs, lib)
		}
	}
	var deps []string
	var objects []string
	for _, lib := range libs {
		libObjects := p.libObjectFiles(lib)
		objects = append(objects, libObjects...)
		deps = append(deps, p.stageLibpath(lib.Name))
	}
	target := p.stagePath(app.Name)

	t := &Task{
		Target: target,
		Deps:   deps,
		Actions: []string{
			`@echo LD ` + app.Name,
			fmt.Sprintf(`@$(CXX) -o $@ ` + strings.Join(objects, " ") + " $(LDFLAGS)"),
		},
	}
	return t, nil
}

func (p *Project) libSourceFiles(lib *Library) []string {
	cpps, _ := filepath.Glob(filepath.Join(lib.Dir, "*.cc"))
	cs, _ := filepath.Glob(filepath.Join(lib.Dir, "*.c"))
	return append(cpps, cs...)
}

func (p *Project) libObjectFiles(lib *Library) []string {
	var objects []string
	srcs := p.libSourceFiles(lib)
	for _, src := range srcs {
		object := p.stagePath(src + ".o")
		objects = append(objects, object)
	}

	return objects
}

func (p *Project) stagePath(fpath string) string {
	return filepath.Join(p.stageDir, fpath)
}

func (p *Project) stageLibpath(name string) string {
	return p.stagePath("lib" + name + ".a")
}

func (p *Project) addDepFile(name string) {
	p.depfiles = append(p.depfiles, name)
}

type compileCommand struct {
	Directory string `json:"directory"`
	Command   string `json:"command"`
	File      string `json:"file"`
}

// GenerateCompileCommands generates compile_commands.json to io.Writer
func (p *Project) GenerateCompileCommands(w io.Writer) error {
	if p.root == "" {
		return errors.New("missing project root")
	}
	var commands []*compileCommand
	for _, lib := range p.libs {
		libcmds, err := p.newLibraryCompileCommand(lib)
		if err != nil {
			return err
		}
		commands = append(commands, libcmds...)
	}
	output, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}
	_, err = w.Write(output)
	return err
}

func (p *Project) newLibraryCompileCommand(lib *Library) ([]*compileCommand, error) {
	var commands []*compileCommand
	cxxflags := strings.Join(p.cxxflags, " ")
	srcs := p.libSourceFiles(lib)
	for _, src := range srcs {
		obj := p.stagePath(src + ".o")
		buildCommand := fmt.Sprintf("g++ %s -c -o%s %s", cxxflags, obj, src)
		command := &compileCommand{
			Directory: p.root,
			File:      src,
			Command:   buildCommand,
		}
		commands = append(commands, command)
	}
	return commands, nil
}
