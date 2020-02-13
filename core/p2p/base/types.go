package base

// CorePeersInfo defines the peers' info for core nodes
// By setting this info, we can keep some core peers always connected directly
// It's useful for keeping DPoS key network security and for some BFT-like consensus
type CorePeersInfo struct {
	Name           string   // distinguished name of the node routing
	CurrentTermNum int64    // the current term number
	CurrentPeerIDs []string // current core peer IDs
	NextPeerIDs    []string // upcoming core peer IDs
}

// XchainAddrInfo xchain addr info
type XchainAddrInfo struct {
	Addr   string
	Pubkey []byte
	Prikey []byte
	PeerID string
}
