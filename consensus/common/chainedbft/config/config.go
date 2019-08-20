package config

const (
	// DefaultNetMsgChanSize is the default size of network msg channel
	DefaultNetMsgChanSize = 1000
)

// Config is the config of ChainedBFT, it initialized by Different Consensus
type Config struct {
	// TODO zq Other Configs
	NetMsgChanSize int64
}
