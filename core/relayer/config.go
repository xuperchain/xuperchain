package relayer

import (
	"fmt"

	"github.com/spf13/viper"
)

// global config such as
/*
type GlobalConfig struct {
}*/

type NodeConfig struct {
	Chains            CrossChainConfig `yaml:"chains,omitempty"`
	AnchorBlockHeight int64            `yaml:"anchorBlockHeight,omitempty"`
}

type ChainConfig struct {
	RPCAddr        string         `yaml:"rpcAddr,omitempty"`
	Bcname         string         `yaml:"bcname,omitempty"`
	Keys           string         `yaml:"keys,omitempty"`
	ContractConfig ContractConfig `yaml:"contractConfig,omitempty"`
}

type ContractConfig struct {
	ModuleName   string `yaml:"moduleName,omitempty"`
	ContractName string `yaml:"contractName,omitempty"`
	UpdateMethod string `yaml:"updateMethod,omitempty"`
	AnchorMethod string `yaml:"anchorMethod,omitempty"`
}

// parameter about chains including src and dst chain
//
type CrossChainConfig struct {
	SrcChain ChainConfig `yaml:"srcChain,omitempty"`
	DstChain ChainConfig `yaml:"dstChain,omitempty"`
}

func NewNodeConfig() *NodeConfig {
	nodeConfig := &NodeConfig{}
	nodeConfig.defaultNodeConfig()
	return nodeConfig
}

func (nc *NodeConfig) LoadConfig() {
	confPath := "conf"
	confName := "relayer"

	if err := nc.loadConfigFile(confPath, confName); err != nil {
		fmt.Println("loadConfigFile error:", err)
		return
	}
}

func (nc *NodeConfig) loadConfigFile(configPath string, confName string) error {
	viper.SetConfigName(confName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	if err := viper.Unmarshal(nc); err != nil {
		fmt.Println("Unmarshal config from file error! error=", err.Error())
		return err
	}

	return nil
}

func (nc *NodeConfig) defaultNodeConfig() {
	nc.Chains = CrossChainConfig{
		SrcChain: ChainConfig{
			RPCAddr: "localhost:6720",
			Bcname:  "xuper",
		},
		DstChain: ChainConfig{
			RPCAddr: "localhost:6720",
			Bcname:  "xuper",
			Keys:    "./data/keys",
			ContractConfig: ContractConfig{
				ModuleName:   "wasm",
				ContractName: "xuper_relayer",
				UpdateMethod: "putBlockHeader",
				AnchorMethod: "initAnchorBlockHeader",
			},
		},
	}
	nc.AnchorBlockHeight = int64(0)
}
