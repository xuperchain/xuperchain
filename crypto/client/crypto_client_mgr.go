package client

import (
	"github.com/xuperchain/xuperunion/common"
	"github.com/xuperchain/xuperunion/crypto/account"
	"github.com/xuperchain/xuperunion/crypto/client/base"
	"github.com/xuperchain/xuperunion/crypto/config"
	"github.com/xuperchain/xuperunion/pluginmgr"

	"encoding/json"
	"errors"
	"sync"
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
	mutex   sync.Mutex
	clients map[string]base.CryptoClient
}

func (ccf *cryptoClientFactory) GetCryptoClient(cryptoType string) (base.CryptoClient, error) {
	if _, ok := ccf.clients[cryptoType]; !ok {
		ccf.mutex.Lock()
		defer ccf.mutex.Unlock()
		if _, ok := ccf.clients[cryptoType]; !ok {
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
			cryptoClient := pluginIns.(base.CryptoClient)
			ccf.clients[cryptoType] = cryptoClient
		}
	}
	return ccf.clients[cryptoType], nil
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
	cryptoByte, err := account.GetCryptoByteFromMnemonic(mnemonic, language)
	if err != nil {
		return nil, err
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
