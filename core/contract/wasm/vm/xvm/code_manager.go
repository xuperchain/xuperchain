package xvm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/xuperchain/xuperunion/common/log"
	"github.com/xuperchain/xuperunion/contract/wasm/vm"
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/xvm/compile"
	"github.com/xuperchain/xuperunion/xvm/exec"
	"golang.org/x/sync/singleflight"
)

type compileFunc func([]byte, string) error
type makeExecCodeFunc func(libpath string) (exec.Code, error)

type contractCode struct {
	ContractName string
	ExecCode     exec.Code
	Desc         pb.WasmCodeDesc
}

type codeManager struct {
	basedir      string
	compileCode  compileFunc
	makeExecCode makeExecCodeFunc

	makeCacheLock singleflight.Group

	mutex sync.Mutex // protect codes
	codes map[string]*contractCode
}

func newCodeManager(basedir string, compile compileFunc, makeExec makeExecCodeFunc) *codeManager {
	return &codeManager{
		basedir:      basedir,
		compileCode:  compile,
		makeExecCode: makeExec,
		codes:        make(map[string]*contractCode),
	}
}

func codeDescEqual(a, b *pb.WasmCodeDesc) bool {
	return bytes.Equal(a.GetDigest(), b.GetDigest())
}

func (c *codeManager) lookupMemCache(name string, desc *pb.WasmCodeDesc) (*contractCode, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ccode, ok := c.codes[name]
	if !ok {
		return nil, false
	}
	if codeDescEqual(&ccode.Desc, desc) {
		return ccode, true
	}
	return nil, false
}

func (c *codeManager) purgeMemCache(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if ccode, ok := c.codes[name]; ok {
		ccode.ExecCode.Release()
	}
	delete(c.codes, name)
}

func (c *codeManager) makeMemCache(name, libpath string, desc *pb.WasmCodeDesc) (*contractCode, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.codes[name]; ok {
		return nil, errors.New("old contract code not purged")
	}

	execCode, err := c.makeExecCode(libpath)
	if err != nil {
		return nil, err
	}
	code := &contractCode{
		ContractName: name,
		ExecCode:     execCode,
		Desc:         *desc,
	}
	c.codes[name] = code

	return code, nil
}

func fileExists(fpath string) bool {
	stat, err := os.Stat(fpath)
	if err == nil && !stat.IsDir() {
		return true
	}
	return false
}

func (c *codeManager) lookupDiskCache(name string, desc *pb.WasmCodeDesc) (string, bool) {
	descpath := filepath.Join(c.basedir, name, "code.desc")
	libpath := filepath.Join(c.basedir, name, "code.so")
	if !fileExists(descpath) || !fileExists(libpath) {
		return "", false
	}
	var localDesc pb.WasmCodeDesc
	descbuf, err := ioutil.ReadFile(descpath)
	if err != nil {
		return "", false
	}
	err = json.Unmarshal(descbuf, &localDesc)
	if err != nil {
		return "", false
	}
	if !codeDescEqual(&localDesc, desc) ||
		localDesc.GetVmCompiler() != compile.Version {
		return "", false
	}
	return libpath, true
}

func (c *codeManager) makeDiskCache(name string, desc *pb.WasmCodeDesc, codebuf []byte) (string, error) {
	basedir := filepath.Join(c.basedir, name)
	descpath := filepath.Join(basedir, "code.desc")
	libpath := filepath.Join(basedir, "code.so")

	err := os.MkdirAll(basedir, 0700)
	if err != nil {
		return "", err
	}

	err = c.compileCode(codebuf, libpath)
	if err != nil {
		return "", err
	}
	localDesc := *desc
	localDesc.VmCompiler = compile.Version
	descbuf, _ := json.Marshal(&localDesc)
	err = ioutil.WriteFile(descpath, descbuf, 0600)
	if err != nil {
		os.RemoveAll(basedir)
		return "", err
	}
	return libpath, nil
}

func (c *codeManager) GetExecCode(name string, cp vm.ContractCodeProvider) (*contractCode, error) {
	desc, err := cp.GetContractCodeDesc(name)
	if err != nil {
		return nil, err
	}
	execCode, ok := c.lookupMemCache(name, desc)
	if ok {
		log.Debug("contract code hit memory cache", "contract", name)
		return execCode, nil
	}

	// Only allow one goroutine make disk and memory cache at given contract name
	// other goroutine will block on the same contract name.
	icode, err, _ := c.makeCacheLock.Do(name, func() (interface{}, error) {
		defer c.makeCacheLock.Forget(name)
		// 对于pending在Do上的goroutine在Do返回后能获取到最新的memory cache
		// 但由于我们在Do完之后立马Forget，因此如果在第一个goroutine在调用Do期间,
		// 另外一个goroutine刚好处在loopupMemCache失败之后和Do之前，这样就不能看到最新的cache，
		// 会重复执行，清理掉正在使用的对象从而造成错误。
		// 这里进行double check来发现最新的cache
		execCode, ok := c.lookupMemCache(name, desc)
		if ok {
			return execCode, nil
		}
		// old code handle should be closed before open new code
		// see https://github.com/xuperchain/xuperunion/issues/352
		c.purgeMemCache(name)
		libpath, ok := c.lookupDiskCache(name, desc)
		if !ok {
			log.Debug("contract code need make disk cache", "contract", name)
			codebuf, err := cp.GetContractCode(name)
			if err != nil {
				return nil, err
			}
			libpath, err = c.makeDiskCache(name, desc, codebuf)
			if err != nil {
				return nil, err
			}
		} else {
			log.Debug("contract code hit disk cache", "contract", name)
		}
		return c.makeMemCache(name, libpath, desc)
	})
	if err != nil {
		return nil, err
	}
	return icode.(*contractCode), nil
}

func (c *codeManager) RemoveCode(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	code, ok := c.codes[name]
	if ok {
		code.ExecCode.Release()
	}
	delete(c.codes, name)
	os.RemoveAll(filepath.Join(c.basedir, name))
}
