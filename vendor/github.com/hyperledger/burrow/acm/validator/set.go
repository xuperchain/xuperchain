package validator

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/hyperledger/burrow/crypto"
)

var big0 = big.NewInt(0)
var big1 = big.NewInt(1)
var big2 = big.NewInt(2)
var big3 = big.NewInt(3)

// A Validator multiset - can be used to capture the global state of validators or as an accumulator each block
type Set struct {
	powers     map[crypto.Address]*big.Int
	publicKeys map[crypto.Address]crypto.Addressable
	totalPower *big.Int
	trim       bool
}

func newSet() *Set {
	return &Set{
		totalPower: new(big.Int),
		powers:     make(map[crypto.Address]*big.Int),
		publicKeys: make(map[crypto.Address]crypto.Addressable),
	}
}

// Create a new Validators which can act as an accumulator for validator power changes
func NewSet() *Set {
	return newSet()
}

// Like Set but removes entries when power is set to 0 this make Count() == CountNonZero() and prevents a set from leaking
// but does mean that a zero will not be iterated over when performing an update which is necessary in Ring
func NewTrimSet() *Set {
	s := newSet()
	s.trim = true
	return s
}

// Implements Writer, but will never error
func (vs *Set) SetPower(id crypto.PublicKey, power *big.Int) (*big.Int, error) {
	return vs.ChangePower(id, power), nil
}

// Add the power of a validator and returns the flow into that validator
func (vs *Set) ChangePower(id crypto.PublicKey, power *big.Int) *big.Int {
	address := id.GetAddress()
	// Calculate flow into this validator (positive means in, negative means out)
	flow := vs.Flow(id, power)
	vs.totalPower.Add(vs.totalPower, flow)

	if vs.trim && power.Sign() == 0 {
		delete(vs.publicKeys, address)
		delete(vs.powers, address)
	} else {
		vs.publicKeys[address] = crypto.NewAddressable(id)
		vs.powers[address] = new(big.Int).Set(power)
	}
	return flow
}

func (vs *Set) TotalPower() *big.Int {
	return new(big.Int).Set(vs.totalPower)
}

// Returns the maximum allowable flow whilst ensuring the majority of validators are non-byzantine after the transition
// So need at most ceiling((Total Power)/3) - 1, in integer division we have ceiling(X*p/q) = (p(X+1)-1)/q
// For p = 1 just X/q so we want (Total Power)/3 - 1
func (vs *Set) MaxFlow() *big.Int {
	max := vs.TotalPower()
	return max.Sub(max.Div(max, big3), big1)
}

// Returns the flow that would be induced by a validator power change
func (vs *Set) Flow(id crypto.PublicKey, power *big.Int) *big.Int {
	return new(big.Int).Sub(power, vs.GetPower(id.GetAddress()))
}

// Returns the power of id but only if it is set
func (vs *Set) MaybePower(id crypto.Address) *big.Int {
	if vs.powers[id] == nil {
		return nil
	}
	return new(big.Int).Set(vs.powers[id])
}

// Version of Power to match interface
func (vs *Set) Power(id crypto.Address) (*big.Int, error) {
	return vs.GetPower(id), nil
}

// Error free version of Power
func (vs *Set) GetPower(id crypto.Address) *big.Int {
	if vs.powers[id] == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(vs.powers[id])
}

// Returns an error if the Sets are not equal describing which part of their structures differ
func (vs *Set) Equal(vsOther *Set) error {
	if vs.Size() != vsOther.Size() {
		return fmt.Errorf("set size %d != other set size %d", vs.Size(), vsOther.Size())
	}
	// Stop iteration IFF we find a non-matching validator
	return vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		otherPower := vsOther.GetPower(id.GetAddress())
		if otherPower.Cmp(power) != 0 {
			return fmt.Errorf("set power %d != other set power %d", power, otherPower)
		}
		return nil
	})
}

// Iterates over validators sorted by address
func (vs *Set) IterateValidators(iter func(id crypto.Addressable, power *big.Int) error) error {
	if vs == nil {
		return nil
	}
	addresses := make(crypto.Addresses, 0, len(vs.powers))
	for address := range vs.powers {
		addresses = append(addresses, address)
	}
	sort.Sort(addresses)
	for _, address := range addresses {
		err := iter(vs.publicKeys[address], new(big.Int).Set(vs.powers[address]))
		if err != nil {
			return err
		}
	}
	return nil
}

func (vs *Set) Flush(output Writer, backend Reader) error {
	return vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		_, err := output.SetPower(id.GetPublicKey(), power)
		return err
	})
}

func (vs *Set) CountNonZero() int {
	var count int
	vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		if power.Sign() != 0 {
			count++
		}
		return nil
	})
	return count
}

func (vs *Set) Size() int {
	return len(vs.publicKeys)
}

func (vs *Set) Validators() []*Validator {
	if vs == nil {
		return nil
	}
	pvs := make([]*Validator, 0, vs.Size())
	vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		pvs = append(pvs, &Validator{PublicKey: id.GetPublicKey(), Power: power.Uint64()})
		return nil
	})
	return pvs
}

func UnpersistSet(pvs []*Validator) *Set {
	vs := NewSet()
	for _, pv := range pvs {
		vs.ChangePower(pv.PublicKey, new(big.Int).SetUint64(pv.Power))
	}
	return vs
}

func (vs *Set) String() string {
	return fmt.Sprintf("Validators{TotalPower: %v; Count: %v; %v}", vs.TotalPower(), vs.Size(),
		vs.Strings())
}

func (vs *Set) Strings() string {
	strs := make([]string, 0, vs.Size())
	vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
		strs = append(strs, fmt.Sprintf("%v->%v", id.GetAddress(), power))
		return nil
	})
	return strings.Join(strs, ", ")
}
