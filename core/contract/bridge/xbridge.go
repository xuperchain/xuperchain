package bridge

import (
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/xuperchain/xuperchain/core/common/config"
	"github.com/xuperchain/xuperchain/core/common/log"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/xmodel"

	log15 "github.com/xuperchain/log15"
)

// XBridge 用于注册用户虚拟机以及向Xchain Core注册可被识别的vm.VirtualMachine
type XBridge struct {
	ctxmgr         *ContextManager
	syscallService *SyscallService
	basedir        string
	vmconfigs      map[ContractType]VMConfig
	creators       map[ContractType]InstanceCreator
	vms            map[string]contract.VirtualMachine
	xmodel         xmodel.XMReader
	config         config.ContractConfig

	debugLogger *log.Logger

	*contractManager
}

type XBridgeConfig struct {
	Basedir   string
	VMConfigs map[ContractType]VMConfig
	XModel    xmodel.XMReader
	Config    config.ContractConfig
	LogWriter io.Writer
}

// New instances a new XBridge
func New(cfg *XBridgeConfig) (*XBridge, error) {
	ctxmgr := NewContextManager()
	xbridge := &XBridge{
		ctxmgr:    ctxmgr,
		basedir:   cfg.Basedir,
		vmconfigs: cfg.VMConfigs,
		creators:  make(map[ContractType]InstanceCreator),
		vms:       make(map[string]contract.VirtualMachine),
		xmodel:    cfg.XModel,
		config:    cfg.Config,
	}
	xbridge.contractManager = &contractManager{
		xbridge:      xbridge,
		codeProvider: newCodeProvider(cfg.XModel),
	}

	syscallService := NewSyscallService(ctxmgr, xbridge)
	xbridge.syscallService = syscallService
	err := xbridge.initVM()
	if err != nil {
		return nil, err
	}
	err = xbridge.initDebugLogger(cfg)
	if err != nil {
		return nil, err
	}
	return xbridge, nil
}

func (v *XBridge) initVM() error {
	types := []ContractType{TypeWasm, TypeNative, TypeEvm}
	for _, tp := range types {
		vmconfig, ok := v.vmconfigs[tp]
		if !ok {
			log.Error("config for contract type not found", "type", tp)
			continue
		}
		if !vmconfig.IsEnable() {
			log.Info("contract type disabled", "type", tp)
			continue
		}
		creatorConfig := &InstanceCreatorConfig{
			Basedir:        filepath.Join(v.basedir, vmconfig.DriverName()),
			SyscallService: v.syscallService,
			VMConfig:       vmconfig,
		}
		creator, err := Open(tp, vmconfig.DriverName(), creatorConfig)
		if err != nil {
			return err
		}
		vm := &vmImpl{
			ctxmgr:       v.ctxmgr,
			xbridge:      v,
			name:         string(tp),
			codeProvider: newCodeProvider(v.xmodel),
		}
		v.creators[tp] = creator
		v.vms[string(tp)] = vm
	}
	return nil
}

func (v *XBridge) initDebugLogger(cfg *XBridgeConfig) error {
	// 如果日志开启，并且没有自定义writter则使用配置文件打开日志对象
	if cfg.Config.EnableDebugLog && cfg.LogWriter == nil {
		debugLogger, err := log.OpenLog(&cfg.Config.DebugLog)
		if err != nil {
			return err
		}
		v.debugLogger = &debugLogger
		return nil
	}

	w := cfg.LogWriter
	if w == nil {
		w = ioutil.Discard
	}
	logger := log15.Root().New()
	logger.SetHandler(log15.StreamHandler(w, log15.LogfmtFormat()))
	v.debugLogger = &log.Logger{Logger: logger}
	return nil
}

func (v *XBridge) getCreator(tp ContractType) InstanceCreator {
	return v.creators[tp]
}

// GetVirtualMachine returns a contract.VirtualMachine from the given name
func (v *XBridge) GetVirtualMachine(name string) (contract.VirtualMachine, bool) {
	vm, ok := v.vms[name]
	return vm, ok
}

// RegisterToXCore register VirtualMachines to xchain core
func (v *XBridge) RegisterToXCore(regfunc func(name string, vm contract.VirtualMachine) error) {
	for _, vm := range v.vms {
		err := regfunc(vm.GetName(), vm)
		if err != nil {
			panic(err)
		}
	}
}
