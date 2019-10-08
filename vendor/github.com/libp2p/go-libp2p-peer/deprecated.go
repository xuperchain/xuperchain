// Deprecated: use github.com/libp2p/go-libp2p-core/peer instead.
package peer

import (
	core "github.com/libp2p/go-libp2p-core/peer"
	ic "github.com/libp2p/go-libp2p-crypto"
)

var (
	// Deprecated: use github.com/libp2p/go-libp2p-core/peer.ErrEmptyPeerID instead.
	ErrEmptyPeerID = core.ErrEmptyPeerID
	// Deprecated: use github.com/libp2p/go-libp2p-core/peer.ErrNoPublicKey instead.
	ErrNoPublicKey = core.ErrNoPublicKey
)

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.AdvanceEnableInlining instead.
// Warning: this variable's type makes it impossible to alias by reference.
// Reads and writes from/to this variable may be inaccurate or not have the intended effect.
var AdvancedEnableInlining = core.AdvancedEnableInlining

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.ID instead.
type ID = core.ID

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDSlice instead.
type IDSlice = core.IDSlice

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDFromString instead.
func IDFromString(s string) (core.ID, error) {
	return core.IDFromString(s)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDFromBytes instead.
func IDFromBytes(b []byte) (core.ID, error) {
	return core.IDFromBytes(b)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDB58Decode instead.
func IDB58Decode(s string) (core.ID, error) {
	return core.IDB58Decode(s)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDB58Encode instead.
func IDB58Encode(id ID) string {
	return core.IDB58Encode(id)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDHexDecode instead.
func IDHexDecode(s string) (core.ID, error) {
	return core.IDHexDecode(s)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDHexEncode instead.
func IDHexEncode(id ID) string {
	return core.IDHexEncode(id)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDFromPublicKey instead.
func IDFromPublicKey(pk ic.PubKey) (core.ID, error) {
	return core.IDFromPublicKey(pk)
}

// Deprecated: use github.com/libp2p/go-libp2p-core/peer.IDFromPrivateKey instead.
func IDFromPrivateKey(sk ic.PrivKey) (core.ID, error) {
	return core.IDFromPrivateKey(sk)
}
