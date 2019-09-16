package config

const (
	// DefaultNetMsgChanSize is the default size of network msg channel
	DefaultNetMsgChanSize = 1000
)

// Config is the config of ChainedBFT, it initialized by Different Consensus
type Config struct {
	// TODO zq Other Configs
	NetMsgChanSize int64 `json:"netMsgChanSize"`
}

// MakeConfig return config from raw json struct
func MakeConfig(rawConf map[string]interface{}) *Config {
	conf := &Config{
		NetMsgChanSize: DefaultNetMsgChanSize,
	}
	return conf
}
