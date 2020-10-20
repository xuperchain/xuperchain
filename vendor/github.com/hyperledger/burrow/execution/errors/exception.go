package errors

import "fmt"

func NewException(code *Code, exception string) *Exception {
	if exception == "" {
		return nil
	}
	return &Exception{
		CodeNumber: code.Number,
		Exception:  exception,
	}
}

// Wraps any error as a Exception
func AsException(err error) *Exception {
	if err == nil {
		return nil
	}
	switch e := err.(type) {
	case *Exception:
		return e
	case CodedError:
		return NewException(e.ErrorCode(), e.ErrorMessage())
	default:
		return NewException(Codes.Generic, err.Error())
	}
}

func Wrapf(err error, format string, a ...interface{}) *Exception {
	ex := AsException(err)
	return NewException(Codes.Get(ex.CodeNumber), fmt.Sprintf(format, a...))
}

func Wrap(err error, message string) *Exception {
	ex := AsException(err)
	return NewException(Codes.Get(ex.CodeNumber), message+": "+ex.Exception)
}

func Errorf(code *Code, format string, a ...interface{}) *Exception {
	return NewException(code, fmt.Sprintf(format, a...))
}

func (e *Exception) AsError() error {
	// We need to return a bare untyped error here so that err == nil downstream
	if e == nil {
		return nil
	}
	return e
}

func (e *Exception) ErrorCode() *Code {
	return Codes.Get(e.CodeNumber)
}

func (e *Exception) Error() string {
	return fmt.Sprintf("error %d - %s: %s", e.CodeNumber, Codes.Get(e.CodeNumber), e.Exception)
}

func (e *Exception) String() string {
	return e.Error()
}

func (e *Exception) ErrorMessage() string {
	if e == nil {
		return ""
	}
	return e.Exception
}

func (e *Exception) Equal(ce CodedError) bool {
	ex := AsException(ce)
	if e == nil || ex == nil {
		return e == nil && ex == nil
	}
	return e.CodeNumber == ex.CodeNumber && e.Exception == ex.Exception
}

func (e *Exception) GetCode() *Code {
	return Codes.Get(e.CodeNumber)
}
