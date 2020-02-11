package p2pv2

import (
	"context"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	peer "github.com/libp2p/go-libp2p-peer"

	log "github.com/xuperchain/log15"
	p2p_base "github.com/xuperchain/xuperchain/core/p2p/base"
	xuper_p2p "github.com/xuperchain/xuperchain/core/p2p/pb"
	"github.com/xuperchain/xuperchain/core/pb"
)

const Address1 = "dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN"
const Pubkey1 = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571}`
const PrivateKey1 = `{"Curvname":"P-256","X":74695617477160058757747208220371236837474210247114418775262229497812962582435,"Y":51348715319124770392993866417088542497927816017012182211244120852620959209571,"D":29079635126530934056640915735344231956621504557963207107451663058887647996601}`
const Peer1 = "QmVxeNubpg1ZQjQT8W5yZC9fD7ZB1ViArwvyGUB53sqf8e"

func InitMsg(t *testing.T) *xuper_p2p.XuperMessage {
	xchainAddr := &p2p_base.XchainAddrInfo{
		Addr:   Address1,
		Pubkey: []byte(Pubkey1),
		Prikey: []byte(PrivateKey1),
		PeerID: "dKYWwnRHc7Ck",
	}

	auth, err := p2p_base.GetAuthRequest(xchainAddr)
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

	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "", "",
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

	msg, err := p2p_base.NewXuperMessage(p2p_base.XuperMsgVersion2, "", "",
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
