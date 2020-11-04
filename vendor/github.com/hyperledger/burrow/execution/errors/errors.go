package errors

type CodedError interface {
	error
	ErrorCode() *Code
	// The error message excluding the code
	ErrorMessage() string
}

// Error sinks are useful for recording errors but continuing with computation. Implementations may choose to just store
// the first error pushed and ignore subsequent ones or they may record an error trace. Pushing a nil error should have
// no effects.
type Sink interface {
	// Push an error to the error. If a nil error is passed then that value should not be pushed. Returns true iff error
	// is non nil.
	PushError(error) bool
}

type Source interface {
	// Returns the an error if errors occurred some execution or nil if none occurred
	Error() error
}

var Codes = codes{
	None:                   code("none"),
	UnknownAddress:         code("unknown address"),
	InsufficientBalance:    code("insufficient balance"),
	InvalidJumpDest:        code("invalid jump destination"),
	InsufficientGas:        code("insufficient gas"),
	MemoryOutOfBounds:      code("memory out of bounds"),
	CodeOutOfBounds:        code("code out of bounds"),
	InputOutOfBounds:       code("input out of bounds"),
	ReturnDataOutOfBounds:  code("data out of bounds"),
	CallStackOverflow:      code("call stack overflow"),
	CallStackUnderflow:     code("call stack underflow"),
	DataStackOverflow:      code("data stack overflow"),
	DataStackUnderflow:     code("data stack underflow"),
	InvalidContract:        code("invalid contract"),
	PermissionDenied:       code("permission denied"),
	NativeContractCodeCopy: code("tried to copy native contract code"),
	ExecutionAborted:       code("execution aborted"),
	ExecutionReverted:      code("execution reverted"),
	NativeFunction:         code("native function error"),
	EventPublish:           code("event publish error"),
	InvalidString:          code("invalid string"),
	EventMapping:           code("event mapping error"),
	Generic:                code("generic error"),
	InvalidAddress:         code("invalid address"),
	DuplicateAddress:       code("duplicate address"),
	InsufficientFunds:      code("insufficient funds"),
	Overpayment:            code("overpayment"),
	ZeroPayment:            code("zero payment error"),
	InvalidSequence:        code("invalid sequence number"),
	ReservedAddress:        code("address is reserved for SNative or internal use"),
	IllegalWrite:           code("callee attempted to illegally modify state"),
	IntegerOverflow:        code("integer overflow"),
	InvalidProposal:        code("proposal is invalid"),
	ExpiredProposal:        code("proposal is expired since sequence number does not match"),
	ProposalExecuted:       code("proposal has already been executed"),
	NoInputPermission:      code("account has no input permission"),
	InvalidBlockNumber:     code("invalid block number"),
	BlockNumberOutOfRange:  code("block number out of range"),
	AlreadyVoted:           code("vote already registered for this address"),
	UnresolvedSymbols:      code("code has unresolved symbols"),
	InvalidContractCode:    code("contract being created with unexpected code"),
	NonExistentAccount:     code("account does not exist"),
}

func init() {
	err := Codes.init()
	if err != nil {
		panic(err)
	}
}
