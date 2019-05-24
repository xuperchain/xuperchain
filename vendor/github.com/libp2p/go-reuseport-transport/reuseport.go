package tcpreuse

import (
	"context"
	"net"
	"syscall"

	reuseport "github.com/libp2p/go-reuseport"
)

var fallbackDialer net.Dialer

// reuseErrShouldRetry diagnoses whether to retry after a reuse error.
// if we failed to bind, we should retry. if bind worked and this is a
// real dial error (remote end didnt answer) then we should not retry.
func reuseErrShouldRetry(err error) bool {
	if err == nil {
		return false // hey, it worked! no need to retry.
	}

	// if it's a network timeout error, it's a legitimate failure.
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		return false
	}

	errno, ok := err.(syscall.Errno)
	if !ok { // not an errno? who knows what this is. retry.
		return true
	}

	switch errno {
	case syscall.EADDRINUSE, syscall.EADDRNOTAVAIL:
		return true // failure to bind. retry.
	case syscall.ECONNREFUSED:
		return false // real dial error
	default:
		return true // optimistically default to retry.
	}
}

// Dials using reuseport and then redials normally if that fails.
// TODO(anacrolix): This shouldn't fail anymore: Remove fallback.
func reuseDial(ctx context.Context, laddr *net.TCPAddr, network, raddr string) (con net.Conn, err error) {
	if laddr == nil {
		return fallbackDialer.DialContext(ctx, network, raddr)
	}

	d := net.Dialer{
		LocalAddr: laddr,
		Control:   reuseport.Control,
	}
	defer func() {
		if err != nil {
			return
		}
		// This is transplanted from go-reuseport, which once set no linger on
		// dialing and may be a requirement for desired behaviour in this
		// package.
		con.(*net.TCPConn).SetLinger(0)
	}()

	con, err = d.DialContext(ctx, network, raddr)
	if err == nil {
		return con, nil
	}

	if reuseErrShouldRetry(err) && ctx.Err() == nil {
		log.Debugf("failed to reuse port, dialing with a random port: %s", err)
		con, err = fallbackDialer.DialContext(ctx, network, raddr)
	}
	return con, err
}
