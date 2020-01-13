package ptree

import (
	"fmt"
	"testing"

	acl_mock "github.com/xuperchain/xuperchain/core/permission/acl/mock"
)

func Test_AccountPTree(t *testing.T) {
	aksuri := make([]string, 7)
	accountName := "Alice"

	aclMgr, err := acl_mock.NewFakeACLManager()
	if err != nil {
		t.Error("NewAclManager failed, err=", err)
		return
	}

	aksuri[0] = "Alice/ak1"
	aksuri[1] = "Alice/ak2"
	aksuri[2] = "Alice/Bob/ak3"
	aksuri[3] = "Alice/ak2"
	aksuri[4] = "Bob/ak5"
	aksuri[5] = "Alice/Kevin/Terry/ak6"
	aksuri[6] = "Alice/Kevin/ak7"

	// build account perm tree
	root, err := BuildAccountPermTree(aclMgr, accountName, aksuri)
	if err != nil {
		t.Error("build account perm tree failed, err=", err)
		return
	}

	plist, err := GetPermTreeList(root)
	if err != nil {
		t.Error("get perm tree list failed, err=", err)
		return
	}

	for idx, node := range plist {
		fmt.Printf("Node index[%d] is %+v\n", idx, node)
	}

	if plist[7].Name != "ak7" || plist[4].Name != "Kevin" || plist[0].Name != "Alice" {
		t.Fatal("Perm tree list error")
	}
}

func Test_ContractMethodPTree(t *testing.T) {
	aksuri := make([]string, 7)
	contractName := "TestContract"
	methodName := "TestMethod"

	aclMgr, err := acl_mock.NewFakeACLManager()
	if err != nil {
		t.Error("NewAclManager failed, err=", err)
		return
	}

	aksuri[0] = "Alice/ak1"
	aksuri[1] = "Alice/ak2"
	aksuri[2] = "Alice/Bob/ak3"
	aksuri[3] = "Alice/ak4"
	aksuri[4] = "Bob/ak5"
	aksuri[5] = "Alice/Kevin/Terry/ak6"
	aksuri[6] = "Alice/Kevin/ak7"

	// build account perm tree
	root, err := BuildMethodPermTree(aclMgr, contractName, methodName, aksuri)
	if err != nil {
		t.Error("build account perm tree failed, err=", err)
		return
	}

	plist, err := GetPermTreeList(root)
	if err != nil {
		t.Error("get perm tree list failed, err=", err)
		return
	}

	for idx, node := range plist {
		fmt.Printf("Node index[%d] is %+v\n", idx, node)
	}

	if plist[11].Name != "ak7" || plist[5].Name != "Bob" || plist[0].Name != "TestMethod" {
		t.Fatal("Perm tree list error")
	}
}
