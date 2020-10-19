// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package names

import (
	"fmt"
	"reflect"

	"github.com/hyperledger/burrow/event/query"
)

var MinNameRegistrationPeriod uint64 = 5

const (

	// NOTE: base costs and validity checks are here so clients
	// can use them without importing state

	// cost for storing a name for a block is
	// CostPerBlock*CostPerByte*(len(data) + 32)
	NameByteCostMultiplier  uint64 = 1
	NameBlockCostMultiplier uint64 = 1

	MaxNameLength = 64
	MaxDataLength = 1 << 16
)

func (e *Entry) String() string {
	return fmt.Sprintf("NameEntry{%v -> %v; Expires: %v, Owner: %v}", e.Name, e.Data, e.Expires, e.Owner)
}

func (e *Entry) Get(key string) (value interface{}, ok bool) {
	return query.GetReflect(reflect.ValueOf(e), key)
}

type Reader interface {
	GetName(name string) (*Entry, error)
}

type Writer interface {
	// Updates the name entry creating it if it does not exist
	UpdateName(entry *Entry) error
	// Remove the name entry
	RemoveName(name string) error
}

type ReaderWriter interface {
	Reader
	Writer
}

type Iterable interface {
	IterateNames(consumer func(*Entry) error) (err error)
}

type IterableReader interface {
	Iterable
	Reader
}

type IterableReaderWriter interface {
	Iterable
	ReaderWriter
}

// base cost is "effective" number of bytes
func NameBaseCost(name, data string) uint64 {
	return uint64(len(data) + 32)
}

func NameCostPerBlock(baseCost uint64) uint64 {
	return NameBlockCostMultiplier * NameByteCostMultiplier * baseCost
}

func NameCostForExpiryIn(name, data string, expiresIn uint64) uint64 {
	return NameCostPerBlock(NameBaseCost(name, data)) * expiresIn
}
