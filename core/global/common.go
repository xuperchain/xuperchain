package global

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xuperchain/xuperchain/core/pb"

	walletRand "github.com/xuperchain/xuperchain/core/hdwallet/rand"
)

const (
	// VMPrivRing0 ring 0 VMs
	VMPrivRing0 = 0

	// VMPrivRing3 ring 3 VMs
	VMPrivRing3 = 3
)

// UniqMacID return global unique ID
func UniqMacID() string {
	return ""
}

// PathExists check if the specified path exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// MarkPoint define data structure of marked point
type MarkPoint struct {
	tag   string
	delta float64
}

// XTimer define the timestamps and marked points
type XTimer struct {
	bornTime   int64
	latestTime int64
	points     []*MarkPoint
}

// NewXTimer create new XTimer instance
func NewXTimer() *XTimer {
	now := time.Now().UnixNano()
	return &XTimer{
		bornTime:   now,
		latestTime: now,
	}
}

// Mark mark a point and record the tag of the point with time delta
func (timer *XTimer) Mark(tag string) {
	now := time.Now().UnixNano()
	delta := float64(now - timer.latestTime)
	point := &MarkPoint{
		tag:   tag,
		delta: delta,
	}
	timer.latestTime = now
	timer.points = append(timer.points, point)
}

// Print all record points and timestamp information
func (timer *XTimer) Print() string {
	now := time.Now().UnixNano()
	deltaTotal := float64(now - timer.bornTime)
	msg := []string{}
	for _, point := range timer.points {
		msg = append(msg, fmt.Sprintf("%s: %.2f ms", point.tag, point.delta/float64(time.Millisecond)))
	}
	msg = append(msg, fmt.Sprintf("total: %.2fms", deltaTotal/float64(time.Millisecond)))
	return strings.Join(msg, ",")
}

// F print byte slice data as hex string
func F(bytes []byte) string {
	return fmt.Sprintf("%x", bytes)
}

var glogid sync.Mutex
var glogidid int64

// Glogid generate global log id
func Glogid() string {
	glogid.Lock()
	glogidid++
	glogid.Unlock()

	t := time.Now().UnixNano()
	return fmt.Sprintf("%d_%d_%d", t, glogidid, rand.Intn(10000))
}

// GHeader make empty header with logid
func GHeader() *pb.Header {
	h := &pb.Header{}
	h.Logid = Glogid()
	return h
}

// GenNonce generate random nonce
func GenNonce() string {
	return fmt.Sprintf("%d%8d", time.Now().Unix(), rand.Intn(100000000))
}

// SetSeed set random seed
func SetSeed() error {
	// 为math
	seedByte, err := walletRand.GenerateSeedWithStrengthAndKeyLen(walletRand.KeyStrengthHard, walletRand.KeyLengthInt64)
	if err != nil {
		return err
	}
	bytesBuffer := bytes.NewBuffer(seedByte)
	var seed int64
	binary.Read(bytesBuffer, binary.BigEndian, &seed)
	rand.Seed(seed)

	return nil
}
