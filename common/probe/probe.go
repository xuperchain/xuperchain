package probe

import (
	"github.com/xuperchain/log15"
	"runtime"
	"sync"
	"time"
)

// SpeedCalc used to calculate the speed
type SpeedCalc struct {
	start    int64
	mutex    sync.Mutex
	dict     map[string]int64
	maxSpeed map[string]float64
	prefix   string
}

// NewSpeedCalc init a speed calculator
func NewSpeedCalc(prefix string) *SpeedCalc {
	return &SpeedCalc{
		start:    0,
		maxSpeed: make(map[string]float64),
		dict:     make(map[string]int64),
		prefix:   prefix,
	}
}

// Clear clear speed calculator
func (sc *SpeedCalc) Clear() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.start = 0
	sc.dict = make(map[string]int64)
}

// Add add speed calculator time
func (sc *SpeedCalc) Add(flag string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	if sc.start <= 0 {
		sc.start = time.Now().UnixNano()
	}
	c, ok := sc.dict[flag]
	if ok {
		sc.dict[flag] = c + 1
	} else {
		sc.dict[flag] = 1
	}
}

// GetMaxSpeed return the max speed
func (sc *SpeedCalc) GetMaxSpeed() map[string]float64 {
	return sc.maxSpeed
}

// ShowInfo show the speed
func (sc *SpeedCalc) ShowInfo(log log.Logger) {
	end := time.Now().UnixNano()
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	dis := float64(end-sc.start) / 1e9
	for k, v := range sc.dict {
		s := float64(v) / dis
		c, ok := sc.maxSpeed[k]
		if ok {
			if s > c && v > 100 { //防止过少不精确
				sc.maxSpeed[k] = s
			}
		} else {
			sc.maxSpeed[k] = s
		}
		log.Info("Speed", "bcname", sc.prefix, "flag", k, "cnt", v, "cost", dis, "qps[/s]", s, "max", sc.maxSpeed[k])
	}
	log.Info("Probe", "number of goroutines", runtime.NumGoroutine())
}

// ShowLoop show the speed with a loop
func (sc *SpeedCalc) ShowLoop(log log.Logger) {
	cnt := 0
	for {
		sc.ShowInfo(log)
		time.Sleep(1e9)
		cnt++
		if cnt > 60 {
			cnt = 0
			sc.Clear()
		}
	}
}
