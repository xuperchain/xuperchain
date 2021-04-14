package native

import (
	"encoding/hex"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sync"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	"github.com/xuperchain/xuperchain/core/pb"
)

const (
	processCount = 4
)

type processManager struct {
	cfg       *config.NativeConfig
	basedir   string
	chainAddr string
	mutex     sync.Mutex
	contracts map[string][]*contractProcess
}

func newProcessManager(cfg *config.NativeConfig, basedir string, chainAddr string) (*processManager, error) {
	return &processManager{
		cfg:       cfg,
		basedir:   basedir,
		chainAddr: chainAddr,
		contracts: make(map[string][]*contractProcess),
	}, nil
}

func (p *processManager) makeProcess(name string, desc *pb.WasmCodeDesc, code []byte) (*contractProcess, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	hash := nativeCodeHash(name, desc)
	processes, ok := p.contracts[hash]
	if !ok {
		processes = make([]*contractProcess, processCount)
		p.contracts[hash] = processes
	}

	processDir := filepath.Join(p.basedir, name)
	err := os.MkdirAll(processDir, 0755)
	if err != nil {
		return nil, err
	}
	contractFile := nativeCodeFileName(desc)
	processBin := filepath.Join(processDir, contractFile)
	if _, err := os.Stat(processBin); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := ioutil.WriteFile(processBin, code, 0755); err != nil {
			return nil, err
		}
	}

	process, err := newContractProcess(p.cfg, name, processDir, p.chainAddr, desc)
	if err != nil {
		return nil, err
	}

	err = process.Start()
	if err != nil {
		return nil, err
	}
	processes[n] = process

	return process, nil
}

func (p *processManager) lookupProcess(name string, desc *pb.WasmCodeDesc, n int) (*contractProcess, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	hash := nativeCodeHash(name, desc)
	processes, ok := p.contracts[hash]
	if !ok {
		return nil, false
	}
	if processes == nil || processes[n] == nil {
		return nil, false
	}
	return processes[n], true
}

func (p *processManager) GetProcess(name string, cp bridge.ContractCodeProvider) (*contractProcess, error) {
	desc, err := cp.GetContractCodeDesc(name)
	if err != nil {
		return nil, err
	}

	n := rand.Int() % processCount

	process, ok := p.lookupProcess(name, desc, n)
	if ok {
		return process, nil
	}

	code, err := cp.GetContractCode(name)
	if err != nil {
		return nil, err
	}

	process, err = p.makeProcess(name, desc, code)
	if err != nil {
		return nil, err
	}
	return process, nil
}

func nativeCodeHash(name string, desc *pb.WasmCodeDesc) string {
	return name + hex.EncodeToString(desc.GetDigest())
}

func nativeCodeFileName(desc *pb.WasmCodeDesc) string {
	var suffix string
	switch desc.GetRuntime() {
	case "java":
		suffix = ".jar"
	}
	hash := hex.EncodeToString(desc.GetDigest()[0:3])
	return "nativecode-" + hash + suffix
}
