package relayer

import (
	"fmt"

	"github.com/spf13/viper"
)

// NodeConfig dst and src chain config
// Chains for dst and src chain, and AnchorBlockHeight as initial block header to synchronize
type NodeConfig struct {
	Chains            CrossChainConfig `yaml:"chains,omitempty"`
	AnchorBlockHeight int64            `yaml:"anchorBlockHeight,omitempty"`
}

// ChainConfig config parameters of a chain to be required for relayer
// RPCAddr: chain's rpc infos, such as localhost:37101
// Bcname: chain name to synchronize
// Keys: address to generate tx to synchronize block header for a relayer
// ContractConfig: parameter for synchronization block header contract
type ChainConfig struct {
	RPCAddr        string         `yaml:"rpcAddr,omitempty"`
	Bcname         string         `yaml:"bcname,omitempty"`
	Keys           string         `yaml:"keys,omitempty"`
	ContractConfig ContractConfig `yaml:"contractConfig,omitempty"`
}

// ContractConfig parameters of block header synchronization contract to be required for a relayer
// ModuleName: default as "wasm"
type ContractConfig struct {
	ModuleName   string `yaml:"moduleName,omitempty"`
	ContractName string `yaml:"contractName,omitempty"`
	UpdateMethod string `yaml:"updateMethod,omitempty"`
	AnchorMethod string `yaml:"anchorMethod,omitempty"`
}

// CrossChainConfig parameter about chains including src and dst chain
type CrossChainConfig struct {
	SrcChain ChainConfig `yaml:"srcChain,omitempty"`
	DstChain ChainConfig `yaml:"dstChain,omitempty"`
}

// NewNodeConfig new a NodeConfig instance
func NewNodeConfig() *NodeConfig {
	nodeConfig := &NodeConfig{}
	nodeConfig.defaultNodeConfig()
	return nodeConfig
}

// LoadConfig load Node Config from the specific path/file
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
