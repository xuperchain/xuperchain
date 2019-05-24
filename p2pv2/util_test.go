package p2pv2

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/libp2p/go-libp2p-peer"
)

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
