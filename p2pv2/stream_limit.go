package p2pv2

import (
	"os"
	"strings"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"

	"github.com/xuperchain/log15"
)

// StreamLimit limit the peerID amount of same ip
type StreamLimit struct {
	// Store all streams available
	// key: addr value: peerID
	streams *sync.Map
	// key:ip   value:  amount of same ip
	ip2cnt map[string]int64
	// mutex for ip2cnt
	mutex *sync.Mutex
	// amount limitation of same ip
	// support config file pass
	limit int64
	// log for StreamLimit
	log log.Logger
}

// Init initialize the StreamLimit
func (sl *StreamLimit) Init(limit int64, lg log.Logger) {
	sl.streams = &sync.Map{}
	sl.ip2cnt = make(map[string]int64)
	sl.mutex = &sync.Mutex{}
	if limit <= 0 {
		sl.limit = 1
	} else {
		sl.limit = limit
	}

	if lg == nil {
		lg = log.New("module", "p2pv2")
		lg.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
	}
	sl.log = lg
}

func (sl *StreamLimit) hasStream(addrStr string) bool {
	_, exist := sl.streams.Load(addrStr)
	return exist
}

// AddStream used to add the amount of same ip, plus one per call
func (sl *StreamLimit) AddStream(addrStr string, peerID peer.ID) bool {
	// check if the stream has existed already
	if sl.hasStream(addrStr) {
		sl.log.Trace("StreamLimit AddStream already exists", "addrStr", addrStr, "peerID", peerID)
		return true
	}
	// parse the addr of a stream to be connected
	ip := sl.parseAddrStr(addrStr)
	if ip == "" {
		sl.log.Warn("StreamLimit AddStream failed, stream is invalid ", "addrStr ", addrStr, "peerID ", peerID)
		return false
	}
	// check if the amount of ip of a stream has been full
	if sl.nearFull(ip) {
		sl.log.Warn("StreamLimit AddStream failed, streams has been full ", "ip ", ip)
		return false
	}

	sl.streams.Store(addrStr, peerID)
	sl.inc(ip)

	return true
}

// DelStream used to dec the amount of same ip, dec one per call
func (sl *StreamLimit) DelStream(addrStr string) {
	ip := sl.parseAddrStr(addrStr)
	if ip == "" {
		sl.log.Warn("StreamLimit DelStream failed, stream is invalid ", "addrStr ", addrStr)
		return
	}
	if sl.hasStream(addrStr) {
		sl.dec(ip)
	}
	sl.streams.Delete(addrStr)
}

func (sl *StreamLimit) inc(ip string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	sl.ip2cnt[ip]++
}

func (sl *StreamLimit) dec(ip string) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	sl.ip2cnt[ip]--
}

func (sl *StreamLimit) parseAddrStr(addrStr string) string {
	strSlice := strings.Split(addrStr, "/")
	if len(strSlice) != 5 {
		sl.log.Warn("parseAddrStr failed, invalid addr ", "addr: ", addrStr)
		return ""
	}
	return strSlice[2]
}

func (sl *StreamLimit) nearFull(ip string) bool {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	return sl.ip2cnt[ip] >= sl.limit
}
