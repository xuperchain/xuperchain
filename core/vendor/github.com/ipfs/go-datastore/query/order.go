package query

import (
	"bytes"
	"sort"
	"strings"
)

// Order is an object used to order objects
type Order interface {
	Compare(a, b Entry) int
}

// OrderByFunction orders the results based on the result of the given function.
type OrderByFunction func(a, b Entry) int

func (o OrderByFunction) Compare(a, b Entry) int {
	return o(a, b)
}

// OrderByValue is used to signal to datastores they should apply internal
// orderings.
type OrderByValue struct{}

func (o OrderByValue) Compare(a, b Entry) int {
	return bytes.Compare(a.Value, b.Value)
}

// OrderByValueDescending is used to signal to datastores they
// should apply internal orderings.
type OrderByValueDescending struct{}

func (o OrderByValueDescending) Compare(a, b Entry) int {
	return -bytes.Compare(a.Value, b.Value)
}

// OrderByKey
type OrderByKey struct{}

func (o OrderByKey) Compare(a, b Entry) int {
	return strings.Compare(a.Key, b.Key)
}

// OrderByKeyDescending
type OrderByKeyDescending struct{}

func (o OrderByKeyDescending) Compare(a, b Entry) int {
	return -strings.Compare(a.Key, b.Key)
}

// Less returns true if a comes before b with the requested orderings.
func Less(orders []Order, a, b Entry) bool {
	for _, cmp := range orders {
		switch cmp.Compare(a, b) {
		case 0:
		case -1:
			return true
		case 1:
			return false
		}
	}

	// This gives us a *stable* sort for free. We don't care
	// preserving the order from the underlying datastore
	// because it's undefined.
	return a.Key < b.Key
}

// Sort sorts the given entries using the given orders.
func Sort(orders []Order, entries []Entry) {
	sort.Slice(entries, func(i int, j int) bool {
		return Less(orders, entries[i], entries[j])
	})
}
