package rule

import (
	"errors"

	"github.com/xuperchain/xuperunion/pb"
	"github.com/xuperchain/xuperunion/permission/ptree"
)

// ACLValidator interface defines common interface for ACL Validator
// Validator only validate account/ak with 1~2 level height
type ACLValidator interface {
	Validate(pnode *ptree.PermNode) (bool, error)
}

// ACLValidatorFactory create ACLValidator for specified permission model
type ACLValidatorFactory struct {
}

// GetACLValidator returns ACLValidator for specified permission model
func (vf *ACLValidatorFactory) GetACLValidator(pr pb.PermissionRule) (ACLValidator, error) {
	switch pr {
	case pb.PermissionRule_NULL:
		return NewNullValidator(), nil
	case pb.PermissionRule_SIGN_THRESHOLD:
		return NewThresholdValidator(), nil
	case pb.PermissionRule_SIGN_AKSET:
		return NewAKSetsValidator(), nil
	case pb.PermissionRule_SIGN_RATE:
		return vf.notImplementedValidator()
	case pb.PermissionRule_SIGN_SUM:
		return vf.notImplementedValidator()
	case pb.PermissionRule_CA_SERVER:
		return vf.notImplementedValidator()
	case pb.PermissionRule_COMMUNITY_VOTE:
		return vf.notImplementedValidator()
	}
	return nil, errors.New("Unknown permission rule")
}

// notImplementedValidator return error for not implemented validator of PermissionRule
func (vf *ACLValidatorFactory) notImplementedValidator() (ACLValidator, error) {
	return nil, errors.New("This permission rule is not implemented")
}
