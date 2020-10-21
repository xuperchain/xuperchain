package validator

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"
	"github.com/tendermint/tendermint/types"
)

// Safety margin determined by Tendermint (see comment on source constant)
var maxTotalPower = big.NewInt(types.MaxTotalVotingPower)
var minTotalPower = big.NewInt(4)

type Bucket struct {
	// Delta tracks the changes to validator power made since the previous rotation
	Delta *Set
	// Previous the value for all validator powers at the point of the last rotation
	// (the sum of all the deltas over all rotations) - these are the history of the complete validator sets at each rotation
	Previous *Set
	// Tracks the current working version of the next set; Previous + Delta
	Next *Set
	// Flow tracks the absolute value of all flows (difference between previous cum bucket and current delta) towards and away from each validator (tracking each validator separately to avoid double counting flows made against the same validator
	Flow *Set
}

func NewBucket(initialSets ...Iterable) *Bucket {
	bucket := &Bucket{
		Previous: NewTrimSet(),
		Next:     NewTrimSet(),
		Delta:    NewSet(),
		Flow:     NewSet(),
	}
	for _, vs := range initialSets {
		vs.IterateValidators(func(id crypto.Addressable, power *big.Int) error {
			bucket.Previous.ChangePower(id.GetPublicKey(), power)
			bucket.Next.ChangePower(id.GetPublicKey(), power)
			return nil
		})
	}
	return bucket
}

// Implement Reader
func (vc *Bucket) Power(id crypto.Address) (*big.Int, error) {
	return vc.Previous.Power(id)
}

// SetPower ensures that validator power would not change too quickly in a single block
func (vc *Bucket) SetPower(id crypto.PublicKey, power *big.Int) (*big.Int, error) {
	const errHeader = "Bucket.SetPower():"
	err := checkPower(power)
	if err != nil {
		return nil, fmt.Errorf("%s %v", errHeader, err)
	}

	nextTotalPower := vc.Next.TotalPower()
	nextTotalPower.Add(nextTotalPower, vc.Next.Flow(id, power))
	// We must not have lower validator power than 4 because this would prevent any flow from occurring
	// min > nextTotalPower
	if minTotalPower.Cmp(nextTotalPower) == 1 {
		return nil, fmt.Errorf("%s cannot change validator power of %v from %v to %v because that would result "+
			"in a total power less than the permitted minimum of 4: would make next total power: %v",
			errHeader, id.GetAddress(), vc.Previous.GetPower(id.GetAddress()), power, nextTotalPower)
	}

	// nextTotalPower > max
	if nextTotalPower.Cmp(maxTotalPower) == 1 {
		return nil, fmt.Errorf("%s cannot change validator power of %v from %v to %v because that would result "+
			"in a total power greater than that allowed by tendermint (%v): would make next total power: %v",
			errHeader, id.GetAddress(), vc.Previous.GetPower(id.GetAddress()), power, maxTotalPower, nextTotalPower)
	}

	// The new absolute flow caused by this AlterPower
	flow := vc.Previous.Flow(id, power)
	absFlow := new(big.Int).Abs(flow)

	// Only check flow if power exists, this allows us to
	// bootstrap the set from an empty state
	if vc.Previous.TotalPower().Sign() > 0 {
		// The max flow we are permitted to allow across all validators
		maxFlow := vc.Previous.MaxFlow()
		// The remaining flow we have to play with
		allowableFlow := new(big.Int).Sub(maxFlow, vc.Flow.totalPower)

		// If we call vc.flow.ChangePower(id, absFlow) (below) will we induce a change in flow greater than the allowable
		// flow we have left to spend?
		if vc.Flow.Flow(id, absFlow).Cmp(allowableFlow) == 1 {
			return nil, fmt.Errorf("%s cannot change validator power of %v from %v to %v because that would result "+
				"in a flow greater than or equal to 1/3 of total power for the next commit: flow induced by change: %v, "+
				"current total flow: %v/%v (cumulative/max), remaining allowable flow: %v",
				errHeader, id.GetAddress(), vc.Previous.GetPower(id.GetAddress()), power, absFlow, vc.Flow.totalPower,
				maxFlow, allowableFlow)
		}
	}
	// Set flow for this id to update flow.totalPower (total flow) for comparison below, keep track of flow for each id
	// so that we only count flow once for each id
	vc.Flow.ChangePower(id, absFlow)
	// Update Delta and Next
	vc.Delta.ChangePower(id, power)
	vc.Next.ChangePower(id, power)
	return absFlow, nil
}

func (vc *Bucket) CurrentSet() *Set {
	return vc.Previous
}

func (vc *Bucket) String() string {
	return fmt.Sprintf("Bucket{Previous: %v; Next: %v; Delta: %v}", vc.Previous, vc.Next, vc.Delta)
}

func (vc *Bucket) Equal(vwOther *Bucket) error {
	err := vc.Delta.Equal(vwOther.Delta)
	if err != nil {
		return fmt.Errorf("bucket delta != other bucket delta: %v", err)
	}
	err = vc.Previous.Equal(vwOther.Previous)
	if err != nil {
		return fmt.Errorf("bucket cum != other bucket cum: %v", err)
	}
	return nil
}

func checkPower(power *big.Int) error {
	if power.Sign() == -1 {
		return fmt.Errorf("cannot set negative validator power: %v", power)
	}
	if !power.IsInt64() {
		return fmt.Errorf("for tendermint compatibility validator power must fit within an int but %v "+
			"does not", power)
	}
	return nil
}
