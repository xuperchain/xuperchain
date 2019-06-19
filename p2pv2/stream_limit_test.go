package p2pv2

import (
	"testing"

	peer "github.com/libp2p/go-libp2p-peer"
)

type TmpStreamLimit struct {
	addr   string
	peerID peer.ID
}

func TestStreamLimitBasic(t *testing.T) {
	sl := &StreamLimit{}
	sl.Init(2, nil)

	testCases := []struct {
		in       *TmpStreamLimit
		expected bool
	}{
		{
			in: &TmpStreamLimit{
				addr:   "/ipv4/127.0.0.1/tcp/47101",
				peerID: peer.ID("1"),
			},
			expected: true,
		},
		{
			in: &TmpStreamLimit{
				addr:   "/ipv4/127.0.0.1/tcp/6718",
				peerID: peer.ID("2"),
			},
			expected: true,
		},
		{
			in: &TmpStreamLimit{
				addr:   "/ipv4/127.0.0.1/tcp/6719",
				peerID: peer.ID("3"),
			},
			expected: false,
		},
		{
			in: &TmpStreamLimit{
				addr:   "/ipv4/127.0.0.2/tcp/6720",
				peerID: peer.ID("4"),
			},
			expected: true,
		},
	}

	for index := range testCases {
		actual := sl.AddStream(testCases[index].in.addr, testCases[index].in.peerID)
		expected := testCases[index].expected
		if actual != expected {
			t.Errorf("expected %v actual %v", expected, actual)
		}
	}
}
