package p2pv2

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	//"reflect"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperunion/common/config"
	xuperp2p "github.com/xuperchain/xuperunion/p2pv2/pb"
)

var fullPath string
var tmpStream net.Stream
var globHost1 host.Host
var globHost2 host.Host

func handleStream(s net.Stream) {
	// test for NewStream
	fmt.Println("Got a new stream")
	cfg := config.P2PConfig{
		Port:            20015,
		KeyPath:         "../data/netkeys/",
		IsNat:           true,
		IsSecure:        true,
		IsHidden:        false,
		MaxStreamLimits: 20,
		MaxMessageSize:  config.DefaultMaxMessageSize,
	}
	lg := log.New("module", "p2pv2")
	node, err1 := NewNode(cfg, lg)
	defer func() {
		if node != nil {
			node.Stop()
		}
	}()
	handleMap, _ := NewHandlerMap(lg)
	srv := &P2PServerV2{
		config:     cfg,
		handlerMap: handleMap,
	}
	node.srv = srv
	if err1 != nil {
		fmt.Println("create node error ", err1.Error())
	}
	// test for NewStream
	if node != nil {
		NewStream(s, node)
		// test for stream pool
		streamPool, err := NewStreamPool(int32(20), node, lg)
		defer func() {
			if streamPool != nil {
				streamPool.Stop()
			}
		}()
		if err != nil {
			fmt.Println("new stream pool failed, error ", err.Error())
		} else {
			fmt.Println("new stream pool succeed ", streamPool)
		}
		if streamPool != nil {
			go streamPool.Start()
		}

		// test for Add
		if streamPool != nil {
			tmpStream := streamPool.Add(s)
			if tmpStream != nil {
				fmt.Println("get a good stream")
			} else {
				fmt.Println("get a bad stream")
			}
			fmt.Println("streamLength ", streamPool.streamLength)
			// test for DelStream
			err = streamPool.DelStream(tmpStream)
			if err != nil {
				fmt.Println("DelStream failed, error ", err.Error())
			} else {
				fmt.Println("DelStream succeed ")
			}
			// test for FindStream
			// case1: good case2: bad
			// case 1: good
			tmpStream = streamPool.Add(s)
			_, err = streamPool.FindStream(tmpStream.p)
			if err != nil {
				fmt.Println("when invoke FindStream tmpStream error ", err.Error())
			} else {
				fmt.Println("when invoke FindStream tmpStream succeed")
			}
			fmt.Println("global_host2 ", globHost2.ID())
			// case 2: bad
			_, err = streamPool.FindStream(globHost1.ID())
			fmt.Println("globalHost1 ", globHost1.ID())
			if err != nil {
				fmt.Println("when invoke FindStream globHost1 error ", err.Error())
			}
			// test for sendMessage
			var mg xuperp2p.XuperMessage
			var mgHeader xuperp2p.XuperMessage_MessageHeader
			var mgData xuperp2p.XuperMessage_MessageData
			mgData.MsgInfo = []byte{1}
			mgHeader.Version = "xuperchain1.0"
			mgHeader.Logid = ""
			mgHeader.Bcname = "xuper"
			mgHeader.Type = xuperp2p.XuperMessage_SENDBLOCK
			mg.Header = &mgHeader
			mg.Data = &mgData
			sendErr := streamPool.sendMessage(context.Background(), &mg, tmpStream.p)
			if sendErr != nil {
				fmt.Println("stream pool SendMessage error ", sendErr.Error())
			} else {
				fmt.Println("stream pool SendMessage succeed ")
			}
			err = streamPool.DelStream(tmpStream)
			if err != nil {
				fmt.Println("DelStream tmpStream error ", err.Error())
			}
			peers := []peer.ID{tmpStream.p}
			sendErr = streamPool.SendMessage(context.Background(), nil, peers)
			if sendErr != nil {
				fmt.Println("stream pool SendMessage error ", sendErr.Error())
			}
		}
	}
}

func makeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return basicHost, nil
}

func makeHost(listenF int, target string, secio bool, seed int64) error {
	ha, err := makeBasicHost(listenF, secio, seed)
	if err != nil {
		fmt.Println(err)
		return err
	}
	getFullPath(ha)
	if target == "" {
		fmt.Println("listening for connections")
		ha.SetStreamHandler("/xuper/2.0.0", handleStream)
		globHost1 = ha
		select {}
	} else {
		globHost2 = ha
		ha.SetStreamHandler("/xuper/2.0.0", handleStream)
		ipfsaddr, err := ma.NewMultiaddr(target)
		if err != nil {
			fmt.Println(err)
			return err
		}
		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			return err
		}
		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			return err
		}
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)
		tmpStream, _ = ha.NewStream(context.Background(), peerid, "/xuper/2.0.0")
		select {}
	}
}

func getFullPath(basicHost host.Host) {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))
	addrs := basicHost.Addrs()
	var addr ma.Multiaddr
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	fullAddr := addr.Encapsulate(hostAddr)
	fullPath = fullAddr.String()
}

func TestStreamBasic(t *testing.T) {
	go makeHost(20013, "", false, 0)
	time.Sleep(time.Duration(5) * time.Second)
	go makeHost(20014, fullPath, false, 0)
	time.Sleep(time.Duration(8) * time.Second)
}
