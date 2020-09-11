package p2pv2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/sec"
	"io/ioutil"
	"net"
)

// ID is the protocol ID (used when negotiating with multistream)
const ID = "/tls/1.0.0"

// Transport constructs secure communication sessions for a peer.
type Transport struct {
	config *tls.Config

	privKey   crypto.PrivKey
	localPeer peer.ID
}

var _ sec.SecureTransport = &Transport{}

func New(certPath, serviceName string) func(key crypto.PrivKey) (*Transport, error) {
	return func(key crypto.PrivKey) (*Transport, error) {
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

		id, err := peer.IDFromPrivateKey(key)
		if err != nil {
			return nil, err
		}

		return &Transport{
			config: &tls.Config{
				ServerName:   serviceName,
				Certificates: []tls.Certificate{certificate},
				RootCAs:      certPool,
				ClientCAs:    certPool,
				ClientAuth:   tls.RequireAndVerifyClientCert,
			},
			privKey:   key,
			localPeer: id,
		}, nil
	}
}

// SecureInbound runs the TLS handshake as a server.
func (t *Transport) SecureInbound(ctx context.Context, insecure net.Conn) (sec.SecureConn, error) {
	conn := tls.Server(insecure, t.config.Clone())
	if err := conn.Handshake(); err != nil {
		insecure.Close()
		return nil, err
	}

	remotePubKey, err := t.getPeerPubKey(conn)
	if err != nil {
		return nil, err
	}

	return t.setupConn(conn, remotePubKey)
}

// SecureOutbound runs the TLS handshake as a client.
func (t *Transport) SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	conn := tls.Client(insecure, t.config.Clone())
	if err := conn.Handshake(); err != nil {
		insecure.Close()
		return nil, err
	}

	remotePubKey, err := t.getPeerPubKey(conn)
	if err != nil {
		return nil, err
	}

	return t.setupConn(conn, remotePubKey)
}

func (t *Transport) getPeerPubKey(conn *tls.Conn) (crypto.PubKey, error) {
	state := conn.ConnectionState()
	if len(state.PeerCertificates) <= 0 {
		return nil, errors.New("expected one certificates in the chain")
	}

	certKeyPub, err := x509.MarshalPKIXPublicKey(state.PeerCertificates[0].PublicKey)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalRsaPublicKey(certKeyPub)
}

func (t *Transport) setupConn(tlsConn *tls.Conn, remotePubKey crypto.PubKey) (sec.SecureConn, error) {
	remotePeerID, err := peer.IDFromPublicKey(remotePubKey)
	if err != nil {
		return nil, err
	}

	return &conn{
		Conn:         tlsConn,
		localPeer:    t.localPeer,
		privKey:      t.privKey,
		remotePeer:   remotePeerID,
		remotePubKey: remotePubKey,
	}, nil
}

// conn is SecureConn instance
type conn struct {
	*tls.Conn

	localPeer peer.ID
	privKey   crypto.PrivKey

	remotePeer   peer.ID
	remotePubKey crypto.PubKey
}

var _ sec.SecureConn = &conn{}

func (c *conn) LocalPeer() peer.ID {
	return c.localPeer
}

func (c *conn) LocalPrivateKey() crypto.PrivKey {
	return c.privKey
}

func (c *conn) RemotePeer() peer.ID {
	return c.remotePeer
}

func (c *conn) RemotePublicKey() crypto.PubKey {
	return c.remotePubKey
}
