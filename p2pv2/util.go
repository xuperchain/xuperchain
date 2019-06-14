package p2pv2

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/xuperchain/xuperunion/common/config"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/crypto/hash"
	"github.com/xuperchain/xuperunion/pb"
)

// GenerateKeyPairWithPath generate xuper net key pair
func GenerateKeyPairWithPath(path string) error {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return err
	}

	if len(path) == 0 {
		path = config.DefaultNetKeyPath
	}

	privData, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(path, 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(path+"net_private.key", []byte(base64.StdEncoding.EncodeToString(privData)), 0700)
}

// GetKeyPairFromPath get xuper net key from file path
func GetKeyPairFromPath(path string) (crypto.PrivKey, error) {
	if len(path) == 0 {
		path = config.DefaultNetKeyPath
	}

	data, err := ioutil.ReadFile(path + "net_private.key")
	if err != nil {
		return nil, err
	}

	privData, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPrivateKey(privData)
}

// XchainAddrInfo my xchain addr info
type XchainAddrInfo struct {
	Addr   string
	Pubkey []byte
	Prikey []byte
	PeerID string
}

// GetAuthRequest get auth request for authentication
func GetAuthRequest(v *XchainAddrInfo) (*pb.IdentityAuth, error) {
	cryptoClient, err := crypto_client.CreateCryptoClientFromJSONPublicKey(v.Pubkey)
	if err != nil {
		return nil, errors.New("GetAuthRequest: Create crypto client error")
	}

	privateKey, err := cryptoClient.GetEcdsaPrivateKeyFromJSON(v.Prikey)
	if err != nil {
		return nil, err
	}

	timeUnix := time.Now().Unix()
	ts := strconv.FormatInt(timeUnix, 10)

	digestHash := hash.UsingSha256([]byte(v.PeerID + v.Addr + ts))
	sign, err := cryptoClient.SignECDSA(privateKey, digestHash)
	if err != nil {
		return nil, err
	}

	identityAuth := &pb.IdentityAuth{
		Sign:      sign,
		Pubkey:    v.Pubkey,
		Addr:      v.Addr,
		PeerID:    v.PeerID,
		Timestamp: ts,
	}

	return identityAuth, nil
}
