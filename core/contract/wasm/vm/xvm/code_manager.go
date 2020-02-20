package xvm

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract/wasm/vm"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xvm/compile"
	"github.com/xuperchain/xuperchain/core/xvm/exec"
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
	rundir       string
	cachedir     string
	compileCode  compileFunc
	makeExecCode makeExecCodeFunc

	makeCacheLock singleflight.Group

	mutex sync.Mutex // protect codes
	codes map[string]*contractCode
}

func newCodeManager(basedir string, compile compileFunc, makeExec makeExecCodeFunc) (*codeManager, error) {
	runDirFull := filepath.Join(basedir, "var", "run")
	cacheDirFull := filepath.Join(basedir, "var", "cache")
	if err := os.MkdirAll(runDirFull, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cacheDirFull, 0755); err != nil {
		return nil, err
	}

	return &codeManager{
		basedir:      basedir,
		rundir:       runDirFull,
		cachedir:     cacheDirFull,
		compileCode:  compile,
		makeExecCode: makeExec,
		codes:        make(map[string]*contractCode),
	}, nil
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

func (c *codeManager) makeMemCache(name, libpath string, desc *pb.WasmCodeDesc) (*contractCode, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// 创建临时文件，这样每个合约版本独享一个so文件，不会相互影响
	tmpfile := fmt.Sprintf("%s-%d-%d.so", name, time.Now().UnixNano(), rand.Int()%10000)
	libpathFull := filepath.Join(c.rundir, tmpfile)
	err := cpfile(libpathFull, libpath)
	if err != nil {
		return nil, err
	}

	execCode, err := c.makeExecCode(libpathFull)
	if err != nil {
		return nil, err
	}
	code := &contractCode{
		ContractName: name,
		ExecCode:     execCode,
		Desc:         *desc,
	}
	runtime.SetFinalizer(code, func(c *contractCode) {
		c.ExecCode.Release()
	})
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
	delete(c.codes, name)
	os.RemoveAll(filepath.Join(c.basedir, name))
}

// not used now
func makeCacheId(desc *pb.WasmCodeDesc) string {
	h := sha1.New()
	h.Write(desc.GetDigest())
	h.Write([]byte(compile.Version))
	return hex.EncodeToString(h.Sum(nil))
}
