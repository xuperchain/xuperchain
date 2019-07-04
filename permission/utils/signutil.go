package utils

import (
	"crypto/ecdsa"
	"errors"

	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/pb"
)

// VerifySign verify if the signature of data and the public key are match
// Return true without error means passed the check
func VerifySign(ak string, si *pb.SignatureInfo, data []byte) (bool, error) {
	bytespk := []byte(si.PublicKey)
	xcc, err := crypto_client.CreateCryptoClientFromJSONPublicKey(bytespk)
	if err != nil {
		return false, err
	}

	ecdsaKey, err := xcc.GetEcdsaPublicKeyFromJSON(bytespk)
	if err != nil {
		return false, err
	}

	isMatch, _ := xcc.VerifyAddressUsingPublicKey(ak, ecdsaKey)
	if !isMatch {
		return false, errors.New("address and public key not match")
	}

	ks := []*ecdsa.PublicKey{}
	ks = append(ks, ecdsaKey)
	return xcc.VerifyXuperSignature(ks, si.Sign, data)
}
