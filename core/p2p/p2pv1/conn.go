package p2pv1

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"sync"

	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Conn maintain the connection of a node
type Conn struct {
	// addr:"IP:Port"
	id     string
	lg     log.Logger
	conn   *grpc.ClientConn
	quitCh chan bool
	lock   *sync.RWMutex
}

func NewConn(lg log.Logger, addr string, certPath, serviceName string, maxMsgSize int) *Conn {
	conn := &Conn{
		id:     addr,
		lg:     lg,
		quitCh: make(chan bool, 1),
		lock:   &sync.RWMutex{},
	}
	conn.NewGrpcConn(maxMsgSize, certPath, serviceName)
	return conn
}

func genCreds(certPath, serviceName string) (credentials.TransportCredentials, error) {
	bs, err := ioutil.ReadFile(certPath + "/cert.crt")
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		return nil, err
	}

	certificate, err := tls.LoadX509KeyPair(certPath+"/key.pem", certPath+"/private.key")
	creds := credentials.NewTLS(
		&tls.Config{
			ServerName:   serviceName,
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
			ClientCAs:    certPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})
	return creds, nil
}

func (c *Conn) NewGrpcConn(maxMsgSize int, certPath string, serviceName string) error {
	if c.conn == nil {
		creds, err := genCreds(certPath, serviceName)
		if err != nil {
			return err
		}
		conn, err := grpc.Dial(c.id, grpc.WithTransportCredentials(creds), grpc.WithMaxMsgSize(maxMsgSize))
		if err != nil {
			return errors.New("New grpcs conn error!")
		}
		c.conn = conn
		return nil
	}
	connState := c.conn.GetState().String()
	if connState == "TRANSIENT_FAILURE" || connState == "SHUTDOWN" || connState == "Invalid-State" {
		c.conn.Close()
		creds, err := genCreds(certPath, serviceName)
		if err != nil {
			return err
		}
		conn, err := grpc.Dial(c.id, grpc.WithTransportCredentials(creds), grpc.WithMaxMsgSize(maxMsgSize))
		if err != nil {
			return errors.New("New grpcs conn error!")
		}
		c.conn = conn
		return nil
	}
	return nil
}

func (c *Conn) Close() {
	c.conn.Close()
}
