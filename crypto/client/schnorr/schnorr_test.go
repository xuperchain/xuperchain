package main

import (
	"bytes"
	"github.com/xuperchain/xuperunion/crypto/account"
	"os"
	"testing"
)

var (
	keypath = "./testkey"
)

func generateKey() error {
	err := os.Mkdir(keypath, os.ModePerm)
	if err != nil {
		return err
	}
	xcc := GetInstance().(*SchnorrCryptoClient)
	err = xcc.ExportNewAccount(keypath)
	return err
}

func readKey() ([]byte, []byte, []byte, error) {
	return account.GetAccInfoFromFile(keypath)
}

func cleanKey() {
	os.Remove(keypath + "/address")
	os.Remove(keypath + "/public.key")
	os.Remove(keypath + "/private.key")
	os.Remove(keypath)
}

func Test_Schnorr(t *testing.T) {
	err := generateKey()
	if err != nil {
		t.Error("generate key failed")
		return
	}
	addr, pub, priv, err := readKey()
	if err != nil {
		t.Error("read key failed")
		return
	}
	t.Logf("created key, address=%s, pub=%s\n", addr, pub)
	defer cleanKey()

	msg := []byte("this is a test msg")

	xcc := GetInstance().(*SchnorrCryptoClient)
	pubkey, err := xcc.GetEcdsaPublicKeyFromJSON(pub)
	if err != nil {
		t.Errorf("GetEcdsaPublicKeyFromJSON failed, err=%v\n", err)
		return
	}
	privkey, err := xcc.GetEcdsaPrivateKeyFromJSON(priv)
	if err != nil {
		t.Errorf("GetEcdsaPrivateKeyFromJSON failed, err=%v\n", err)
		return
	}

	// test encrypt and decrypt
	ciper, err := xcc.Encrypt(pubkey, msg)
	if err != nil {
		t.Errorf("encrypt data failed, err=%v\n", err)
		return
	}

	decode, err := xcc.Decrypt(privkey, ciper)
	if err != nil {
		t.Errorf("Decrypt data failed, err=%v\n", err)
		return
	}

	if bytes.Compare(msg, decode) != 0 {
		t.Errorf("Decrypt data is invalid, decoded=%s\n", string(decode))
		return
	}

	// test sign and verify
	sign, err := xcc.SignECDSA(privkey, msg)
	if err != nil {
		t.Errorf("SignECDSA failed, err=%v\n", err)
		return
	}

	ok, err := xcc.VerifyECDSA(pubkey, sign, msg)
	if err != nil {
		t.Errorf("VerifyECDSA data failed, err=%v\n", err)
		return
	}
	if !ok {
		t.Errorf("VerifyECDSA failed, result is not ok")
		return
	}
}
