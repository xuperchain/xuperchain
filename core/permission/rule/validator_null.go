package rule

import (
	"github.com/xuperchain/xuperunion/permission/ptree"
)

// NullValidator is Valiator for Null permission model
type NullValidator struct{}

// NewNullValidator return instance of NullValidator
func NewNullValidator() *NullValidator {
	return &NullValidator{}
}

// Validate always return true for NullValidator
func (nv *NullValidator) Validate(pnode *ptree.PermNode) (bool, error) {
	return true, nil
}
