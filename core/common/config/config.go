package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

// default settings
const (
	// NodeModeNormal NODE_MODE_NORMAL node mode for normal
	NodeModeNormal = "Normal"
	// NodeModeFastSync NODE_MODE_FAST_SYNC node mode for fast
	NodeModeFastSync        = "FastSync"
	DefaultNetPort          = 47101            // p2p port
	DefaultNetKeyPath       = "./data/netkeys" // node private key path
	DefaultCertPath         = "./data/cert"
	DefaultNetIsNat         = true  // use NAT
	DefaultNetIsTls         = false // use tls secure transport
	DefaultNetIsHidden      = false
	DefaultNetIsIpv6        = false
	DefaultMaxStreamLimits  = 1024
	DefaultMaxMessageSize   = 128
	DefaultTimeout          = 3
	DefaultIsAuthentication = false
	DefautltAuthTimeout     = 30
	// limitation size for same ip
	DefaultStreamIPLimitSize     = 10
	DefaultMaxBroadcastPeers     = 20
	DefaultMaxBroadcastCorePeers = 17
	DefaultIsStorePeers          = false
	DefaultP2PDataPath           = "./data/p2p"
	DefaultP2PModuleName         = "p2pv2"
	DefaultServiceName           = ""
	DefaultIsBroadCast           = true
)

// LogConfig is the log config of node
type LogConfig struct {
	Module         string `yaml:"module,omitempty"`
	Filepath       string `yaml:"filepath,omitempty"`
	Filename       string `yaml:"filename,omitempty"`
	Fmt            string `yaml:"fmt,omitempty"`
	Console        bool   `yaml:"console,omitempty"`
	Level          string `yaml:"level,omitempty"`
	Async          bool   `yaml:"async,omitempty"`
	RotateInterval int    `yaml:"rotateinterval,omitempty"`
	RotateBackups  int    `yaml:"rotatebackups,omitempty"`
}

// TCPServerConfig is the tcp port of the node
type TCPServerConfig struct {
	Port                  string `yaml:"port,omitempty"`
	HTTPSPort             string `yaml:"httpsPort,omitempty"`
	MetricPort            string `yaml:"metricPort,omitempty"`
	MaxMsgSize            int    `yaml:"maxmsgsize,omitempty"`
	TLS                   bool   `yaml:"tls,omitempty"`
	TLSPath               string `yaml:"tlsPath,omitempty"`
	MServerURL            string `yaml:"mServerUrl,omitempty"`
	MServerName           string `yaml:"mServerName,omitempty"`
	CachePeriod           int64  `yaml:"cachePeriod,omitempty"`
	InitialWindowSize     int32  `yaml:"initialWindowSize,omitempty"`
	InitialConnWindowSize int32  `yaml:"initialConnWindowSize"`
	ReadBufferSize        int    `yaml:"readBufferSize"`
	WriteBufferSize       int    `yaml:"writeBufferSize"`
}

