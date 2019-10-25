package impl

import (
	"testing"

	"github.com/xuperchain/xuperunion/pb"
)

func Test_GetAccountAddressesWithThreshold(t *testing.T) {
	aclMgr, err := NewACLManager(nil)
	if err != nil {
		t.Error("NewAclManager failed, err=", err)
		return
	}

	// test threshold model
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

	addresses, err := aclMgr.getAddressesByACL(aclObj)
	if err != nil {
		t.Error("get addresses failed, err=", err)
		return
	}
	if len(addresses) != 2 {
		t.Error("get addresses failed, invalid addresses length, len=", len(addresses))
		return
	}
	t.Log("addresses in ACL are:", addresses)
}

func Test_GetAccountAddressesWithAKSet(t *testing.T) {
	aclMgr, err := NewACLManager(nil)
	if err != nil {
		t.Error("NewAclManager failed, err=", err)
		return
	}

	// test ak set
	pm := &pb.PermissionModel{
		Rule:        pb.PermissionRule_SIGN_AKSET,
		AcceptValue: 1,
	}
	aksets := &pb.AkSets{
		Sets:       make(map[string]*pb.AkSet),
		Expression: "",
	}
	aclObj := &pb.Acl{
		Pm:     pm,
		AkSets: aksets,
	}
	set1 := &pb.AkSet{
		Aks: []string{"ak1", "ak2"},
	}

	set2 := &pb.AkSet{
		Aks: []string{"ak3"},
	}

	aclObj.AkSets.Sets["1"] = set1
	aclObj.AkSets.Sets["2"] = set2

	addresses, err := aclMgr.getAddressesByACL(aclObj)
	if err != nil {
		t.Error("get addresses failed, err=", err)
		return
	}
	if len(addresses) != 3 {
		t.Error("get addresses failed, invalid addresses length, len=", len(addresses))
		return
	}
	t.Log("addresses in ACL are:", addresses)

}
