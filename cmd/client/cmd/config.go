package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/viper"
)

// TLSOptions TLS part
type TLSOptions struct {
	Cert   string `yaml:"cert,omitempty"`
	Server string `yaml:"server,omitempty"`
	Enable bool   `yaml:"enable,omitempty"`
}

// ComplianceCheckConfig: config of xendorser service control
// IsNeedComplianceCheck: is need compliance check
// IsNeedComplianceCheckFee: is need pay for compliance check
// ComplianceCheckEndorseServiceFee: fee for compliance check
// ComplianceCheckEndorseServiceAddr: compliance check addr
type ComplianceCheckConfig struct {
	IsNeedComplianceCheck                bool   `yaml:"isNeedComplianceCheck,omitempty"`
	IsNeedComplianceCheckFee             bool   `yaml:"isNeedComplianceCheckFee,omitempty"`
	ComplianceCheckEndorseServiceFee     int    `yaml:"complianceCheckEndorseServiceFee,omitempty"`
	ComplianceCheckEndorseServiceFeeAddr string `yaml:"complianceCheckEndorseServiceFeeAddr,omitempty"`
	ComplianceCheckEndorseServiceAddr    string `yaml:"complianceCheckEndorseServiceAddr,omitempty"`
}

// NewRootOptions new a RootOptions instance
func NewRootOptions() RootOptions {
	rootOptionConfig := RootOptions{}
	rootOptionConfig.setDefaultConf()
	return rootOptionConfig
}

// LoadConfig load Node Config from the specific path/file
func (nc *RootOptions) LoadConfig(fileName string) error {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err

	}
	confPath, fileNameOnly := filepath.Split(fileName)
	fileSuffix := path.Ext(fileName)
	confName := fileNameOnly[0 : len(fileNameOnly)-len(fileSuffix)]

	if err := nc.loadConfigFile(confPath, confName); err != nil {
		fmt.Printf("loadConfigFile error:%v", err)
		return err
	}
	return nil
}

func (nc *RootOptions) loadConfigFile(configPath string, confName string) error {
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

func (nc *RootOptions) setDefaultConf() {
	nc.Host = "127.0.0.1:37101"
	nc.Name = "xuper"
	nc.Keys = "./data/keys"
	nc.Crypto = "default"
	nc.TLS = TLSOptions{
		Cert:   "",
		Server: "",
		Enable: false,
	}
	nc.EndorseServiceHost = "127.0.0.1:8848"
	nc.ComplianceCheck = ComplianceCheckConfig{
		IsNeedComplianceCheck:             false,
		IsNeedComplianceCheckFee:          true,
		ComplianceCheckEndorseServiceFee:  400,
		ComplianceCheckEndorseServiceAddr: "jknGxa6eyum1JrATWvSJKW3thJ9GKHA9n",
	}
	nc.MinNewChainAmount = "100"
}