// P2PConfig is the config of xuper p2p server. Attention, config of dht are not expose
type P2PConfig struct {
	// Module is the name of p2p module plugin
	Module string `yaml:"module,omitempty"`
	// port the p2p network listened
	Port int32 `yaml:"port,omitempty"`
	// keyPath is the node private key path, xuper will gen a random one if is nil
	KeyPath string `yaml:"keyPath,omitempty"`
	// isNat config whether the node use NAT manager
	IsNat bool `yaml:"isNat,omitempty"`
	// isTls config the node use tls secure transparent
	IsTls bool `yaml:"isTls,omitempty"`
	// isHidden config whether the node can be found
	IsHidden bool `yaml:"isHidden,omitempty"`
	// IsIpv6 config whether the node use ipv6
	IsIpv6 bool `yaml:"isIpv6,omitempty"`
	// bootNodes config the bootNodes the node to connect
	BootNodes []string `yaml:"bootNodes,omitempty"`
	// staticNodes config the nodes which you trust
	StaticNodes map[string][]string `yaml:"staticNodes,omitempty"`
	// isBroadCast config whether broadcast to all StaticNodes
	IsBroadCast bool `yaml:"isBroadCast,omitempty"`
	// maxStreamLimits config the max stream num
	MaxStreamLimits int32 `yaml:"maxStreamLimits,omitempty"`
	// maxMessageSize config the max message size
	MaxMessageSize int64 `yaml:"maxMessageSize,omitempty"`
	// timeout config the timeout of Request with response
	Timeout int64 `yaml:"timeout,omitempty"`
	// IsAuthentication determine whether peerID and Xchain addr correspond
	IsAuthentication bool `yaml:"isauthentication,omitempty"`
	// StreamIPLimitSize set the limitation size for same ip
	StreamIPLimitSize int64 `yaml:"streamIPLimitSize,omitempty"`
	// MaxBroadcastPeers limit the number of common peers in a broadcast,
	// this number do not include MaxBroadcastCorePeers.
	MaxBroadcastPeers int `yaml:"maxBroadcastPeers,omitempty"`
	// MaxBroadcastCorePeers limit the number of core peers in a broadcast,
	// this only works when NodeConfig.CoreConnection is true. Note that the number
	// of core peers is included in MaxBroadcastPeers.
	MaxBroadcastCorePeers int `yaml:"maxBroadcastCorePeers,omitempty"`
	// P2PDataPath stores the peer info connected last time
	P2PDataPath string `yaml:"p2PDataPath,omitempty"`
	// IsStorePeers determine wherther storing the peers infos
	IsStorePeers bool `yaml:"isStorePeers,omitempty"`
	// CertPath define the path of certificate
	CertPath string `yaml:"certPath,omitempty"`
	// IsUseCert define whether to use certificate, default true
	IsUseCert bool `yaml:"isUseCert,omitempty"`
	// ServiceName
	ServiceName string `yaml:"serviceName,omitempty"`
}

// MinerConfig is the config of miner
type MinerConfig struct {
	Keypath string `yaml:"keypath,omitempty"`
}

// UtxoConfig is the config of UtxoVM
type UtxoConfig struct {
	CacheSize             int                        `yaml:"cachesize,omitempty"`
	TmpLockSeconds        int                        `yaml:"tmplockSeconds,omitempty"`
	AsyncMode             bool                       `yaml:"asyncMode,omitempty"`
	AsyncBlockMode        bool                       `yaml:"asyncBlockMode,omitempty"`
	ContractExecutionTime int                        `yaml:"contractExecutionTime,omitempty"`
	ContractWhiteList     map[string]map[string]bool `yaml:"contractWhiteList,omitempty"`
	// 是否开启新版本tx k = bcname, v = isBetaTx
	IsBetaTx          map[string]bool `yaml:"isBetaTx,omitempty"`
	MaxConfirmedDelay uint32          `yaml:"maxConfirmedDelay,omitempty"`
}

// NativeDockerConfig native contract use docker config
type NativeDockerConfig struct {
	Enable    bool
	ImageName string
	Cpus      float32
	Memory    string
}

// NativeConfig contains the two above config
type NativeConfig struct {
	Driver string
	// Timeout (in seconds) to stop native code process
	StopTimeout int
	Docker      NativeDockerConfig
	Enable      bool
}

func (n *NativeConfig) DriverName() string {
	if n.Driver != "" {
		return n.Driver
	}

	return "native"
}

func (n *NativeConfig) IsEnable() bool {
	return n.Enable
}

// XVMConfig contains the xvm configuration
type XVMConfig struct {
	// From 0 to 3
	// The higher the number, the faster the program runs,
	// but the compilation speed will be slower
	OptLevel int `yaml:"optlevel"`
}

// WasmConfig wasm config
type WasmConfig struct {
	Driver        string
	External      bool
	XVM           XVMConfig
	EnableUpgrade bool
	TEEConfig     TEEConfig `yaml:"teeConfig,omitempty"`
}

func (w *WasmConfig) DriverName() string {
	return w.Driver
}

func (w *WasmConfig) IsEnable() bool {
	return true
}

// ContractConfig define the config of XuperBridge
type ContractConfig struct {
	EnableDebugLog bool
	DebugLog       LogConfig
	EnableUpgrade  bool
}

// TEEConfig sets up the private ledger
type TEEConfig struct {
	Enable     bool   `yaml:"enable"`     // enable: on or off to enable private ledger
	PluginPath string `yaml:"pluginPath"` // path to dynamic library
	ConfigPath string `yaml:"configPath"` // config path for the dynamic
}

