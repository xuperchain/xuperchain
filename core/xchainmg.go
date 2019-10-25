package xchaincore

import (
	"errors"
	"io/ioutil"
	"sync"

	"github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	"github.com/xuperchain/xuperunion/common/events"
	"github.com/xuperchain/xuperunion/common/probe"
	"github.com/xuperchain/xuperunion/contract/kernel"
	"github.com/xuperchain/xuperunion/p2pv2"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
	pm "github.com/xuperchain/xuperunion/pluginmgr"
)

// XChainMG manage all chains
type XChainMG struct {
	Log   log.Logger
	Cfg   *config.NodeConfig
	P2pv2 p2pv2.P2PServer
	// msgChan is the message subscribe from net
	msgChan    chan *xuper_p2p.XuperMessage
	chains     *sync.Map
	rootKernel *kernel.Kernel
	datapath   string
	Ukeys      *sync.Map //address -> scrkey
	Speed      *probe.SpeedCalc
	Quit       chan struct{}
	nodeMode   string
	// the switch of compressed
	enableCompress bool
}

// Init init instance of XChainMG
func (xm *XChainMG) Init(log log.Logger, cfg *config.NodeConfig,
	p2pV2 p2pv2.P2PServer) error {
	xm.Log = log
	xm.chains = new(sync.Map)
	xm.datapath = cfg.Datapath
	xm.Cfg = cfg
	xm.P2pv2 = p2pV2
	xm.msgChan = make(chan *xuper_p2p.XuperMessage, p2pv2.MsgChanSize)

	xm.Speed = probe.NewSpeedCalc("sum")
	xm.Quit = make(chan struct{})
	xm.nodeMode = cfg.NodeMode
	xm.enableCompress = cfg.EnableCompress

	// auto-load plugins here
	if err := pm.Init(cfg); err != nil {
		xm.Log.Error("can't initialize plugin manager", "error", err)
		return err
	}

	dir, err := ioutil.ReadDir(xm.datapath)
	if err != nil {
		xm.Log.Error("can't open data", "datapath", xm.datapath)
		return err
	}
	for _, fi := range dir {
		if fi.IsDir() { // 忽略非目录
			xm.Log.Trace("--------find " + fi.Name())
			aKernel := &kernel.Kernel{}
			aKernel.Init(xm.datapath, xm.Log, xm, fi.Name())
			x := &XChainCore{}
			err := x.Init(fi.Name(), log, cfg, p2pV2, aKernel, xm.nodeMode)
			if err != nil {
				return err
			}
			if fi.Name() == "xuper" {
				xm.rootKernel = aKernel
			}
			xm.chains.Store(fi.Name(), x)
		}
	}
	if xm.rootKernel == nil {
		err := errors.New("xuper chain not found")
		xm.Log.Error("can not find xuper chain, please create it first", "err", err)
		return err
	}
	xm.rootKernel.SetNewChainWhiteList(cfg.Kernel.NewChainWhiteList)
	xm.rootKernel.SetMinNewChainAmount(cfg.Kernel.MinNewChainAmount)
	/*for _, x := range xm.chains {
		go x.SyncBlocks()
	}*/
	if err := xm.RegisterSubscriber(); err != nil {
		return err
	}
	go xm.Speed.ShowLoop(xm.Log)
	xm.notifyInitialized()
	return nil
}

// Get return specific instance of blockchain by blockchain name from map
func (xm *XChainMG) Get(name string) *XChainCore {
	v, ok := xm.chains.Load(name)
	if ok {
		xc := v.(*XChainCore)
		return xc
	}
	return nil
}

// Set put <blockname, blockchain instance> into map
func (xm *XChainMG) Set(name string, xc *XChainCore) {
	xm.chains.Store(name, xc)
}

// GetAll returns all blockchains name
func (xm *XChainMG) GetAll() []string {
	var bcs []string
	xm.chains.Range(func(k, v interface{}) bool {
		xc := v.(*XChainCore)
		bcs = append(bcs, xc.bcname)
		return true
	})
	return bcs
}

// Start start all blockchain instances
func (xm *XChainMG) Start() {
	xm.chains.Range(func(k, v interface{}) bool {
		xc := v.(*XChainCore)
		xm.Log.Trace("start chain " + k.(string))
		go xc.Miner()
		return true
	})
	go xm.StartLoop()
}

// Stop stop all blockchain instances
func (xm *XChainMG) Stop() {
	xm.notifyStopping()
	xm.chains.Range(func(k, v interface{}) bool {
		xc := v.(*XChainCore)
		xm.Log.Trace("stop chain " + k.(string))
		xc.Stop()
		return true
	})
	if xm.P2pv2 != nil {
		xm.P2pv2.Stop()
	}
}

// CreateBlockChain create an instance of blockchain
func (xm *XChainMG) CreateBlockChain(name string, data []byte) (*XChainCore, error) {
	if _, ok := xm.chains.Load(name); ok {
		xm.Log.Warn("chains[" + name + "] is exist")
		return nil, ErrBlockChainIsExist
	}

	if err := xm.rootKernel.CreateBlockChain(name, data); err != nil {
		return nil, err
	}
	return xm.addBlockChain(name)
}

func (xm *XChainMG) addBlockChain(name string) (*XChainCore, error) {
	x := &XChainCore{}
	aKernel := xm.rootKernel
	if name != "xuper" {
		aKernel = &kernel.Kernel{}
		aKernel.Init(xm.datapath, xm.Log, xm, name)
	}
	err := x.Init(name, xm.Log, xm.Cfg, xm.P2pv2, aKernel, xm.nodeMode)
	if err != nil {
		xm.Log.Warn("XChainCore init error")
		xm.rootKernel.RemoveBlockChainData(name)
		return nil, err
	}
	xm.Set(name, x)
	return x, nil
}

// RegisterBlockChain load an instance of blockchain and start it dynamically
func (xm *XChainMG) RegisterBlockChain(name string) error {
	xc, err := xm.addBlockChain(name)
	if err != nil {
		return err
	}
	go xc.Miner()
	return err
}

// UnloadBlockChain unload an instance of blockchain and stop it dynamically
func (xm *XChainMG) UnloadBlockChain(name string) error {
	v, ok := xm.chains.Load(name)
	if !ok {
		return ErrBlockChainIsExist
	}
	xm.chains.Delete(name) //从xchainmg的map里面删了，就不会收到新的请求了
	//然后停止这个链
	xc := v.(*XChainCore)
	xc.Stop()
	return nil
}

// GetXchainmgConfig GetXchainMg CFG
func (xm *XChainMG) GetXchainmgConfig() *config.NodeConfig {
	return xm.Cfg
}

func (xm *XChainMG) notifyInitialized() {
	em := &events.EventMessage{
		BcName:   "",
		Type:     events.SystemInitialized,
		Priority: 0,
		Sender:   xm,
		Message:  "System Started",
	}
	_, err := events.GetEventBus().FireEventAsync(em)
	if err != nil {
		xm.Log.Warn("xchainmg notifyInitialized failed", "error", err)
	}
}

func (xm *XChainMG) notifyStopping() {
	em := &events.EventMessage{
		BcName:   "",
		Type:     events.SystemStopping,
		Priority: 0,
		Sender:   xm,
		Message:  "System Stopping",
	}
	_, err := events.GetEventBus().FireEventAsync(em)
	if err != nil {
		xm.Log.Warn("xchainmg notifyStopping failed", "error", err)
	}
}
