package ptree

import (
	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/acl"
	"github.com/xuperchain/xuperunion/permission/utils"
)

// ValidateStatus define the validation status of a perm node
type ValidateStatus int

const (
	_ ValidateStatus = iota
	// NotVerified : not verified by ACLValidator
	NotVerified
	// Success : ACLValidator verified successful
	Success
	// Failed : ACLValidator verified failed
	Failed
)

// PermNode defines the node of perm tree
type PermNode struct {
	Name     string         // the name(id) of account/ak/method
	ACL      *pb.Acl        // the ACL definition of this account/method
	Status   ValidateStatus // the ACL validation status of this node
	Children []*PermNode    // the children of this node, usually are ACL members of account/method
}

// NewPermNode return a default PermNode
func NewPermNode(akName string, acl *pb.Acl) *PermNode {
	return &PermNode{
		Name:     akName,
		ACL:      acl,
		Status:   NotVerified,
		Children: make([]*PermNode, 0),
	}
}

// FindChild returns the child node with specified name, nil if not found
func (pn *PermNode) FindChild(name string) *PermNode {
	if pn == nil || len(pn.Children) == 0 {
		return nil
	}

	for _, node := range pn.Children {
		if node.Name == name {
			return node
		}
	}

	return nil
}

// BuildAccountPermTree build PermTree for account
func BuildAccountPermTree(aclMgr acl.ManagerInterface, account string, aksuri []string) (*PermNode, error) {
	accountACL, err := utils.GetAccountACL(aclMgr, account)
	if err != nil {
		return nil, err
	}

	root := NewPermNode(account, accountACL)
	root, err = buildPermTree(root, aksuri, aclMgr, true)
	if err != nil {
		return nil, err
	}
	return root, nil
}

// BuildMethodPermTree build PermTree for contract method
func BuildMethodPermTree(aclMgr acl.ManagerInterface, contractName string, methodName string, aksuri []string) (*PermNode, error) {
	methodACL, err := utils.GetContractMethodACL(aclMgr, contractName, methodName)
	if err != nil {
		return nil, err
	}

	root := NewPermNode(methodName, methodACL)
	root, err = buildPermTree(root, aksuri, aclMgr, false)
	if err != nil {
		return nil, err
	}
	return root, nil
}

// build perm tree, not test
func buildPermTree(root *PermNode, aksuri []string, aclMgr acl.ManagerInterface, rootIsAccount bool) (*PermNode, error) {
	akslen := len(aksuri)
	for i := 0; i < akslen; i++ {
		akuri := aksuri[i]

		// split account uri into account/ak list
		aklist := utils.SplitAccountURI(akuri)
		aklen := len(aklist)
		pnode := root
		currentIdx := 0
		// Account PTree has a root node of account, so only accept aksuri start with root.Name
		if rootIsAccount {
			if aklen < 2 || aklist[0] != root.Name {
				continue
			}
			currentIdx = 1
		}
		for ; currentIdx < aklen; currentIdx++ {
			akname := aklist[currentIdx]
			childNode := pnode.FindChild(akname)
			// find current path in perm tree, so go to next level
			if childNode != nil {
				pnode = childNode
				continue
			}
			// not found current path in perm tree, so create new node in tree
			accountACL, err := utils.GetAccountACL(aclMgr, akname)
			if err != nil {
				return nil, err
			}
			newNode := NewPermNode(akname, accountACL)
			pnode.Children = append(pnode.Children, newNode)
			pnode = newNode
		}
	}
	return root, nil
}

// GetPermTreeList return a BFS list of a perm tree
func GetPermTreeList(root *PermNode) ([]*PermNode, error) {
	nlist := make([]*PermNode, 0)
	if root == nil {
		return nlist, nil
	}

	nlist = append(nlist, root)
	pn := 0
	for pn < len(nlist) {
		if nlist[pn].Children != nil {
			for _, node := range nlist[pn].Children {
				nlist = append(nlist, node)
			}
		}
		pn++
	}

	return nlist, nil
}