type EVMConfig struct {
	Enable bool
	Driver string
}

func (e *EVMConfig) DriverName() string {
	return e.Driver
}

func (e *EVMConfig) IsEnable() bool {
	return e.Enable
}

type CloudStorageConfig struct {
	Bucket        string `yaml:"bucket"`        //bucket name of s3 or bos
	Path          string `yaml:"path"`          //path in the bucket
	Ak            string `yaml:"ak"`            //access key
	Sk            string `yaml:"sk"`            //secrete key
	Region        string `yaml:"region"`        //region, eg. bj
	Endpoint      string `yaml:"endpoint"`      //endpoint, eg. s3.bj.bcebos.com
	LocalCacheDir string `yaml:"localCacheDir"` //cache directory on local disk
}

func (w *WasmConfig) applyFlags(flags *pflag.FlagSet) {
	flags.StringVar(&w.Driver, "vm", w.Driver, "contract vm driver")
}

// ConsoleConfig is the command config user input
type ConsoleConfig struct {
	Keys       string
	Name       string
	Host       string
	MaxMsgSize int
}

// ApplyFlags apply flag to console command
func (cmd *ConsoleConfig) ApplyFlags(flags *pflag.FlagSet) {
}

// NodeConfig is the main config of the xchain node
type NodeConfig struct {
	Version         string          `yaml:"version,omitempty"`
	Log             LogConfig       `yaml:"log,omitempty"`
	TCPServer       TCPServerConfig `yaml:"tcpServer,omitempty"`
	P2p             P2PConfig       `yaml:"p2pV2,omitempty"`
	Miner           MinerConfig     `yaml:"miner,omitempty"`
	Datapath        string          `yaml:"datapath,omitempty"`
	DatapathOthers  []string        `yaml:"datapathOthers,omitempty"` //扩展盘的路径
	ConsoleConfig   ConsoleConfig
	Utxo            UtxoConfig      `yaml:"utxo,omitempty"`
	DedupCacheSize  int             `yaml:"dedupCacheSize,omitempty"`
	DedupTimeLimit  int             `yaml:"dedupTimeLimit,omitempty"`
	Kernel          KernelConfig    `yaml:"kernel,omitempty"`
	CPUProfile      string          `yaml:"cpuprofile,omitempty"`
	MemProfile      string          `yaml:"memprofile,omitempty"`
	MemberWhiteList map[string]bool `yaml:"memberWhiteList,omitempty"`

	// 合约相关配置
	Contract ContractConfig `yaml:"contract,omitempty"`
	Native   NativeConfig   `yaml:"native,omitempty"`
	Wasm     WasmConfig     `yaml:"wasm,omitempty"`
	EVM      EVMConfig      `yaml:"evm,omitempty"`

	DBCache DBCacheConfig `yaml:"dbcache,omitempty"`
	// 节点模式: NORMAL | FAST_SYNC 两种模式
	// NORMAL: 为普通的全节点模式
	// FAST_SYNC 模式下:节点需要连接一个可信的全节点; 拒绝事务提交; 同步区块时跳过块验证和tx验证; 去掉load未确认事务;
	NodeMode        string `yaml:"nodeMode,omitempty"`
	PluginConfPath  string `yaml:"pluginConfPath,omitempty"` // plugin config file path
	PluginLoadPath  string `yaml:"pluginLoadPath,omitempty"` // plugin auto-load path
	EtcdClusterAddr string `yaml:"etcdClusterAddr,omitempty"`
	GatewaySwitch   bool   `yaml:"gatewaySwitch,omitempty"`
	CoreConnection  bool   `yaml:"coreConnection,omitempty"`
	FailSkip        bool   `yaml:"failSkip,omitempty"`
	// XEndorser the endorser module config
	XEndorser XEndorserConfig `yaml:"xendorser,omitempty"`
	// TxCacheExpiredTime expired time for tx cache
	TxidCacheExpiredTime time.Duration `yaml:"txidCacheExpiredTime,omitempty"`
	// local switch of compressed
	EnableCompress bool `yaml:"enableCompress,omitempty"`
	// prune ledger option
	Prune PruneOption `yaml:"prune,omitempty"`

	// BlockBroadcaseMode is the mode for broadcast new block
	//  * Full_BroadCast_Mode = 0, means send full block data
	//  * Interactive_BroadCast_Mode = 1, means send block id and the receiver get block data by itself
	//  * Mixed_BroadCast_Mode = 2, means miner use Full_BroadCast_Mode, other nodes use Interactive_BroadCast_Mode
	//  1. 一种是完全块广播模式(Full_BroadCast_Mode)，即直接广播原始块给所有相邻节点;
	//  2. 一种是问询式块广播模式(Interactive_BroadCast_Mode)，即先广播新块的头部给相邻节点，
	//     相邻节点在没有相同块的情况下通过GetBlock主动获取块数据.
	//  3. Mixed_BroadCast_Mode是指出块节点将新块用Full_BroadCast_Mode模式广播，其他节点使用Interactive_BroadCast_Mode
	BlockBroadcaseMode uint8 `yaml:"blockBroadcaseMode,omitempty"`
	// cloud storage config
	CloudStorage CloudStorageConfig `yaml:"cloudStorage,omitempty"`

	Event EventConfig
}

