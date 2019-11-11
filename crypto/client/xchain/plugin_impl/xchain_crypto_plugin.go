// Package main is the plugin for xuperchain default crypto client
package main

import (
	"github.com/xuperchain/xuperunion/crypto/client/xchain"
)

// GetInstance returns the an instance of XchainCryptoClient
func GetInstance() interface{} {
	return &eccdefault.XchainCryptoClient{}
}
