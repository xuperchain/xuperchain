package validator

import (
	"math/big"

	"github.com/hyperledger/burrow/crypto"
)

type Writer interface {
	SetPower(id crypto.PublicKey, power *big.Int) (flow *big.Int, err error)
}

type Reader interface {
	Power(id crypto.Address) (*big.Int, error)
}

type Iterable interface {
	IterateValidators(func(id crypto.Addressable, power *big.Int) error) error
}

type IterableReader interface {
	Reader
	Iterable
}

type ReaderWriter interface {
	Reader
	Writer
}

type IterableReaderWriter interface {
	ReaderWriter
	Iterable
}

type History interface {
	ValidatorChanges(blocksAgo int) IterableReader
	Validators(blocksAgo int) IterableReader
}

func AddPower(vs ReaderWriter, id crypto.PublicKey, power *big.Int) error {
	// Current power + power
	currentPower, err := vs.Power(id.GetAddress())
	if err != nil {
		return err
	}
	_, err = vs.SetPower(id, new(big.Int).Add(currentPower, power))
	return err
}

func SubtractPower(vs ReaderWriter, id crypto.PublicKey, power *big.Int) error {
	currentPower, err := vs.Power(id.GetAddress())
	if err != nil {
		return err
	}
	_, err = vs.SetPower(id, new(big.Int).Sub(currentPower, power))
	return err
}

// Returns the asymmetric difference, diff, between two Sets such that applying diff to before results in after
func Diff(before, after IterableReader) (*Set, error) {
	diff := NewSet()
	err := after.IterateValidators(func(id crypto.Addressable, powerAfter *big.Int) error {
		powerBefore, err := before.Power(id.GetAddress())
		if err != nil {
			return err
		}
		// Exclude any powers from before that much after
		if powerBefore.Cmp(powerAfter) != 0 {
			diff.ChangePower(id.GetPublicKey(), powerAfter)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Make sure to zero any validators in before but not in after
	err = before.IterateValidators(func(id crypto.Addressable, powerBefore *big.Int) error {
		powerAfter, err := after.Power(id.GetAddress())
		if err != nil {
			return err
		}
		// If there is a difference value then add to diff
		if powerAfter.Cmp(powerBefore) != 0 {
			diff.ChangePower(id.GetPublicKey(), powerAfter)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return diff, nil
}

func Write(vs Writer, vsOther Iterable) error {
	return vsOther.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		_, err := vs.SetPower(id.GetPublicKey(), power)
		return err
	})
}

// Adds vsOther to vs
func Add(vs ReaderWriter, vsOther Iterable) error {
	return vsOther.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		return AddPower(vs, id.GetPublicKey(), power)
	})
}

// Subtracts vsOther from vs
func Subtract(vs ReaderWriter, vsOther Iterable) error {
	return vsOther.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		return SubtractPower(vs, id.GetPublicKey(), power)
	})
}

func copySet(trim bool, vss []Iterable) *Set {
	vsCopy := newSet()
	vsCopy.trim = trim
	for _, vs := range vss {
		vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
			vsCopy.ChangePower(id.GetPublicKey(), power)
			return nil
		})
	}
	return vsCopy
}

// Copy each of iterable in vss into a new Set - note any iterations errors thrown by the iterable itself will be swallowed
// Use Write instead if source iterables may error
func Copy(vss ...Iterable) *Set {
	return copySet(false, vss)
}

func CopyTrim(vss ...Iterable) *Set {
	return copySet(true, vss)
}