// KernelConfig kernel config
type KernelConfig struct {
	MinNewChainAmount           string          `yaml:"minNewChainAmount,omitempty"`           //创建平行链的最小花费
	NewChainWhiteList           map[string]bool `yaml:"newChainWhiteList,omitempty"`           //能创建链的address白名单
	DisableCreateChainWhiteList bool            `yaml:"disableCreateChainWhiteList,omitempty"` //是否允许任何人创建链
	EnableStopChain             bool            `yaml:"enableStopChain,omitempty"`             //是否开启停用平行链功能
	ModifyBlockAddr             string          `yaml:"modifyBlockAddr,omitempty"`             //是否为可变更区块链
}

// PruneOption ledger prune option
type PruneOption struct {
	Switch        bool   `yaml:"switch,omitempty"`
	Bcname        string `yaml:"bcname,omitempty"`
	TargetBlockid string `yaml:"targetBlockid,omitempty"`
}

// DBCacheConfig db cache config
type DBCacheConfig struct {
	MemCacheSize int `yaml:"memcache,omitempty"`
	FdCacheSize  int `yaml:"fdcache,omitempty"`
}

// XEndorserConfig XEndorser config
type XEndorserConfig struct {
	// Enable use xendorser if true
	Enable bool `yaml:"enable,omitempty"`
	// Module the plugin name for xendorser
	Module   string `yaml:"module,omitempty"`
	ConfPath string `yaml:"confPath,omitempty"`
}

// EventConfig is the config of event service
type EventConfig struct {
	// whether enable event service
	Enable bool
	// max conn count per IP
	AddrMaxConn int
}

