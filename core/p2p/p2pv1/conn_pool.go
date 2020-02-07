package p2pv1

import (
	log "github.com/xuperchain/log15"
)

// TODO
// ConnPool manage all the connection
type ConnPool struct {
	log log.Logger
	// key: peer id, value: conn
	conns         map[string]*Conn
	maxConnsLimit int32
}

func NewConnPool(lg log.Logger) (*ConnPool, error) {
	return nil, nil
}

// Add add conn to connPool
func Add(*Conn) error {
	return nil
}

// Remove add conn to connpool
func Remove(*Conn) error {
	return nil
}

// Find find conn from connpool
func Find(string) error {
	return nil
}
