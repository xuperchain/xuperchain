// Copyright 2014 dong<ddliuhb@gmail.com>.
// Licensed under the MIT license.
//
// Motto - Modular Javascript environment.
package motto

import (
	"path/filepath"
	"sync"

	"github.com/robertkrimen/otto"
)

// Globally registered modules
var globalModules map[string]ModuleLoader = make(map[string]ModuleLoader)

// Globally registered paths (paths to search for modules)
var globalPaths []string

// Globally protects global state
var globalMu sync.RWMutex

// Motto is modular vm environment
type Motto struct {
	// Motto is based on otto
	*otto.Otto

	// try to read source map
	SourceMapEnabled bool

	// Modules that registered for current vm
	modules   map[string]ModuleLoader
	modulesMu sync.RWMutex

	// Location to search for modules
	paths   []string
	pathsMu sync.RWMutex

	// Onece a module is required by vm, the exported value is cached for further
	// use.
	moduleCache   map[string]otto.Value
	moduleCacheMu sync.RWMutex
}

// Run a module or file
func (m *Motto) Run(name string) (otto.Value, error) {
	if ok, _ := isFile(name); ok {
		name, _ = filepath.Abs(name)
	}

	return m.Require(name, ".")
}

// Require a module with cache
func (m *Motto) Require(id, pwd string) (otto.Value, error) {
	if cache, ok := m.cachedModule(id); ok {
		return cache, nil
	}

	loader := m.module(id)
	if loader == nil {
		loader = Module(id)
	}

	if loader != nil {
		v, err := loader(m)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		m.addCachedModule(id, v)
		return v, nil
	}

	filename, err := FindFileModule(id, pwd, append(m.paths, globalPaths...))
	if err != nil {
		return otto.UndefinedValue(), err
	}

	// resove id
	id = filename

	if cache, ok := m.cachedModule(id); ok {
		return cache, nil
	}

	v, err := CreateLoaderFromFile(id)(m)

	if err != nil {
		return otto.UndefinedValue(), err
	}

	m.addCachedModule(id, v)
	return v, nil
}

func (m *Motto) addCachedModule(id string, v otto.Value) {
	m.moduleCacheMu.Lock()
	m.moduleCache[id] = v
	m.moduleCacheMu.Unlock()
}

func (m *Motto) cachedModule(id string) (otto.Value, bool) {
	m.moduleCacheMu.RLock()
	defer m.moduleCacheMu.RUnlock()
	v, ok := m.moduleCache[id]
	return v, ok
}

// ClearModule clear all registered module from current vm
func (m *Motto) ClearModule() {
	m.moduleCacheMu.Lock()
	m.moduleCache = make(map[string]otto.Value)
	m.moduleCacheMu.Unlock()
}

// AddModule registers a new module to current vm.
func (m *Motto) AddModule(id string, l ModuleLoader) {
	m.modulesMu.Lock()
	m.modules[id] = l
	m.modulesMu.Unlock()
}

func (m *Motto) module(id string) ModuleLoader {
	m.modulesMu.RLock()
	defer m.modulesMu.RUnlock()
	return m.modules[id]
}

// AddPath adds paths to search for modules.
func (m *Motto) AddPath(paths ...string) {
	m.pathsMu.Lock()
	m.paths = append(m.paths, paths...)
	m.pathsMu.Unlock()
}

// AddModule registers global module
func AddModule(id string, m ModuleLoader) {
	globalMu.Lock()
	globalModules[id] = m
	globalMu.Unlock()
}

// Module returns ModuleLoader for a given ID.
func Module(id string) ModuleLoader {
	globalMu.RLock()
	defer globalMu.RUnlock()

	return globalModules[id]
}

// AddPath registers global path.
func AddPath(paths ...string) {
	globalMu.Lock()
	globalPaths = append(globalPaths, paths...)
	globalMu.Unlock()
}

// Run module by name in the motto module environment.
func Run(name string) (*Motto, otto.Value, error) {
	vm := New()
	v, err := vm.Run(name)

	return vm, v, err
}

// New creates a new motto vm instance.
func New() *Motto {
	return &Motto{
		Otto:        otto.New(),
		modules:     make(map[string]ModuleLoader),
		moduleCache: make(map[string]otto.Value),
	}
}
