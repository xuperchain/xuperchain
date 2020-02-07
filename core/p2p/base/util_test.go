package base

import (
	"io/ioutil"
	"os"
	"testing"

	peer "github.com/libp2p/go-libp2p-peer"
)

const Address1 = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
const Pubkey1 = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
const PrivateKey1 = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
const Peer1 = "QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e"

func TestGenerateKeyPairWithPath(t *testing.T) {
	workSpace, _ := ioutil.TempDir("", "tmp")
	defer os.RemoveAll(workSpace)
	err := GenerateKeyPairWithPath(workSpace)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetKeyPairFromPath(t *testing.T) {
	priv, err := GetKeyPairFromPath("")
	if err != nil {
		t.Error(err.Error())
	}
	id, err := peer.IDFromPublicKey(priv.GetPublic())
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(id.Pretty())
	}
}

func TestGenerateUniqueRandList(t *testing.T) {
	max := 10
	size := 12
	resList := GenerateUniqueRandList(size, max)
	if len(resList) != max {
		t.Errorf("generate rand list failed, list len should be equal to max\n")
	}
	size = 0
	resList = GenerateUniqueRandList(size, max)
	if len(resList) != size {
		t.Errorf("generate rand list failed, list len should be equal to size\n")
	}

	size = 10
	resList = GenerateUniqueRandList(size, max)
	dupCheck := make(map[int]bool)
	for i := 0; i < len(resList); i++ {
		if dupCheck[resList[i]] {
			t.Errorf("duplicate value found, list=%v\n", resList)
			break
		}
		dupCheck[resList[i]] = true
	}
}

func TestGetAuthRequest(t *testing.T) {
	xchainAddr := &XchainAddrInfo{
		Addr:   Address1,
		Pubkey: []byte(Pubkey1),
		Prikey: []byte(PrivateKey1),
		PeerID: Peer1,
	}

	auth, err := GetAuthRequest(xchainAddr)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(auth)
	}
}
