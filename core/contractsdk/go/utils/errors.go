package utils

import "errors"

var (
	ErrMissingCaller    = errors.New("missing caller")
	ErrPermissionDenied = errors.New("you do not have permission to call this method")
	ErrBalanceLow       = errors.New("balance not enough")
)
