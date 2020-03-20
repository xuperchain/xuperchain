package wasm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"

	log15 "github.com/xuperchain/log15"

	"github.com/golang/protobuf/proto"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/wasm/vm"
	"github.com/xuperchain/xuperchain/core/crypto/hash"

	"github.com/xuperchain/xuperchain/core/pluginmgr"

	// import xvm wasm virtual machine
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	_ "github.com/xuperchain/xuperchain/core/contract/wasm/vm/xvm"
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/xmodel"
)

// VMManager manages wasm contracts, include deploy contracts, instance wasm virtual machine, etc...
type VMManager struct {
	basedir      string
	config       *config.WasmConfig
	xmodel       xmodel.XMReader
	syscall      *bridge.SyscallService
	vmimpl       vm.InstanceCreator
	xbridge      *bridge.XBridge
	codeProvider vm.ContractCodeProvider
	debugLogger  *log.Logger
}

// New instances a new VMManager
func New(cfg *config.WasmConfig, basedir string, xbridge *bridge.XBridge, xmodel xmodel.XMReader) (*VMManager, error) {
	vmm := &VMManager{
		basedir:      basedir,
		config:       cfg,
		xmodel:       xmodel,
		xbridge:      xbridge,
		codeProvider: newCodeProvider(xmodel),
	}

	if cfg.External {
		pluginMgr, err := pluginmgr.GetPluginMgr()
		if err != nil {
			return nil, err
		}
		if _, err = pluginMgr.PluginMgr.CreatePluginInstance("wasm", cfg.Driver); err != nil {
			return nil, err
		}
	}

	if cfg.EnableDebugLog {
		debugLogger, err := log.OpenLog(&cfg.DebugLog)
		if err != nil {
			return nil, err
		}
		vmm.debugLogger = &debugLogger
	} else {
		logger := log15.Root().New()
		logger.SetHandler(log15.DiscardHandler())
		vmm.debugLogger = &log.Logger{Logger: logger}
	}

	return vmm, nil
}

func (v *VMManager) getVMConfig(name string) (interface{}, error) {
	configv := reflect.ValueOf(v.config).Elem()
	value := configv.FieldByNameFunc(func(field string) bool {
		return name == strings.ToLower(field)
	})
	if value.IsValid() && value.Type().Kind() == reflect.Struct {
		return value.Interface(), nil
	}
	return nil, fmt.Errorf("config for %s not found", name)
}

// RegisterSyscallService implements bridge.Executor
func (v *VMManager) RegisterSyscallService(syscall *bridge.SyscallService) {
	v.syscall = syscall
	vmconfig, _ := v.getVMConfig(v.config.Driver)
	vmimpl, err := vm.Open(v.config.Driver, &vm.InstanceCreatorConfig{
		Basedir:        filepath.Join(v.basedir, v.config.Driver),
		SyscallService: syscall,
		VMConfig:       vmconfig,
		DebugLogger:    v.debugLogger,
		TEEConfig:      v.config.TEEConfig,
	})
	if err != nil {
		panic(err)
	}
	v.vmimpl = vmimpl
}

func ContractCodeDescKey(contractName string) []byte {
	return []byte(contractName + "." + "desc")
}

func contractCodeKey(contractName string) []byte {
	return []byte(contractName + "." + "code")
}

// NewInstance implements bridge.Executor
func (v *VMManager) NewInstance(ctx *bridge.Context) (bridge.Instance, error) {
	desc, err := newCodeProvider(ctx.Cache).GetContractCodeDesc(ctx.ContractName)
	if err != nil {
		return nil, err
	}
	cp := newDescProvider(v.codeProvider, desc)
	ins, err := v.vmimpl.CreateInstance(ctx, cp)
	if err != nil {
		return nil, err
	}
	return &bridgeInstance{
		ctx:        ctx,
		vmInstance: ins,
		codeDesc:   desc,
	}, nil
}

// TODO:校验名字
func (v *VMManager) verifyContractName(name string) error {
	return nil
}

