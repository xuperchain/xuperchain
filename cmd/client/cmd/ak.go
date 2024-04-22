/*
 * Copyright (c) 2022. Baidu Inc. All Rights Reserved.
 */

package cmd

import (
	"math/big"
	"os"
	"path/filepath"

	"github.com/xuperchain/xupercore/lib/crypto/client/base"

	"github.com/xuperchain/xuperchain/models"
	"github.com/xuperchain/xuperchain/service/common"
	"github.com/xuperchain/xuperchain/service/pb"
)

// file names in AK directory
const (
	FileAddress    = "address"
	FilePublicKey  = "public.key"
	FilePrivateKey = "private.key"
)

/*
	AK with its directory like:

└── <AK name>

	├── address
	├── private.key
	└── public.key
*/
type AK struct {
	path string
}

// create AK with path
// Params:
//
//	path: path refers to <AK name>
func newAK(path string) AK {
	return AK{
		path: path,
	}
}

/*
	 listAKs lists all AK by root, which could be like:

	 └── <root>	// normally as `data`
		├── keys	// default AK name
		│	├── address
		│	├── private.key
		│	└── public.key
		├── <AK name>
		│	├── address
		│	├── private.key
		│	└── public.key
		└── <other dir>

Params:

	root: root path refers to <root>
*/
func listAKs(root string) ([]AK, error) {
	dirs, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	akList := make([]AK, 0)
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		akPath := filepath.Join(root, dir.Name())
		ak := newAK(akPath)
		if ak.IsAKDir() {
			akList = append(akList, ak)
		}
	}
	return akList, nil
}

// IsAKDir return true for AK directory
func (ak AK) IsAKDir() bool {
	// if `address` file not exists, it isn't an AK directory
	address := filepath.Join(ak.path, FileAddress)
	_, err := os.Stat(address)
	return err == nil
}

// AuthRequirementFrom return an auth require text for the source account
func (ak *AK) AuthRequirementFrom(account string) (string, error) {
	address, err := readAddress(ak.path)
	if err != nil {
		return "", err
	}

	authRequirement := account + "/" + address
	return authRequirement, nil
}

// Info gets AK info from its directory,
// includes: address, key pair
func (ak *AK) Info() (info AKInfo, err error) {
	info.address, err = readAddress(ak.path)
	if err != nil {
		return AKInfo{}, err
	}

	info.KeyPair, err = ak.keyPair()
	if err != nil {
		return AKInfo{}, err
	}

	return info, nil
}

// keyPair gets key pair for AK
func (ak *AK) keyPair() (KeyPair, error) {

	pk, err := readPublicKey(ak.path)
	if err != nil {
		return KeyPair{}, err
	}
	sk, err := readPrivateKey(ak.path)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		publicKey: pk,
		secretKey: sk,
	}, nil
}

// SignTx signs for a transaction with given crypto client
func (ak *AK) SignTx(tx *pb.Transaction, crypto base.CryptoClient) (*pb.SignatureInfo, error) {
	pk, err := ak.keyPair()
	if err != nil {
		return nil, err
	}
	return pk.SignTx(tx, crypto)
}

// AK information which is store in files
type AKInfo struct {
	address string
	KeyPair // public & private key pair
}

// SignUtxo self signs for an UTXO with given crypto client
// Params:
//
//	bcName: UTXO blockchain name
//	amount: UTXO amount
//	crypto: crypto client
func (i *AKInfo) SignUtxo(bcName string, amount *big.Int, crypto base.CryptoClient) (pb.SignatureInfo, error) {
	lockedUtxo := models.NewLockedUtxo(bcName, i.address, amount)
	return i.KeyPair.SignUtxo(lockedUtxo, crypto)
}

// KeyPair is key-pair for crypto
type KeyPair struct {
	publicKey, secretKey string
}

// SignTx signs for a transaction with given crypto client by AK key pair
func (p *KeyPair) SignTx(tx *pb.Transaction, crypto base.CryptoClient) (*pb.SignatureInfo, error) {
	sign, err := common.ComputeTxSign(crypto, tx, []byte(p.secretKey))
	if err != nil {
		return nil, err
	}
	return &pb.SignatureInfo{
		PublicKey: p.publicKey,
		Sign:      sign,
	}, nil
}

// SignUtxo signs for an UTXO with given crypto client by AK key pair
// Params:
//
//	utxo: locked UTXO
//	crypto: crypto client
func (p *KeyPair) SignUtxo(utxo *models.LockedUtxo, crypto base.CryptoClient) (pb.SignatureInfo, error) {

	// prepare private key
	ecdsaPrivateKey, err := crypto.GetEcdsaPrivateKeyFromJsonStr(p.secretKey)
	if err != nil {
		return pb.SignatureInfo{}, err
	}

	// sign
	sign, err := crypto.SignECDSA(ecdsaPrivateKey, utxo.Hash())
	if err != nil {
		return pb.SignatureInfo{}, err
	}
	return pb.SignatureInfo{
		PublicKey: p.publicKey,
		Sign:      sign,
	}, nil
}
