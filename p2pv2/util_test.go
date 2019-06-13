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
