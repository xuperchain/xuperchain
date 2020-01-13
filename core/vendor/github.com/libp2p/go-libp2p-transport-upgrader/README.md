# go-libp2p-transport-upgrader

[![](https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square)](https://protocol.ai)
[![](https://img.shields.io/badge/project-libp2p-yellow.svg?style=flat-square)](https://libp2p.io/)
[![](https://img.shields.io/badge/freenode-%23libp2p-yellow.svg?style=flat-square)](http://webchat.freenode.net/?channels=%23libp2p)
[![GoDoc](https://godoc.org/github.com/libp2p/go-libp2p-transport-upgrader?status.svg)](https://godoc.org/github.com/libp2p/go-libp2p-transport-upgrader)
[![Build Status](https://travis-ci.org/libp2p/go-libp2p-transport-upgrader.svg?branch=master)](https://travis-ci.org/libp2p/go-libp2p-transport-upgrader)
[![Discourse posts](https://img.shields.io/discourse/https/discuss.libp2p.io/posts.svg)](https://discuss.libp2p.io)

> Stream connection to libp2p connection upgrader

This package provides the necessary logic to upgrade [multiaddr-net][manet] connections listeners into full [libp2p-transport][tpt] connections and listeners.

To use, construct a new `Upgrader` with:

* An optional [pnet][pnet] `Protector`.
* An optional [multiaddr-net][manet] address `Filter`.
* A mandatory [stream security transport][ss].
* A mandatory [stream multiplexer transport][smux].

[tpt]: https://github.com/libp2p/go-libp2p-transport
[manet]: https://github.com/multiformats/go-multiaddr-net
[ss]: https://github.com/libp2p/go-conn-security
[smux]: https://github.com/libp2p/go-stream-muxer
[pnet]: https://github.com/libp2p/go-libp2p-interface-pnet

Note: This package largely replaces the functionality of [go-libp2p-conn](https://github.com/libp2p/go-libp2p-conn) but with half the code.

## Install

`go-libp2p-transport-upgrader` is a standard Go module which can be installed with:

```sh
go get github.com/libp2p/go-libp2p-transport-upgrader
```

This repo is [gomod](https://github.com/golang/go/wiki/Modules)-compatible, and users of
go 1.11 and later with modules enabled will automatically pull the latest tagged release
by referencing this package. Upgrades to future releases can be managed using `go get`,
or by editing your `go.mod` file as [described by the gomod documentation](https://github.com/golang/go/wiki/Modules#how-to-upgrade-and-downgrade-dependencies).

## Usage

## Example

Below is a simplified TCP transport implementation using the transport upgrader. In practice, you'll want to use [go-tcp-transport](https://github.com/libp2p/go-tcp-transport) (which has reuseport support).

```go
package tcptransport

import (
	"context"

	tptu "github.com/libp2p/go-libp2p-transport-upgrader"

	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/whyrusleeping/mafmt"
	manet "github.com/multiformats/go-multiaddr-net"
	tpt "github.com/libp2p/go-libp2p-transport"
)

// TcpTransport is a simple TCP transport.
type TcpTransport struct {
	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	Upgrader *tptu.Upgrader
}

var _ tpt.Transport = &TcpTransport{}

// NewTCPTransport creates a new TCP transport instance.
func NewTCPTransport(upgrader *tptu.Upgrader) *TcpTransport {
	return &TcpTransport{Upgrader: upgrader}
}

// CanDial returns true if this transport believes it can dial the given
// multiaddr.
func (t *TcpTransport) CanDial(addr ma.Multiaddr) bool {
	return mafmt.TCP.Matches(addr)
}

// Dial dials the peer at the remote address.
func (t *TcpTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (tpt.Conn, error) {
    var dialer manet.Dialer
    conn, err := dialer.DialContext(ctx, raddr)
	if err != nil {
		return nil, err
	}
	return t.Upgrader.UpgradeOutbound(ctx, t, conn, p)
}

// Listen listens on the given multiaddr.
func (t *TcpTransport) Listen(laddr ma.Multiaddr) (tpt.Listener, error) {
	list, err := manet.Listen(laddr)
	if err != nil {
		return nil, err
	}
	return t.Upgrader.UpgradeListener(t, list), nil
}

// Protocols returns the list of terminal protocols this transport can dial.
func (t *TcpTransport) Protocols() []int {
	return []int{ma.P_TCP}
}

// Proxy always returns false for the TCP transport.
func (t *TcpTransport) Proxy() bool {
	return false
}
```

## Contribute

Feel free to join in. All welcome. Open an [issue](https://github.com/libp2p/go-libp2p-transport-upgrader/issues)!

This repository falls under the IPFS [Code of Conduct](https://github.com/libp2p/community/blob/master/code-of-conduct.md).

### Want to hack on IPFS?

[![](https://cdn.rawgit.com/jbenet/contribute-ipfs-gif/master/img/contribute.gif)](https://github.com/ipfs/community/blob/master/contributing.md)

## License

MIT

---

The last gx published version of this module was: 0.1.28: QmeqC5shQjEBRG9B8roZqQCJ9xb7Pq6AbWxJFMyLgqBBWh
