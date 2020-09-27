package main

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/spf13/viper"
)

type CliConfig struct {
	EndorseServiceHost string                `yaml:"endorseServiceHost,omitempty"`
	ComplianceCheck    ComplianceCheckConfig `yaml:"complianceCheck,omitempty"`
	MinNewChainAmount  string                `yaml:"minNewChainAmount,omitempty"`
	Crypto             string                `yaml:"crypto,omitempty"`
}

// ComplianceCheckConfig: config of xendorser service control
// IsNeedComplianceCheck: is need compliance check
// IsNeedComplianceCheckFee: is need pay for compliance check
// ComplianceCheckEndorseServiceFee: fee for compliance check
// ComplianceCheckEndorseServiceAddr: compliance check addr
type ComplianceCheckConfig struct {
	IsNeedComplianceCheck             bool   `yaml:"isNeedComplianceCheck,omitempty"`
	IsNeedComplianceCheckFee          bool   `yaml:"isNeedComplianceCheckFee,omitempty"`
	ComplianceCheckEndorseServiceFee  int    `yaml:"complianceCheckEndorseServiceFee,omitempty"`
	ComplianceCheckEndorseServiceAddr string `yaml:"complianceCheckEndorseServiceAddr,omitempty"`
}

// NewNodeConfig new a NodeConfig instance
func NewCliConfig() *CliConfig {
	xendorserConfig := &CliConfig{}
	xendorserConfig.defaultCliConfig()
	return xendorserConfig
}

// LoadConfig load Node Config from the specific path/file
func (nc *CliConfig) LoadConfig(fileName string) error {
	confPath, fileNameOnly := filepath.Split(fileName)
	fileSuffix := path.Ext(fileName)
	confName := fileNameOnly[0 : len(fileNameOnly)-len(fileSuffix)]

	if err := nc.loadConfigFile(confPath, confName); err != nil {
		fmt.Printf("loadConfigFile error:%v", err)
		return err
	}
	return nil
}

func (nc *CliConfig) loadConfigFile(configPath string, confName string) error {
	viper.SetConfigName(confName)
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	if err := viper.Unmarshal(nc); err != nil {
		fmt.Printf("Unmarshal config from file error! error:%v", err.Error())
		return err
	}

	return nil
}

func (nc *CliConfig) defaultCliConfig() {
	nc.ComplianceCheck = ComplianceCheckConfig{
		IsNeedComplianceCheck:             false,
		IsNeedComplianceCheckFee:          true,
		ComplianceCheckEndorseServiceFee:  400,
		ComplianceCheckEndorseServiceAddr: "jknGxa6eyum1JrATWvSJKW3thJ9GKHA9n",
	}
	nc.EndorseServiceHost = "localhost:8848"
	nc.MinNewChainAmount = "100"
	nc.Crypto = "xchain"
}
