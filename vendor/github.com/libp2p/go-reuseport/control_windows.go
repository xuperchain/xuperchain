package reuseport

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func Control(network, address string, c syscall.RawConn) (err error) {
	return c.Control(func(fd uintptr) {
		err = windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
}
