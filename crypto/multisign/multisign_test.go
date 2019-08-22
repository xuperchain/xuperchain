package multisign

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/xuperchain/xuperunion/crypto/common"
	"testing"

	"github.com/xuperchain/xuperunion/crypto/account"
)

var (
	addrs   = []string{"dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN", "WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT", "akf7qunmeaqb51Wu418d6TyPKp4jdLdpV"}
	pubkeys = []string{
		`{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`,
		`{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980}`,
		`{"Curvname":"P-256","X":82701086955329320728418181640262300520017105933207363210165513352476444381539,"Y":23833609129887414146586156109953595099225120577035152268521694007099206660741}`,
	}
	privkeys = []string{
		`{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`,
		`{"Curvname":"P-256","X":38583161743450819602965472047899931736724287060636876073116809140664442044200,"Y":73385020193072990307254305974695788922719491565637982722155178511113463088980,"D":98698032903818677365237388430412623738975596999573887926929830968230132692775}`,
		`{"Curvname":"P-256","X":82701086955329320728418181640262300520017105933207363210165513352476444381539,"Y":23833609129887414146586156109953595099225120577035152268521694007099206660741,"D":57537645914107818014162200570451409375770015156750200591470574847931973776404}`,
	}
)

func Test_Multisig(t *testing.T) {
	var privateKeys []*ecdsa.PrivateKey
	var publicKeys []*ecdsa.PublicKey
	var partialSigns [][]byte
	msg := []byte("this is a test message")

	for idx := range addrs {
		priv, _ := account.GetEcdsaPrivateKeyFromJSON([]byte(privkeys[idx]))
		pub, _ := account.GetEcdsaPublicKeyFromJSON([]byte(pubkeys[idx]))
		privateKeys = append(privateKeys, priv)
		publicKeys = append(publicKeys, pub)
	}

	// generate common data
	msc, ks, err := GenCommonPublicKey(publicKeys, msg)
	if err != nil {
		t.Error("GenCommonPublicKey failed with err:", err)
		return
	}

	// generate partial signatures
	for idx, priv := range privateKeys {
		sign, err := GetPartialSign(priv, ks[idx], msc, msg)
		if err != nil {
			t.Error("GetPartialSign failed with err:", err)
			return
		}
		partialSigns = append(partialSigns, sign)
	}

	// merge signatures
	msign, err := MergeMultiSig(partialSigns, msc.R)
	if err != nil {
		t.Error("MergeMultiSig failed with err:", err)
		return
	}

	// verify multisig
	xsign := &common.XuperSignature{}
	err = json.Unmarshal(msign, xsign)
	if err != nil {
		t.Error("unmarshal msign failed with err:", err)
		return
	}

	if xsign.SigType != common.MultiSig {
		t.Error("msign is not Multisig type, SigType:", xsign.SigType)
		return
	}

	ok, err := VerifyMultiSig(publicKeys, xsign.SigContent, msg)
	if err != nil {
		t.Error("VerifyMultiSig failed with err:", err)
		return
	}

	if !ok {
		t.Error("VerifyMultiSig returned false, sign check failed")
		return
	}
}
