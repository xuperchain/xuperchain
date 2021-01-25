package code

import "errors"

var (
	ErrMissingInitiator = errors.New("missing initiator")
	ErrPermissionDenied = errors.New("you do not have permission to call this method")
	ErrBalanceLow       = errors.New("balance not enough")
)
