package errors

import (
	"github.com/hyperledger/burrow/crypto"
)

type Maybe struct {
	// Any error that may have occurred
	error CodedError
}

func (m *Maybe) Error() error {
	if m.error == nil {
		return nil
	}
	return m.error
}

// Errors pushed to state may end up in TxExecutions and therefore the merkle state so it is essential that errors are
// deterministic and independent of the code path taken to execution (e.g. replay takes a different path to that of
// normal consensus reactor so stack traces may differ - as they may across architectures)
func (m *Maybe) PushError(err error) bool {
	if err == nil {
		return false
	}
	if m.error == nil {
		// Make sure we are not wrapping a known nil value
		ex := AsException(err)
		if ex != nil {
			m.error = ex
		}
	}
	return true
}

func (m *Maybe) Uint64(value uint64, err error) uint64 {
	if err != nil {
		m.PushError(err)
	}
	return value
}

func (m *Maybe) Bool(value bool, err error) bool {
	if err != nil {
		m.PushError(err)
	}
	return value
}

func (m *Maybe) Bytes(value []byte, err error) []byte {
	if err != nil {
		m.PushError(err)
	}
	return value
}

func (m *Maybe) Address(value crypto.Address, err error) crypto.Address {
	if err != nil {
		m.PushError(err)
	}
	return value
}

func (m *Maybe) Void(err error) {
	m.PushError(err)
}
