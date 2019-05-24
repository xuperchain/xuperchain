package permission

import (
	crypto_client "github.com/xuperchain/xuperunion/crypto/client"
	"github.com/xuperchain/xuperunion/pb"
	acl_mock "github.com/xuperchain/xuperunion/permission/acl/mock"

	"testing"
)

func Test_IdentifyAK(t *testing.T) {
	ak := "Alice/dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	pubkey := "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571}"
	prikey := "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571,\"D\":29079635126530934056640915735344231956621504557963207107451663058887647996601}"
	msg := "this is the test message from permission"

	xcc, err := crypto_client.CreateCryptoClientFromJSONPublicKey([]byte(pubkey))
	if err != nil {
		t.Error("create crypto client failed, err=", err)
		return
	}

	ecdsaPriKey, err := xcc.GetEcdsaPrivateKeyFromJSON([]byte(prikey))
	if err != nil {
		t.Error("get private key failed, err=", err)
		return
	}

	sign, err := xcc.SignECDSA(ecdsaPriKey, []byte(msg))
	if err != nil {
		t.Error("sign failed, err=", err)
		return
	}

	signInfo := &pb.SignatureInfo{
		Sign:      sign,
		PublicKey: pubkey,
	}

	result, err := IdentifyAK(ak, signInfo, []byte(msg))
	if err != nil || !result {
		t.Error("IdentifyAK failed, result=", result)
	}

	fakesign := "this is a fake signature"
	signInfo2 := &pb.SignatureInfo{
		Sign:      []byte(fakesign),
		PublicKey: pubkey,
	}
	_, err = IdentifyAK(ak, signInfo2, []byte(msg))
	if err == nil {
		t.Error("IdentifyAK fake sign failed, err=", err)
	}
}

func Test_IdentifyAccount(t *testing.T) {
	account := "XC0000000000000001@xuper"
	ak := "XC0000000000000001@xuper/dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	pubkey := "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571}"
	prikey := "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571,\"D\":29079635126530934056640915735344231956621504557963207107451663058887647996601}"
	msg := "this is the test message from permission"

	xcc, err := crypto_client.CreateCryptoClientFromJSONPublicKey([]byte(pubkey))
	if err != nil {
		t.Error("create crypto client failed, err=", err)
		return
	}

	ecdsaPriKey, err := xcc.GetEcdsaPrivateKeyFromJSON([]byte(prikey))
	if err != nil {
		t.Error("get private key failed, err=", err)
		return
	}

	sign, err := xcc.SignECDSA(ecdsaPriKey, []byte(msg))
	if err != nil {
		t.Error("sign failed, err=", err)
		return
	}

	signInfo := &pb.SignatureInfo{
		Sign:      sign,
		PublicKey: pubkey,
	}

	aks := make([]string, 1)
	signs := make([]*pb.SignatureInfo, 1)

	aks[0] = ak
	signs[0] = signInfo
	aclMgr, err := acl_mock.NewFakeACLManager()
	if err != nil {
		t.Error("NewAclManager failed, err=", err)
		return
	}
	pm := &pb.PermissionModel{
		Rule:        pb.PermissionRule_SIGN_THRESHOLD,
		AcceptValue: 1,
	}
	aclObj := &pb.Acl{
		Pm:        pm,
		AksWeight: make(map[string]float64),
	}

	aclObj.AksWeight["dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"] = 1
	aclMgr.SetAccountACL(account, aclObj)
	result, err := IdentifyAccount(account, aks, signs, []byte(msg), aclMgr)
	if err != nil || !result {
		t.Error("IdentifyAccount failed, err=", err)
		return
	}

	fakesign := "this is a fake signature"
	signInfo2 := &pb.SignatureInfo{
		Sign:      []byte(fakesign),
		PublicKey: pubkey,
	}

	signs[0] = signInfo2
	result, err = IdentifyAccount(account, aks, signs, []byte(msg), aclMgr)
	if result == true {
		t.Error("IdentifyAccount fake sign failed, result=", result)
	}
}

func Test_CheckContractMethodPerm(t *testing.T) {
	contractName := "test"
	methodName := "test"
	account := "XC0000000000000001@xuper"
	ak := "XC0000000000000001@xuper/dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
	pubkey := "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571}"
	prikey := "{\"Curvname\":\"P-256\",\"X\":74695617477160058757747208220371236837474210247114418775262229497812962582435,\"Y\":51348715319124770392993866417088542497927816017012182211244120852620959209571,\"D\":29079635126530934056640915735344231956621504557963207107451663058887647996601}"
	msg := "this is the test message from permission"

	xcc, err := crypto_client.CreateCryptoClientFromJSONPublicKey([]byte(pubkey))
	if err != nil {
		t.Error("create crypto client failed, err=", err)
		return
	}

	ecdsaPriKey, err := xcc.GetEcdsaPrivateKeyFromJSON([]byte(prikey))
	if err != nil {
		t.Error("get private key failed, err=", err)
		return
	}

	sign, err := xcc.SignECDSA(ecdsaPriKey, []byte(msg))
	if err != nil {
		t.Error("sign failed, err=", err)
		return
	}

	signInfo := &pb.SignatureInfo{
		Sign:      sign,
		PublicKey: pubkey,
	}

	aks := make([]string, 1)
	signs := make([]*pb.SignatureInfo, 1)

	aks[0] = ak
	signs[0] = signInfo
	aclMgr, err := acl_mock.NewFakeACLManager()
	if err != nil {
		t.Error("NewAclManager failed, err=", err)
		return
	}
	pm := &pb.PermissionModel{
		Rule:        pb.PermissionRule_SIGN_THRESHOLD,
		AcceptValue: 1,
	}
	aclObj := &pb.Acl{
		Pm:        pm,
		AksWeight: make(map[string]float64),
	}

	aclObj.AksWeight[account] = 1
	aclMgr.SetContractMethodACL(contractName, methodName, aclObj)
	result, err := CheckContractMethodPerm(aks, signs, []byte(msg), contractName, methodName, aclMgr)
	if err != nil || !result {
		t.Error("CheckContractMethodPerm failed, err=", err)
		return
	}

	fakesign := "this is a fake signature"
	signInfo2 := &pb.SignatureInfo{
		Sign:      []byte(fakesign),
		PublicKey: pubkey,
	}

	signs[0] = signInfo2
	result, err = CheckContractMethodPerm(aks, signs, []byte(msg), contractName, methodName, aclMgr)
	if result == true {
		t.Error("CheckContractMethodPerm fake sign failed, result=", result)
	}
}
