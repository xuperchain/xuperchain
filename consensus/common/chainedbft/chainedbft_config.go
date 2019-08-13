package chainedbft

// Config is the config of ChainedBFT, it initialized by Different Consensus
type Config struct {
	// BroadCastFilter different consensus may have different BroadCastFilter strategy
	BroadCastFilter string
	// TODO zq Other Configs

}