// DeployContract deploy contract and initialize contract
func (v *VMManager) DeployContract(contextConfig *contract.ContextConfig, args map[string][]byte) (*contract.Response, contract.Limits, error) {
	store := contextConfig.XMCache
	name := args["contract_name"]
	if name == nil {
		return nil, contract.Limits{}, errors.New("bad contract name")
	}
	contractName := string(name)
	err := v.verifyContractName(contractName)
	if err != nil {
		return nil, contract.Limits{}, err
	}
	_, err = v.codeProvider.GetContractCodeDesc(contractName)
	if err == nil {
		return nil, contract.Limits{}, fmt.Errorf("contract %s already exists", contractName)
	}

	code := args["contract_code"]
	if code == nil {
		return nil, contract.Limits{}, errors.New("missing contract code")
	}
	initArgsBuf := args["init_args"]
	if initArgsBuf == nil {
		return nil, contract.Limits{}, errors.New("missing args field in args")
	}
	var initArgs map[string][]byte
	err = json.Unmarshal(initArgsBuf, &initArgs)
	if err != nil {
		return nil, contract.Limits{}, err
	}

	descbuf := args["contract_desc"]
	var desc pb.WasmCodeDesc
	err = proto.Unmarshal(descbuf, &desc)
	if err != nil {
		return nil, contract.Limits{}, err
	}
	desc.Digest = hash.DoubleSha256(code)
	descbuf, _ = proto.Marshal(&desc)

	store.Put("contract", ContractCodeDescKey(contractName), descbuf)
	store.Put("contract", contractCodeKey(contractName), code)
	// 由于部署合约的时候代码还没有持久化，构造一个从ModelCache获取代码的对象
	// 在执行init函数的时候，代码已经进入vm cache，因此使用VMManager的默认CodeProvider没有问题
	// FIXME: 确保InstanceCreator缓存了已经编译的代码
	cp := newCodeProvider(store)
	instance, err := v.vmimpl.CreateInstance(&bridge.Context{
		ContractName:   contractName,
		ResourceLimits: contextConfig.ResourceLimits,
	}, cp)
	if err != nil {
		v.vmimpl.RemoveCache(contractName)
		log.Error("create contract instance error when deploy contract", "error", err, "contract", contractName)
		return nil, contract.Limits{}, err
	}
	instance.Release()

	initConfig := *contextConfig
	initConfig.ContractName = contractName
	initConfig.CanInitialize = true
	out, resourceUsed, err := v.initContract(&initConfig, initArgs)
	if err != nil {
		if _, ok := err.(*bridge.ContractError); !ok {
			v.vmimpl.RemoveCache(contractName)
		}
		log.Error("call contract initialize method error", "error", err, "contract", contractName)
		return nil, contract.Limits{}, err
	}
	return out, resourceUsed, nil
}

func (v *VMManager) initContract(contextConfig *contract.ContextConfig, args map[string][]byte) (*contract.Response, contract.Limits, error) {
	vm, ok := v.xbridge.GetVirtualMachine("wasm")
	if !ok {
		return nil, contract.Limits{}, errors.New("wasm vm not registered")
	}

	ctx, err := vm.NewContext(contextConfig)
	if err != nil {
		return nil, contract.Limits{}, err
	}
	out, err := ctx.Invoke("initialize", args)
	if err != nil {
		return nil, contract.Limits{}, err
	}
	return out, ctx.ResourceUsed(), nil
}

// UpgradeContract deploy contract and initialize contract
func (v *VMManager) UpgradeContract(contextConfig *contract.ContextConfig, args map[string][]byte) (*contract.Response, contract.Limits, error) {
	if !v.config.EnableUpgrade {
		return nil, contract.Limits{}, errors.New("contract upgrade disabled")
	}

	name := args["contract_name"]
	if name == nil {
		return nil, contract.Limits{}, errors.New("bad contract name")
	}
	contractName := string(name)
	err := v.verifyContractName(contractName)
	if err != nil {
		return nil, contract.Limits{}, err
	}
	desc, err := v.codeProvider.GetContractCodeDesc(contractName)
	if err != nil {
		return nil, contract.Limits{}, fmt.Errorf("contract %s not exists", contractName)
	}

	code := args["contract_code"]
	if code == nil {
		return nil, contract.Limits{}, errors.New("missing contract code")
	}
	desc.Digest = hash.DoubleSha256(code)
	descbuf, _ := proto.Marshal(desc)

	store := contextConfig.XMCache
	store.Put("contract", ContractCodeDescKey(contractName), descbuf)
	store.Put("contract", contractCodeKey(contractName), code)

	cp := newCodeProvider(store)
	instance, err := v.vmimpl.CreateInstance(&bridge.Context{
		ContractName:   contractName,
		ResourceLimits: contract.MaxLimits,
	}, cp)
	if err != nil {
		log.Error("create contract instance error when upgrade contract", "error", err, "contract", contractName)
		return nil, contract.Limits{}, err
	}
	instance.Release()

	return &contract.Response{
			Status: 200,
			Body:   []byte("upgrade success"),
		}, contract.Limits{
			Disk: modelCacheDiskUsed(store),
		}, nil
}

// SetLogOutput set the output of contract log
func (v *VMManager) SetLogOutput(w io.Writer) {
	v.debugLogger.SetHandler(log15.StreamHandler(w, log15.LogfmtFormat()))
}

func modelCacheDiskUsed(cache *xmodel.XMCache) int64 {
	size := int64(0)
	_, wset, _ := cache.GetRWSets()
	for _, w := range wset {
		size += int64(len(w.GetKey()))
		size += int64(len(w.GetValue()))
	}
	return size
}