func (nc *NodeConfig) defaultNodeConfig() {
	nc.Version = "1.0"
	nc.Log = LogConfig{
		Module:         "xchain",
		Filepath:       "logs",
		Filename:       "xchain",
		Fmt:            "logfmt",
		Console:        true,
		Level:          "debug",
		Async:          false,
		RotateInterval: 60,  // rotate every 60 minutes
		RotateBackups:  168, // keep old log files for 7 days
	}

	nc.TCPServer = TCPServerConfig{
		Port:                  ":37101",
		TLS:                   false,
		TLSPath:               "./data/tls",
		HTTPSPort:             "localhost:37102",
		MetricPort:            "",
		CachePeriod:           2,
		MaxMsgSize:            128 << 20,
		InitialWindowSize:     128 << 10,
		InitialConnWindowSize: 64 << 10,
		ReadBufferSize:        32 << 10,
		WriteBufferSize:       32 << 10,
	}
	nc.P2p = newP2pConfigWithDefault()
	nc.Miner = MinerConfig{
		Keypath: "./data/keys",
	}
	nc.PluginConfPath = "./conf/plugins.conf"
	nc.PluginLoadPath = "./plugins/autoload/"
	nc.Datapath = "./data/blockchain"
	nc.Utxo = UtxoConfig{
		CacheSize:             100000,
		TmpLockSeconds:        60,
		AsyncMode:             false,
		AsyncBlockMode:        false,
		ContractExecutionTime: 500,
		ContractWhiteList:     make(map[string]map[string]bool),
		IsBetaTx:              make(map[string]bool),
		MaxConfirmedDelay:     300,
	}
	nc.DedupCacheSize = 50000
	nc.Kernel = KernelConfig{
		MinNewChainAmount:           "0",
		DisableCreateChainWhiteList: false,
		EnableStopChain:             false,
		ModifyBlockAddr:             "",
	}
	nc.DBCache = DBCacheConfig{
		MemCacheSize: 128,  //MB for each leveldb
		FdCacheSize:  1024, //fd count for each leveldb
	}
	nc.DedupTimeLimit = 15 //seconds
	nc.MemberWhiteList = make(map[string]bool)
	nc.NodeMode = NodeModeNormal
	nc.Wasm = WasmConfig{
		Driver: "xvm",
		XVM: XVMConfig{
			OptLevel: 0,
		},
	}
	nc.EVM = EVMConfig{
		Driver: "evm",
		Enable: false,
	}
	nc.Contract = ContractConfig{
		EnableDebugLog: true,
		DebugLog: LogConfig{
			Module:         "contract",
			Filepath:       "logs",
			Filename:       "contract",
			Fmt:            "logfmt",
			Console:        false,
			Level:          "debug",
			Async:          false,
			RotateInterval: 60 * 24, // rotate every 1 day
			RotateBackups:  14,      // keep old log files for two weeks
		},
		EnableUpgrade: false,
	}
	nc.CoreConnection = false
	nc.FailSkip = false
	nc.XEndorser = XEndorserConfig{
		Enable:   false,
		Module:   "default",
		ConfPath: "",
	}
	nc.BlockBroadcaseMode = 0
	nc.Event = EventConfig{
		Enable:      true,
		AddrMaxConn: 5,
	}
}

// NewNodeConfig returns a config of a node
func NewNodeConfig() *NodeConfig {
	nodeConfig := &NodeConfig{}
	nodeConfig.defaultNodeConfig()
	return nodeConfig
}

// newP2pConfigWithDefault create default p2p configuration
func newP2pConfigWithDefault() P2PConfig {
	return P2PConfig{
		Module:           DefaultP2PModuleName,
		Port:             DefaultNetPort,
		KeyPath:          DefaultNetKeyPath,
		IsNat:            DefaultNetIsNat,
		IsTls:            DefaultNetIsTls,
		IsHidden:         DefaultNetIsHidden,
		IsIpv6:           DefaultNetIsIpv6,
		MaxStreamLimits:  DefaultMaxStreamLimits,
		MaxMessageSize:   DefaultMaxMessageSize,
		Timeout:          DefaultTimeout,
		IsAuthentication: DefaultIsAuthentication,
		// default stream ip limit size
		StreamIPLimitSize:     DefaultStreamIPLimitSize,
		MaxBroadcastPeers:     DefaultMaxBroadcastPeers,
		MaxBroadcastCorePeers: DefaultMaxBroadcastCorePeers,
		IsStorePeers:          DefaultIsStorePeers,
		P2PDataPath:           DefaultP2PDataPath,
		StaticNodes:           make(map[string][]string),
		CertPath:              DefaultCertPath,
		IsUseCert:             true,
		ServiceName:           DefaultServiceName,
		IsBroadCast:           DefaultIsBroadCast,
	}
}

// Validate valid if
func (nc *NodeConfig) Validate() error {
	if nc.NodeMode != NodeModeNormal && nc.NodeMode != NodeModeFastSync {
		return errors.New("Node mode not legal")
	}
	return nil
}

func (nc *NodeConfig) loadConfigFile(configPath string, confName string) error {
	viper.SetConfigName(confName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		//fmt.Println("Read config file error!", "err", err.Error())
		return err
	}
	if err := viper.Unmarshal(nc); err != nil {
		fmt.Println("Unmarshal config from file error! error=", err.Error())
		return err
	}
	return nil
}

// LoadConfig load config from config file
func (nc *NodeConfig) LoadConfig() {

	confPath := "conf"
	confName := "xchain"

	if err := nc.loadConfigFile(confPath, confName); err != nil {
		//fmt.Println("LoadConfigFile error, error = ", err.Error())
		return
	}
}

