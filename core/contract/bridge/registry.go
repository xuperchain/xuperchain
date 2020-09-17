package bridge

import (
	"fmt"
	"sync"
)

type ContractType string

const (
	TypeWasm   ContractType = "wasm"
	TypeNative ContractType = "native"
	TypeEvm    ContractType = "evm"
)

var defaultRegistry = newRegistry()

type registry struct {
	mutex   sync.Mutex
	drivers map[ContractType]map[string]NewInstanceCreatorFunc
}

func newRegistry() *registry {
	return &registry{
		drivers: make(map[ContractType]map[string]NewInstanceCreatorFunc),
	}
}

func (r *registry) Register(tp ContractType, name string, driver NewInstanceCreatorFunc) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	m, ok := r.drivers[tp]
	if !ok {
		m = make(map[string]NewInstanceCreatorFunc)
		r.drivers[tp] = m
	}
	if _, ok := m[name]; ok {
		panic(fmt.Sprintf("driver %s for %s exists", name, tp))
	}
	m[name] = driver
}

func (r *registry) Open(tp ContractType, name string, config *InstanceCreatorConfig) (InstanceCreator, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	m, ok := r.drivers[tp]
	if !ok {
		return nil, fmt.Errorf("driver for contract type %s not found", tp)
	}
	driverFunc, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("driver %s for %s not found", name, tp)
	}
	return driverFunc(config)
}

// Register makes a contract driver available by the provided type and name
func Register(tp ContractType, name string, driver NewInstanceCreatorFunc) {
	defaultRegistry.Register(tp, name, driver)
}

// Open opens a contract virtual machine specified by its driver type and name
func Open(tp ContractType, name string, config *InstanceCreatorConfig) (InstanceCreator, error) {
	return defaultRegistry.Open(tp, name, config)
}
