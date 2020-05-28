package config

// Config is the config of ChainedBFT, it initialized by Different Consensus
type Config struct {
}

// MakeConfig return config from raw json struct
func MakeConfig(rawConf map[string]interface{}) *Config {
	return &Config{}
}
