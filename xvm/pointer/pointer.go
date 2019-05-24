// Package pointer exchange pointer between cgo and go
package pointer

import (
	"sync"
)

var (
	mutex sync.Mutex
	idx   uintptr
	store = make(map[uintptr]interface{})
)

// Save convert a go object to a unique token which can be safely passed to cgo
// The token must be deleted by calling Delete after used
func Save(p interface{}) uintptr {
	mutex.Lock()
	idx++
	store[idx] = p
	mutex.Unlock()
	return idx
}

// Restore restore the token to go object, a invalid token will return nil
func Restore(token uintptr) interface{} {
	var p interface{}
	mutex.Lock()
	p = store[token]
	mutex.Unlock()
	return p
}

// Delete deletes token from internal cache
func Delete(token uintptr) {
	mutex.Lock()
	delete(store, token)
	mutex.Unlock()
}
