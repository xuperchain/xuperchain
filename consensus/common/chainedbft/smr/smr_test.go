package smr

import (
	"testing"
	"time"

	cons_base "github.com/xuperchain/xuperunion/consensus/base"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/config"
	"github.com/xuperchain/xuperunion/consensus/common/chainedbft/external"
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/p2pv2"
	"github.com/xuperchain/xuperunion/pb"
)

type user struct {
	address    string
	publicKey  string
	privateKey string
}

func MakeSmr(t *testing.T) (*Smr, error) {
	user := &user{
		address:    `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
		publicKey:  `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		privateKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`,
	}
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Error("TestPhaseMsgSign CreateCryptoClient error ", "error", err)
		return nil, err
	}
	externalCons := &external.MockExternalConsensus{}
	mockP2p := &p2pv2.MockP2pServer{}
	privateKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(user.privateKey))

	smr, err := NewSmr(
		&config.Config{},
		"xuper",
		user.address,
		user.publicKey,
		privateKey,
		[]*cons_base.CandidateInfo{},
		externalCons,
		cryptoClient,
		mockP2p,
		nil, &pb.QuorumCert{
			ViewNumber: 100,
		}, nil,
	)
	return smr, nil
}

func TestNewSmr(t *testing.T) {
	smr, err := MakeSmr(t)
	if err != nil {
		t.Error("NewSmr error", "error", err)
	}
	go func() {
		time.Sleep(1 * time.Second)
		smr.QuitCh <- true
	}()
	smr.Start()
	return
}
