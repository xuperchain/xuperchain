package rule

import (
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/permission/ptree"

	"testing"
)

func Test_ValidatorFactory(t *testing.T) {
	vf := ACLValidatorFactory{}
	_, err := vf.GetACLValidator(pb.PermissionRule_SIGN_THRESHOLD)
	if err != nil {
		t.Error("SIGN_THRESHOLD create failed")
		return
	}

	_, err = vf.GetACLValidator(pb.PermissionRule_CA_SERVER)
	if err == nil || err.Error() != "This permission rule is not implemented" {
		t.Error("CA_SERVER error not match")
		return
	}

	_, err = vf.GetACLValidator(pb.PermissionRule(100))
	if err == nil || err.Error() != "Unknown permission rule" {
		t.Error("Unknown permission rule test failed")
		return
	}
}

func Test_NullValidator(t *testing.T) {
	vf := ACLValidatorFactory{}
	tv, err := vf.GetACLValidator(pb.PermissionRule_NULL)
	if err != nil {
		t.Error("NULL create failed")
		return
	}

	// build perm tree
	rootNode := ptree.NewPermNode("Alice", nil)
	ak1Node := ptree.NewPermNode("ak1", nil)
	ak1Node.Status = ptree.Success
	rootNode.Children = append(rootNode.Children, ak1Node)
	result, err := tv.Validate(rootNode)

	// should success
	if err != nil || !result {
		t.Error("validate failed, should have no error and result is true")
		return
	}
}

func Test_ThresholdValidator(t *testing.T) {
	vf := ACLValidatorFactory{}
	tv, err := vf.GetACLValidator(pb.PermissionRule_SIGN_THRESHOLD)
	if err != nil {
		t.Error("SIGN_THRESHOLD create failed")
		return
	}
	pm := &pb.PermissionModel{
		Rule:        pb.PermissionRule_SIGN_THRESHOLD,
		AcceptValue: 2,
	}
	aclObj := &pb.Acl{
		Pm:        pm,
		AksWeight: make(map[string]float64),
	}

	aclObj.AksWeight["ak1"] = 1
	aclObj.AksWeight["ak2"] = 1
	aclObj.AksWeight["ak3"] = 1

	// build perm tree
	rootNode := ptree.NewPermNode("Alice", aclObj)
	ak1Node := ptree.NewPermNode("ak1", nil)
	ak1Node.Status = ptree.Success
	rootNode.Children = append(rootNode.Children, ak1Node)
	result, err := tv.Validate(rootNode)

	// should failed
	if err != nil || result {
		t.Error("validate failed, should have no error and result is false")
		return
	}

	ak3Node := ptree.NewPermNode("ak3", nil)
	ak3Node.Status = ptree.Success
	rootNode.Children = append(rootNode.Children, ak3Node)
	result, err = tv.Validate(rootNode)
	// should success
	if err != nil || !result {
		t.Error("validate failed, should have no error and result is true. result=", result)
		return
	}
}

func Test_AkSetsValidator(t *testing.T) {
	vf := ACLValidatorFactory{}
	akv, err := vf.GetACLValidator(pb.PermissionRule_SIGN_AKSET)
	if err != nil {
		t.Error("SIGN_AKSET create failed")
		return
	}
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

	// build perm tree
	rootNode := ptree.NewPermNode("Alice", aclObj)
	ak1Node := ptree.NewPermNode("ak1", nil)
	ak1Node.Status = ptree.Success
	//ak2Node := ptree.NewPermNode("ak2", nil)
	rootNode.Children = append(rootNode.Children, ak1Node)
	//rootNode.Children = append(rootNode.Children, ak2Node)
	result, err := akv.Validate(rootNode)

	// should failed
	if err != nil || result {
		t.Error("validate failed, should have no error and result is false")
		return
	}

	ak3Node := ptree.NewPermNode("ak3", nil)
	ak3Node.Status = ptree.Success
	rootNode.Children = append(rootNode.Children, ak3Node)
	result, err = akv.Validate(rootNode)
	// should success
	if err != nil || !result {
		t.Error("validate failed, should have no error and result is true")
		return
	}
}
