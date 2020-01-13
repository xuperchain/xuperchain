package utils

import (
	"testing"

	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/pb"
)

type User struct {
	address    string
	publicKey  string
	privateKey string
}

func TestPhaseMsgSign(t *testing.T) {
	user := &User{
		address:    `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
		publicKey:  `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		privateKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`,
	}
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Error("TestPhaseMsgSign CreateCryptoClient error ", "error", err)
		return
	}

	msg := &pb.ChainedBftPhaseMessage{
		Type:       pb.QCState_PREPARE,
		ViewNumber: 1000,
		ProposalQC: &pb.QuorumCert{},
		JustifyQC:  &pb.QuorumCert{},
		Signature: &pb.SignInfo{
			Address:   user.address,
			PublicKey: user.publicKey,
		},
	}
	priKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(user.privateKey))
	msg, err = MakePhaseMsgSign(cryptoClient, priKey, msg)
	if err != nil {
		t.Error("TestPhaseMsgSign MakePhaseMsgSign error", "error", err)
		return
	}

	ok, err := VerifyPhaseMsgSign(cryptoClient, msg)
	if err != nil || !ok {
		t.Error("TestPhaseMsgSign VerifyPhaseMsgSign error", "error", err, "ok", ok)
		return
	}
}

func TestVoteMsgSign(t *testing.T) {
	user := &User{
		address:    `dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN`,
		publicKey:  `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		privateKey: `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`,
	}
	cryptoClient, err := crypto_client.CreateCryptoClient(crypto_client.CryptoTypeDefault)
	if err != nil {
		t.Error("TestPhaseMsgSign CreateCryptoClient error ", "error", err)
		return
	}
	priKey, _ := cryptoClient.GetEcdsaPrivateKeyFromJSON([]byte(user.privateKey))
	msg := []byte("testmsg")
	sig := &pb.SignInfo{
		Address:   user.address,
		PublicKey: user.publicKey,
	}
	sig, err = MakeVoteMsgSign(cryptoClient, priKey, sig, msg)
	if err != nil {
		t.Error("TestVoteMsgSign MakeVoteMsgSign error", "error", err)
		return
	}
	ok, err := VerifyVoteMsgSign(cryptoClient, sig, msg)
	if err != nil || !ok {
		t.Error("TestVoteMsgSign VerifyVoteMsgSign error", "error", err)
		return
	}
}
