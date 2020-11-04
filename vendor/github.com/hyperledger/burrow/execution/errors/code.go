package errors

import (
	"fmt"
)

// An annotated version of the pure numeric error code
type Code struct {
	Number      uint32
	Name        string
	Description string
}

func code(description string) *Code {
	return &Code{Description: description}
}

func (c *Code) Equal(other *Code) bool {
	if c == nil {
		return false
	}
	return c.Number == other.Number
}

func (c *Code) ErrorCode() *Code {
	return c
}

func (c *Code) Uint32() uint32 {
	if c == nil {
		return 0
	}
	return c.Number
}

func (c *Code) Error() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("Error %d: %s", c.Number, c.Description)
}

func (c *Code) ErrorMessage() string {
	if c == nil {
		return ""
	}
	return c.Description
}

func GetCode(err error) *Code {
	exception := AsException(err)
	if exception == nil {
		return Codes.None
	}
	return exception.GetCode()
}
