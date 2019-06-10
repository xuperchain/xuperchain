package xchaincore

import (
    "testing"

    "github.com/golang/protobuf/proto"

    "github.com/xuperchain/xuperunion/p2pv2"
    xuper_p2p "github.com/xuperchain/xuperunion/p2pv2/pb"
    "github.com/xuperchain/xuperunion/pb"
)

func InitMsg(t *testing.T) *xuper_p2p.XuperMessage {
    xchainAddr := &p2pv2.XchainAddrInfo{
        Addr:   BobAddress,
		Pubkey: []byte(BobPubkey),
		Prikey: []byte(BobPrivateKey),
		PeerID: "dKYWwnRHc7Ck",
    }

    auth, err := p2pv2.GetAuthRequest(xchainAddr)
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

