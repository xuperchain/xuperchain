package js

import "fmt"

var (
	// ExceptionNotfound wraps the EEXIS errno
	ExceptionNotfound = NewException("EEXIS", "not found")
	// ExceptionNoSys wraps the ENOSYS errno
	ExceptionNoSys = NewException("ENOSYS", "not implemention")
	// ExceptionInvalidArgument wraps the EINVAL errno
	ExceptionInvalidArgument = NewException("EINVAL", "invalid argument")
	// ExceptionUndefined wraps the EINVAL errno
	ExceptionUndefined = NewException("EINVAL", "undefined")
)

// Exception simulates js Exception
type Exception struct {
	Code    string
	Message string
}

// NewException instances a Exception
func NewException(code, msg string) *Exception {
	return &Exception{
		Code:    code,
		Message: msg,
	}
}

// Error returns the error message of Exception
func (e *Exception) Error() string {
	return e.Message
}

// ExceptionRefNotFound is the Exception throwed when Ref is not found by VM
func ExceptionRefNotFound(ref Ref) *Exception {
	return NewException("EEXIS", fmt.Sprintf("ref %x not found", ref))
}

// ThrowException throw an exception
func ThrowException(e *Exception) {
	panic(e)
}

// Throw uses a fmt like style to throw an exception
func Throw(fmtstr string, args ...interface{}) {
	ThrowException(&Exception{
		Message: fmt.Sprintf(fmtstr, args...),
	})
}
