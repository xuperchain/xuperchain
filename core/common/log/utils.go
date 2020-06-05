package log

import (
	"fmt"
	"math/rand"
	"path"
	"runtime"
	"strconv"
	"time"
)

// Generate unique id, Not strictly unique
// But the probability of repetition is very low
// Run unit test verification
func GenPseudoUniqId() uint64 {
	nano := time.Now().UnixNano()
	rand.Seed(nano)

	randNum1 := rand.Int63()
	randNum2 := rand.Int63()
	shift1 := rand.Intn(16) + 2
	shift2 := rand.Intn(8) + 1

	uId := ((randNum1 >> uint(shift1)) + (randNum2 >> uint(shift2)) + (nano >> 1)) &
		0x1FFFFFFFFFFFFF
	return uint64(uId)
}

// Generate log id, Not strictly unique
// But the probability of repetition is very low
// Run unit test verification
func GenLogId() string {
	return fmt.Sprintf("%d_%d", time.Now().Unix(), GenPseudoUniqId())
}

// Get call method by runtime.Caller
func GetFuncCall(callDepth int) (string, string) {
	pc, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		return "???:0", "???"
	}

	f := runtime.FuncForPC(pc)
	_, function := path.Split(f.Name())
	_, filename := path.Split(file)

	fline := filename + ":" + strconv.Itoa(line)
	return fline, function
}
