package client

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/xuperchain/xuperchain/core/common"
	"github.com/xuperchain/xuperchain/core/crypto/account"
	"github.com/xuperchain/xuperchain/core/crypto/client/base"
	"github.com/xuperchain/xuperchain/core/crypto/config"
	"github.com/xuperchain/xuperchain/core/pluginmgr"
)

const (
	// PluginName : plugin name
	PluginName = "crypto"

	// CryptoTypeDefault : default Nist ECC
	CryptoTypeDefault = "default"
	// CryptoTypeGM : support for GM
	CryptoTypeGM = "gm"
	// CryptoTypeSchnorr : support for Nist + Schnorr
	CryptoTypeSchnorr = "schnorr"
)

// cryptoClientFactory is the factory to hold all kinds of crypto clients' instance
type cryptoClientFactory struct {
	mtx     sync.RWMutex
	clients map[string]base.CryptoClient
}

// GetCryptoClient returns a client for the given cryptoType.
func (ccf *cryptoClientFactory) GetCryptoClient(cryptoType string) (base.CryptoClient, error) {
	// case 1: Return the global instance if exists.
	// Since this should be the most frequent case, a read lock is used for better performance.
	ccf.mtx.RLock()
	if v, ok := ccf.clients[cryptoType]; ok {
		ccf.mtx.RUnlock()
		return v, nil
	}
	ccf.mtx.RUnlock()

	// Otherwise, acquire the lock and construct the fresh instance.
	ccf.mtx.Lock()
	defer ccf.mtx.Unlock()

	// Case 2: just return the instance constructed by the 1st locker.
	// This should be the case of the 2nd most frequent.
	if v, ok := ccf.clients[cryptoType]; ok {
		return v, nil
	}

	// Case3: 1st locker made the wanted CryptoClient, which is the least frequent.
	// load crypto plugin
	pluginMgr, err := pluginmgr.GetPluginMgr()
	if err != nil {
		return nil, errors.New("CreateCryptoClient: get plugin mgr failed " + err.Error())
	}

	pluginIns, err := pluginMgr.PluginMgr.CreatePluginInstance(PluginName, cryptoType)
	if err != nil {
		errmsg := "CreateCryptoClient: create plugin failed! name=" + cryptoType
		return nil, errors.New(errmsg)
	}

	cryptoClient := pluginIns.(base.CryptoClient) // missing checking failure of type assertions??
	ccf.clients[cryptoType] = cryptoClient

	return cryptoClient, nil
}

var ccf cryptoClientFactory

// init cryptoClientFactory
func init() {
	ccf.clients = make(map[string]base.CryptoClient)
}

// CreateCryptoClient create CryptoClient of specified cryptoType
func CreateCryptoClient(cryptoType string) (base.CryptoClient, error) {
	return ccf.GetCryptoClient(cryptoType)
}

// CreateCryptoClientFromFilePublicKey create CryptoClient by public key in local file
func CreateCryptoClientFromFilePublicKey(path string) (base.CryptoClient, error) {
	cryptoType, err := getCryptoTypeByFilePublicKey(path)
	if err != nil {
		return nil, err
	}
	// create crypto client
	return CreateCryptoClient(cryptoType)
}

// CreateCryptoClientFromFilePrivateKey create CryptoClient by private key in local file
func CreateCryptoClientFromFilePrivateKey(path string) (base.CryptoClient, error) {
	cryptoType, err := getCryptoTypeByFilePrivateKey(path)
	if err != nil {
		return nil, err
	}
	// create crypto client
	return CreateCryptoClient(cryptoType)
}

// CreateCryptoClientFromJSONPublicKey create CryptoClient by json encoded public key
func CreateCryptoClientFromJSONPublicKey(jsonKey []byte) (base.CryptoClient, error) {
	cryptoType, err := getCryptoTypeByJSONPublicKey(jsonKey)
	if err != nil {
		return nil, err
	}
	// create crypto client
	return CreateCryptoClient(cryptoType)
}

// CreateCryptoClientFromJSONPrivateKey create CryptoClient by json encoded private key
func CreateCryptoClientFromJSONPrivateKey(jsonKey []byte) (base.CryptoClient, error) {
	cryptoType, err := getCryptoTypeByJSONPrivateKey(jsonKey)
	if err != nil {
		return nil, err
	}
	// create crypto client
	return CreateCryptoClient(cryptoType)
}

// CreateCryptoClientFromMnemonic create CryptoClient by mnemonic
func CreateCryptoClientFromMnemonic(mnemonic string, language int) (base.CryptoClient, error) {
	isOld, cryptoByte, err := account.GetCryptoByteFromMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
	}
	// for old mnemonic, only Nist is supported
	if isOld {
		cryptoByte = config.Nist
	}
	cryptoType, err := getTypeByCryptoByte(cryptoByte)
	if err != nil {
		return nil, err
	}
	// create crypto client
	return CreateCryptoClient(cryptoType)
}

func getCryptoTypeByFilePublicKey(path string) (string, error) {
	jsonKey := common.GetFileContent(path)
	return getCryptoTypeByJSONPublicKey([]byte(jsonKey))
}

func getCryptoTypeByFilePrivateKey(path string) (string, error) {
	jsonKey := common.GetFileContent(path)
	return getCryptoTypeByJSONPrivateKey([]byte(jsonKey))
}

func getCryptoTypeByJSONPublicKey(jsonKey []byte) (string, error) {
	publicKey := new(account.ECDSAPublicKey)
	err := json.Unmarshal(jsonKey, publicKey)
	if err != nil {
		return "", err //json有问题
	}
	curveName := publicKey.Curvname
	return getTypeByCurveName(curveName)
}

func getCryptoTypeByJSONPrivateKey(jsonKey []byte) (string, error) {
	privateKey := new(account.ECDSAPrivateKey)
	err := json.Unmarshal(jsonKey, privateKey)
	if err != nil {
		return "", err
	}
	curveName := privateKey.Curvname
	return getTypeByCurveName(curveName)
}

func getTypeByCurveName(name string) (string, error) {
	switch name {
	case "P-256":
		return CryptoTypeDefault, nil
	case "SM2-P-256":
		return CryptoTypeGM, nil
	case "P-256-SN":
		return CryptoTypeSchnorr, nil
	default:
		return "", errors.New("Unknown curve name")
	}
}

func getTypeByCryptoByte(cb uint8) (string, error) {
	switch cb {
	case config.Nist:
		return CryptoTypeDefault, nil
	case config.Gm:
		return CryptoTypeGM, nil
	case config.NistSN:
		return CryptoTypeSchnorr, nil
	default:
		return "", errors.New("Unknown crypto byte")
	}
}
