// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package slice

// Convenience function
func Slice(elements ...interface{}) []interface{} {
	return elements
}

func EmptySlice() []interface{} {
	return []interface{}{}
}

// Like append but on the interface{} type and always to a fresh backing array
// so can be used safely with slices over arrays you did not create.
func CopyAppend(slice []interface{}, elements ...interface{}) []interface{} {
	sliceLength := len(slice)
	newSlice := make([]interface{}, sliceLength+len(elements))
	copy(newSlice, slice)
	copy(newSlice[sliceLength:], elements)
	return newSlice
}

// Prepend elements to slice in the order they appear
func CopyPrepend(slice []interface{}, elements ...interface{}) []interface{} {
	elementsLength := len(elements)
	newSlice := make([]interface{}, len(slice)+elementsLength)
	copy(newSlice, elements)
	copy(newSlice[elementsLength:], slice)
	return newSlice
}

// Concatenate slices into a single slice
func Concat(slices ...[]interface{}) []interface{} {
	offset := 0
	for _, slice := range slices {
		offset += len(slice)
	}
	concat := make([]interface{}, offset)
	offset = 0
	for _, slice := range slices {
		for i, e := range slice {
			concat[offset+i] = e
		}
		offset += len(slice)
	}
	return concat
}

// Deletes n elements starting with the ith from a slice by splicing.
// Beware uses append so the underlying backing array will be modified!
func Delete(slice []interface{}, i int, n int) []interface{} {
	return append(slice[:i], slice[i+n:]...)
}

// Flatten a slice by a list by splicing any elements of the list that are
// themselves lists into the slice elements to the list in place of slice itself
func Flatten(slice []interface{}) []interface{} {
	return DeepFlatten(slice, 1)
}

// Recursively flattens a list by splicing any sub-lists into their parent until
// depth is reached. If a negative number is passed for depth then it continues
// until no elements of the returned list are lists
func DeepFlatten(slice []interface{}, depth int) []interface{} {
	if depth == 0 {
		return slice
	}
	returnSlice := []interface{}{}

	for _, element := range slice {
		if s, ok := element.([]interface{}); ok {
			returnSlice = append(returnSlice, DeepFlatten(s, depth-1)...)
		} else {
			returnSlice = append(returnSlice, element)
		}

	}

	return returnSlice
}
