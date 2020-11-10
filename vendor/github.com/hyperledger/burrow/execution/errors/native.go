package errors

import (
	"fmt"

	"github.com/hyperledger/burrow/crypto"
)

type LacksNativePermission struct {
	Address    crypto.Address
	NativeName string
}

var _ CodedError = &LacksNativePermission{}

func (e *LacksNativePermission) ErrorMessage() string {
	return fmt.Sprintf("account %s does not have native function call permission: %s", e.Address, e.NativeName)
}

func (e *LacksNativePermission) Error() string {
	return e.ErrorMessage()
}

func (e *LacksNativePermission) ErrorCode() *Code {
	return Codes.NativeFunction
}
