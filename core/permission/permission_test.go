package permission

import (
	crypto_client "github.com/xuperchain/xuperchain/core/crypto/client"
	"github.com/xuperchain/xuperchain/core/pb"
	acl_mock "github.com/xuperchain/xuperchain/core/permission/acl/mock"

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

	aks := make([]string, 1)

	aks[0] = ak
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

	aclObj.AksWeight["dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"] = 0.5
	aclObj.AksWeight["WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT"] = 0.5
	aclMgr.SetAccountACL(account, aclObj)

	// should not pass
	result, err := IdentifyAccount(account, aks, aclMgr)
	if result {
		t.Error("IdentifyAccount test failed, acl should not pass")
		return
	}

	// should pass
	aks = append(aks, "XC0000000000000001@xuper/WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT")
	result, err = IdentifyAccount(account, aks, aclMgr)
	if !result {
		t.Error("IdentifyAccount test failed , should pass, err", err)
	}
}

func Test_CheckContractMethodPerm(t *testing.T) {
	contractName := "test"
	methodName := "test"
	account := "XC0000000000000001@xuper"
	ak := "XC0000000000000001@xuper/dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"

	aks := make([]string, 1)

	aks[0] = ak
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

	aclObj.AksWeight["dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"] = 0.5
	aclObj.AksWeight["WNWk3ekXeM5M2232dY2uCJmEqWhfQiDYT"] = 0.5
	aclMgr.SetAccountACL(account, aclObj)

	pm2 := &pb.PermissionModel{
		Rule:        pb.PermissionRule_SIGN_THRESHOLD,
		AcceptValue: 1,
	}
	aclObj2 := &pb.Acl{
		Pm:        pm2,
		AksWeight: make(map[string]float64),
	}

	aclObj2.AksWeight[account] = 1
	aclMgr.SetContractMethodACL(contractName, methodName, aclObj2)
	result, err := CheckContractMethodPerm(aks, contractName, methodName, aclMgr)
	if err != nil || !result {
		t.Error("CheckContractMethodPerm failed, err=", err)
		return
	}
}
