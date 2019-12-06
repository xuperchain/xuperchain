package main

import "github.com/xuperchain/xuperunion/crypto/client/schnorr"

// GetInstance returns the an instance of SchnorrCryptoClient
func GetInstance() interface{} {
	return &schnorr.SchnorrCryptoClient{}
}
