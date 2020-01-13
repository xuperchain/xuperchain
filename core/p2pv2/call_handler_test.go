package p2pv2

import (
	"context"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-peer"

	log "github.com/xuperchain/log15"
	xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
	"github.com/xuperchain/xuperunion/pb"
)

func InitMsg(t *testing.T) *xuper_p2p.XuperMessage {
	xchainAddr := &XchainAddrInfo{
		Addr:   Address1,
		Pubkey: []byte(Pubkey1),
		Prikey: []byte(PrivateKey1),
		PeerID: "dKYWwnRHc7Ck",
	}

	auth, err := GetAuthRequest(xchainAddr)
	auths := []*pb.IdentityAuth{}
	auths = append(auths, auth)

	authRes := &pb.IdentityAuths{
		Auth: auths,
	}

	t.Log(authRes)

	msgbuf, err := proto.Marshal(authRes)
	if err != nil {
		t.Log(err.Error())
	}

	msg, err := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion2, "", "",
		xuper_p2p.XuperMessage_GET_AUTHENTICATION, msgbuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		t.Log(err.Error())
	}

	return msg
}

func InitEmptyMsg(t *testing.T) *xuper_p2p.XuperMessage {
	auths := []*pb.IdentityAuth{}
	authRes := &pb.IdentityAuths{
		Auth: auths,
	}

	msgbuf, err := proto.Marshal(authRes)
	if err != nil {
		t.Log(err.Error())
	}

	msg, err := xuper_p2p.NewXuperMessage(xuper_p2p.XuperMsgVersion2, "", "",
		xuper_p2p.XuperMessage_GET_AUTHENTICATION, msgbuf, xuper_p2p.XuperMessage_NONE)
	if err != nil {
		t.Log(err.Error())
	}

	t.Log("InitEmpty success")
	return msg
}

func TestHandleGetAuthentication(t *testing.T) {
	logger := log.New("module", "xchain")
	logger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	p := &P2PServerV2{
		log: logger,
	}

	ctx := context.WithValue(context.Background(), "Stream", MockNewStream())
	_, err := p.handleGetAuthentication(ctx, InitMsg(t))
	if err != nil {
		t.Log(err.Error())
	}

	_, err = p.handleGetAuthentication(ctx, InitEmptyMsg(t))
	if err != nil {
		t.Log(err.Error())
	}

}

// MockNewStream mock new stream
func MockNewStream() *Stream {
	logger := log.New("module", "p2pv2")
	logger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	return &Stream{
		p:        peer.ID("123456789"),
		authAddr: []string{},
		node: &Node{
			log: logger,
		},
	}
}
