package p2pv1

import (
	"errors"
	"sync"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
)

var (
	ErrAlreadyExist = errors.New("Conn already exist")
	ErrNotExist     = errors.New("Conn not exist")
	ErrRemoveConn   = errors.New("Remove conn error")
)

// ConnPool manage all the connection
type ConnPool struct {
	log    log.Logger
	config config.P2PConfig
	// key: peer id, value: conn
	conns map[string]*Conn
	lock  sync.Mutex
}

// NewConnPool create new connection pool for p2pv1
func NewConnPool(lg log.Logger, cfg config.P2PConfig) (*ConnPool, error) {
	return &ConnPool{
		log:    lg,
		config: cfg,
		conns:  make(map[string]*Conn),
	}, nil
}

// Add add conn to connPool
func (cp *ConnPool) Add(conn *Conn) error {
	cp.log.Info("Add conn", "id", conn.GetConnID())
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if cp.conns[conn.GetConnID()] != nil {
		cp.log.Error("Add conn error", "error", ErrAlreadyExist.Error())
		return ErrAlreadyExist
	}
	cp.conns[conn.GetConnID()] = conn
	return nil
}

// Update add conn to connPool
func (cp *ConnPool) Update(conn *Conn) error {
	cp.log.Info("Update conn", "id", conn.GetConnID())
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if cp.conns[conn.GetConnID()] == nil {
		cp.log.Error("Update conn error", "error", ErrNotExist.Error())
		return ErrNotExist
	}
	cp.conns[conn.GetConnID()].Close()
	delete(cp.conns, conn.GetConnID())
	cp.conns[conn.GetConnID()] = conn
	return nil
}

// Remove add conn to connpool
func (cp *ConnPool) Remove(conn *Conn) error {
	cp.log.Info("Remove conn", "id", conn.GetConnID())
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if cp.conns[conn.GetConnID()] == nil {
		cp.log.Error("Remove conn error, this conn not found")
		return ErrRemoveConn
	}
	cp.conns[conn.GetConnID()].Close()
	delete(cp.conns, conn.GetConnID())
	return nil
}

// Find find conn from connpool, it will establish with the addr if haven't been connected
func (cp *ConnPool) Find(addr string) (*Conn, error) {
	cp.log.Info("Find conn", "id", addr)
	if cp.conns[addr] != nil {
		cp.log.Info("Find conn finded", "id", addr)
		return cp.conns[addr], nil
	}
	conn, err := NewConn(cp.log, addr, cp.config.CertPath, cp.config.ServiceName, cp.config.IsUseCert, int(cp.config.MaxMessageSize)<<20, cp.config.Timeout)
	if err != nil {
		cp.log.Error("Find NewConn error", "error", err.Error(), "id", addr)
		return nil, err
	}
	cp.Add(conn)
	return conn, nil
}

// GetConns return all conn from connpool
func (cp *ConnPool) GetConns() (map[string]*Conn, error) {
	return cp.conns, nil
}
