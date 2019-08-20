package sign

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
)

func Test_Signature(t *testing.T) {
	curve := elliptic.P256()
	curve.Params().Name = "P-256-SN"
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Errorf("generate key failed, err=%v\n", err)
		return
	}
	// sign
	msg := []byte("this is a test message")
	sign, err := Sign(privateKey, msg)
	if err != nil {
		t.Errorf("sign message failed, err=%v\n", err)
		return
	}
	t.Log("sign message success")

	// verify
	ok, err := Verify(&privateKey.PublicKey, sign, msg)
	if err != nil {
		t.Errorf("verify message failed, err=%v\n", err)
		return
	}

	if !ok {
		t.Error("verify message failed, ok is false")
		return
	}
}