func (lc *LogConfig) applyFlags(flags *pflag.FlagSet) {

	flags.StringVar(&lc.Module, "module", lc.Module, "used for config overwrite --module <log module>")
	flags.StringVar(&lc.Filename, "filename", lc.Filename, "used for config overwrite --filename <log name>")
	flags.StringVar(&lc.Filepath, "filepath", lc.Filepath, "used for config overwrite --filepath <log name>")
	flags.StringVar(&lc.Fmt, "fmt", lc.Fmt, "used for config overwrite --fmt <log fmt>")
	flags.BoolVar(&lc.Console, "console", lc.Console, "used for config overwrite --console <>")
	flags.StringVar(&lc.Level, "level", lc.Level, "used for config overwrite --level <log level>")
	flags.IntVar(&lc.RotateInterval, "rotateinterval",
		lc.RotateInterval, "used for config overwrite --rotateinterval <log rotate interval>")
	flags.IntVar(&lc.RotateBackups, "rotatebackups",
		lc.RotateBackups, "used for config overwrite --rotatebackups <log rotate backup files>")
}

func (tc *TCPServerConfig) applyFlags(flags *pflag.FlagSet) {

	flags.StringVar(&tc.Port, "port", tc.Port, "used for config overwrite --port <tcp port>, such as: localhost:8888")
	flags.IntVar(&tc.MaxMsgSize, "maxMsgSize", tc.MaxMsgSize,
		"used for config overwrite --maxMsgSize <MAX_MSG_SIZE>, default 4MB")
}

func (mc *MinerConfig) applyFlags(flags *pflag.FlagSet) {
	flags.StringVar(&mc.Keypath, "keypath", mc.Keypath, "used for config overwrite --keypath <node keypath>")
}

func (utxo *UtxoConfig) applyFlags(flags *pflag.FlagSet) {
	flags.IntVar(&utxo.CacheSize, "cachesize", utxo.CacheSize, "used for config overwrite --cachesize <utxo LRU cache size>")
	flags.IntVar(&utxo.TmpLockSeconds, "tmplockSeconds", utxo.TmpLockSeconds, "used for config overwrite --tmplockSeconds <How long to lock utxo referenced by GenerateTx>")
	flags.BoolVar(&utxo.AsyncMode, "asyncMode", utxo.AsyncMode, "used for config overwrite --asyncMode")
	flags.BoolVar(&utxo.AsyncBlockMode, "asyncBlockMode", utxo.AsyncBlockMode, "used for config overwrite --asyncBlockMode")
}

// ApplyFlags install flags and use flags to overwrite config file
func (nc *NodeConfig) ApplyFlags(flags *pflag.FlagSet) {

	nc.Log.applyFlags(flags)
	nc.TCPServer.applyFlags(flags)
	nc.Miner.applyFlags(flags)
	nc.ConsoleConfig.ApplyFlags(flags)
	nc.Utxo.applyFlags(flags)
	nc.Wasm.applyFlags(flags)

	// for backward compatibility
	if nc.Wasm.EnableUpgrade {
		nc.Contract.EnableUpgrade = true
	}

	flags.StringVar(&nc.Datapath, "datapath", nc.Datapath, "used for config overwrite --datapath <data path>")
	flags.StringVar(&nc.CPUProfile, "cpuprofile", nc.CPUProfile, "used to store cpu profile data --cpuprofile <pprof file>")
	flags.StringVar(&nc.MemProfile, "memprofile", nc.MemProfile, "used to store mem profile data --memprofile <pprof file>")

	flags.StringVar(&nc.PluginConfPath, "pluginConfPath", nc.PluginConfPath, "used for config overwrite --pluginConfPath <plugin conf path>")

	flags.BoolVar(&nc.FailSkip, "failSkip", nc.FailSkip, "used for config overwrite --failSkip <>")
	flags.StringVar(&nc.Kernel.ModifyBlockAddr, "modifyBlockAddr", nc.Kernel.ModifyBlockAddr, "used for config overwrite --modifyBlockAddr <>")
}

// VisitAll print all config of node
func (nc *NodeConfig) VisitAll() {
	fmt.Printf("Config of xchain %#v\n", nc)
}
