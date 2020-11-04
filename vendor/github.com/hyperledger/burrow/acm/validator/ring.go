package validator

import (
	"fmt"
	"math/big"

	"github.com/hyperledger/burrow/crypto"
)

// Ring stores the validator power history in buckets as a riNng buffer. The primary storage is a the difference between
// each rotation (commit - i.e. block) in 'delta' and the cumulative result of each delta in cum, where the result of
// the delta at i is stored in the cum at i+1. For example suppose we have 4 buckets then graphically:
//
// delta [d1| d2  |   d3    |    d4      ]
// cum   [v0|v0+d1|v0+d1+d2 |v0+d1+d2+d3 ]
//
// After the fourth rotation we loop back to the 0th bucket (with [pwer
//
// delta [d5            | d6| d7   | d8  ]
// cum   [v0+d1+d2+d3+d4|...|      |     ]
type Ring struct {
	buckets []*Bucket
	// Power tracks the sliding sum of all powers for each validator added by each delta bucket - power is added for the newest delta and subtracted from the oldest delta each rotation
	power *Set
	// Index of current head bucket
	head int
	// Number of buckets
	size int
	// Number of buckets that have so far had any increments made to them - equivalently the number of rotations made up to a maximum of the number of buckets available
	populated int
}

var _ History = &Ring{}

// NewRing provides a sliding window over the last size buckets of validator power changes
func NewRing(initialSet Iterable, windowSize int) *Ring {
	if windowSize < 1 {
		windowSize = 1
	}
	vc := &Ring{
		buckets: make([]*Bucket, windowSize),
		power:   NewTrimSet(),
		size:    windowSize,
	}
	for i := 0; i < windowSize; i++ {
		vc.buckets[i] = NewBucket()
	}
	if initialSet != nil {
		vc.populated = 1
		vc.buckets[0] = NewBucket(initialSet)
	}
	return vc
}

// Implement Reader

// Power gets the balance at index from the delta bucket then falling through to the cumulative
func (vc *Ring) Power(id crypto.Address) (*big.Int, error) {
	return vc.GetPower(id), nil
}

func (vc *Ring) GetPower(id crypto.Address) *big.Int {
	return vc.Head().Previous.GetPower(id)
}

func (vc *Ring) SetPower(id crypto.PublicKey, power *big.Int) (*big.Int, error) {
	return vc.Head().SetPower(id, power)
}

// CumulativePower gets the sum of all powers added in any bucket
func (vc *Ring) CumulativePower() *Set {
	return vc.power
}

// Rotate the current head bucket to the next bucket and returns the change in total power between the previous bucket
// and the current head, and the total flow which is the sum of absolute values of all changes each validator's power
// after rotation the next head is a copy of the current head
func (vc *Ring) Rotate() (totalPowerChange *big.Int, totalFlow *big.Int, err error) {
	// Subtract the tail bucket (if any) from the total
	err = Subtract(vc.power, vc.Next().Delta)
	if err != nil {
		return
	}
	// Capture current head as previous before advancing buffer
	prevHead := vc.Head()
	// Add head delta to total power
	err = Add(vc.power, prevHead.Delta)
	if err != nil {
		return
	}
	// Advance the ring buffer
	vc.head = vc.index(1)
	// Overwrite new head bucket (previous tail) with a fresh bucket with Previous_i+1 = Next_i = Previous_i + Delta_i
	vc.buckets[vc.head] = NewBucket(prevHead.Next)
	// Capture flow before we wipe it
	totalFlow = prevHead.Flow.totalPower
	// Subtract the previous bucket total power so we can add on the current buckets power after this
	totalPowerChange = new(big.Int).Sub(vc.Head().Previous.TotalPower(), prevHead.Previous.TotalPower())
	// Record how many of our buckets we have cycled over
	if vc.populated < vc.size {
		vc.populated++
	}
	return
}

func (vc *Ring) ReIndex(newHead int) {
	buckets := make([]*Bucket, len(vc.buckets))
	for i := 0; i < len(buckets); i++ {
		buckets[(i+newHead)%len(buckets)] = vc.buckets[vc.index(i)]
	}
	vc.head = newHead
	vc.buckets = buckets
}

func (vc *Ring) CurrentSet() *Set {
	return vc.Head().Previous
}

// Get the current accumulator bucket
func (vc *Ring) Head() *Bucket {
	return vc.buckets[vc.head]
}

func (vc *Ring) ValidatorChanges(blocksAgo int) IterableReader {
	return vc.PreviousDelta(blocksAgo)
}

func (vc *Ring) Validators(blocksAgo int) IterableReader {
	return vc.PreviousSet(blocksAgo)
}

func (vc *Ring) PreviousSet(delay int) *Set {
	// report the oldest cumulative set (i.e. genesis) if given a longer delay than populated
	if delay >= vc.populated {
		delay = vc.populated - 1
	}
	return vc.buckets[vc.index(-delay)].Previous
}

func (vc *Ring) PreviousDelta(delay int) *Set {
	if delay >= vc.populated {
		return NewSet()
	}
	return vc.buckets[vc.index(-delay)].Delta
}

func (vc *Ring) Next() *Bucket {
	return vc.buckets[vc.index(1)]
}

func (vc *Ring) index(i int) int {
	return (vc.size + vc.head + i) % vc.size
}

// Get the number of buckets in the ring (use Current().Count() to get the current number of validators)
func (vc *Ring) Size() int {
	return vc.size
}

// Returns buckets in order head, previous, ...
func (vc *Ring) OrderedBuckets() []*Bucket {
	buckets := make([]*Bucket, len(vc.buckets))
	for i := int(0); i < vc.size; i++ {
		index := vc.index(-i)
		buckets[i] = vc.buckets[index]
	}
	return buckets
}

func (vc *Ring) String() string {
	buckets := vc.OrderedBuckets()
	return fmt.Sprintf("ValidatorsRing{Total: %v; Buckets: %v}", vc.power, buckets)
}

func (vc *Ring) Equal(vcOther *Ring) error {
	if vc.size != vcOther.size {
		return fmt.Errorf("ring size %d != other ring size %d", vc.size, vcOther.size)
	}
	if vc.head != vcOther.head {
		return fmt.Errorf("ring head index %d != other head index %d", vc.head, vcOther.head)
	}
	err := vc.power.Equal(vcOther.power)
	if err != nil {
		return fmt.Errorf("ring power != other ring power: %v", err)
	}
	for i := 0; i < len(vc.buckets); i++ {
		err = vc.buckets[i].Equal(vcOther.buckets[i])
		if err != nil {
			return fmt.Errorf("ring buckets do not match at index %d: %v", i, err)
		}
	}
	return nil
}
