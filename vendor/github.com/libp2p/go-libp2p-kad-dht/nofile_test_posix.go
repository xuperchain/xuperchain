// +build !windows,!wasm

package dht

import "syscall"

func curFileLimit() uint64 {
	var n syscall.Rlimit
	syscall.Getrlimit(syscall.RLIMIT_NOFILE, &n)
	// cast because some platforms use int64 (e.g., freebsd)
	return uint64(n.Cur)
}
