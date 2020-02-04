package main

import (
	"github.com/xuperchain/xuperchain/core/p2p/p2pv2"
)

// GetInstance returns the an instance of SchnorrCryptoClient
func GetInstance() interface{} {
	return &p2pv2.P2PServerV2{}
}
