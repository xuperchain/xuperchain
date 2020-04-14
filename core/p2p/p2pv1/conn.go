package p2pv1

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
	log "github.com/xuperchain/log15"
	p2pPb "github.com/xuperchain/xuperchain/core/p2p/pb"
	p2p_pb "github.com/xuperchain/xuperchain/core/p2p/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Conn maintain the connection of a node
type Conn struct {
	// addr:"IP:Port"
	id          string
	lg          log.Logger
	conn        *grpc.ClientConn
	maxMsgSize  int
	certPath    string
	serviceName string
	timeOut     int64
	quitCh      chan bool
}

// NewConn create new connection with addr
func NewConn(lg log.Logger, addr string, certPath, serviceName string, maxMsgSize int, timeOut int64) (*Conn, error) {
	conn := &Conn{
		id:          addr,
		lg:          lg,
		maxMsgSize:  maxMsgSize,
		certPath:    certPath,
		serviceName: serviceName,
		timeOut:     timeOut,
		quitCh:      make(chan bool, 1),
	}
	if err := conn.NewGrpcConn(); err != nil {
		lg.Error("NewConn error", "error", err.Error())
		return nil, err
	}
	return conn, nil
}

func genCreds(certPath, serviceName string) (credentials.TransportCredentials, error) {
	bs, err := ioutil.ReadFile(certPath + "/cacert.pem")
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		return nil, err
	}

	certificate, err := tls.LoadX509KeyPair(certPath+"/cert.pem", certPath+"/private.key")
	if err != nil {
		return nil, err
	}
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

func (c *Conn) NewGrpcConn() error {
	conn := &grpc.ClientConn{}
	creds, err := genCreds(c.certPath, c.serviceName)
	if err != nil {
		return err
	}
	conn, err = grpc.Dial(c.id, grpc.WithTransportCredentials(creds), grpc.WithMaxMsgSize(c.maxMsgSize))
	if err != nil {
		c.lg.Error("newGrpcConn error", "error", err, "id", c.id)
		return errors.New("New grpcs conn error")
	}
	c.conn = conn
	return nil
}

func (c *Conn) newClient() (p2p_pb.P2PServiceClient, error) {
	connState := c.conn.GetState().String()
	if connState == "TRANSIENT_FAILURE" || connState == "SHUTDOWN" || connState == "Invalid-State" {
		c.lg.Error("newClient conn state not ready", "state", connState, "id", c.id)
		c.Close()
		err := c.NewGrpcConn()
		if err != nil {
			c.lg.Error("newClient newGrpcConn error", "error", err.Error(), "id", c.id)
			return nil, err
		}
	}
	return p2p_pb.NewP2PServiceClient(c.conn), nil
}

// SendMessage send message to a peer
func (c *Conn) SendMessage(ctx context.Context, msg *p2pPb.XuperMessage) error {
	client, err := c.newClient()
	if err != nil {
		c.lg.Error("SendMessage new client error", "error", err.Error(), "id", c.id)
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := client.SendP2PMessage(ctx)
	if err != nil {
		c.lg.Error("SendMessage new stream error", "error", err.Error(), "id", c.id)
		return err
	}
	waitc := make(chan struct{})
	go func() {
		for {
			_, err = stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				c.lg.Error("SendMessage Recv error", "error", err.Error())
				close(waitc)
				return
			}
		}
	}()
	c.lg.Trace("SendMessage", "logid", msg.GetHeader().GetLogid(), "type", msg.GetHeader().GetType(), "id", c.id)
	err = stream.Send(msg)
	if err != nil {
		c.lg.Error("SendMessage Send error", "error", err.Error(), "id", c.id)
		return err
	}
	stream.CloseSend()
	<-waitc
	if err == io.EOF {
		return nil
	}
	return err
}

// SendMessageWithResponse send message to a peer with responce
func (c *Conn) SendMessageWithResponse(ctx context.Context, msg *p2pPb.XuperMessage) (*p2pPb.XuperMessage, error) {
	client, err := c.newClient()
	if err != nil {
		c.lg.Error("SendMessageWithResponse new client error", "error", err.Error(), "id", c.id)
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := client.SendP2PMessage(ctx)
	if err != nil {
		c.lg.Error("SendMessageWithResponse new stream error", "error", err.Error(), "id", c.id)
		return nil, err
	}
	res := &p2pPb.XuperMessage{}
	waitc := make(chan struct{})
	go func() {
		for {
			res, err = stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				c.lg.Error("SendMessageWithResponse Recv error", "error", err.Error())
				close(waitc)
				return
			}
			if res != nil {
				close(waitc)
				return
			}
		}
	}()
	c.lg.Trace("SendMessageWithResponse", "logid", msg.GetHeader().GetLogid(), "type", msg.GetHeader().GetType(), "id", c.id)
	err = stream.Send(msg)
	if err != nil {
		c.lg.Error("SendMessageWithResponse error", "error", err.Error(), "id", c.id)
		return nil, err
	}
	stream.CloseSend()
	<-waitc
	c.lg.Trace("SendMessageWithResponse return ", "logid", res.GetHeader().GetLogid(), "res", res, "id", c.id)
	return res, err
}

// Close close this conn
func (c *Conn) Close() {
	c.lg.Info("Conn Close", "id", c.id)
	c.conn.Close()
}

// GetConnID return conn id
func (c *Conn) GetConnID() string {
	return c.id
}
