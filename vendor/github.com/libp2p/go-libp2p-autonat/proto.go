package autonat

import (
	pb "github.com/libp2p/go-libp2p-autonat/pb"

	"github.com/libp2p/go-libp2p-core/peer"

	logging "github.com/ipfs/go-log"
)

const AutoNATProto = "/libp2p/autonat/1.0.0"

var log = logging.Logger("autonat")

func newDialMessage(pi peer.AddrInfo) *pb.Message {
	msg := new(pb.Message)
	msg.Type = pb.Message_DIAL.Enum()
	msg.Dial = new(pb.Message_Dial)
	msg.Dial.Peer = new(pb.Message_PeerInfo)
	msg.Dial.Peer.Id = []byte(pi.ID)
	msg.Dial.Peer.Addrs = make([][]byte, len(pi.Addrs))
	for i, addr := range pi.Addrs {
		msg.Dial.Peer.Addrs[i] = addr.Bytes()
	}

	return msg
}
