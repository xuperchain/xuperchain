package poa

import (
	"fmt"
	"sync"
	"time"
)

type MyTimer struct {
	Name      string
	Delay     uint64
	Target    func()

	Timer     *time.Timer
	Running   bool
	RunningMu sync.Mutex
}

func NewMyTimer(name string, delay uint64, target func()) *MyTimer {
	newTimer := &MyTimer{
		Name:    name,
		Delay:   delay,
		Target:  target,
		Timer:   new(time.Timer),
		Running: false,
	}
	return newTimer
}

func (mt *MyTimer) Start(delay uint64) {
	mt.RunningMu.Lock()
	defer func() {
		mt.RunningMu.Unlock()
		mt.Running = true
	}()

	if delay == uint64(0) {
		fmt.Println(mt.Name, "start, ", mt.Delay, " seconds later call target")
		mt.Timer = time.AfterFunc(time.Duration(mt.Delay)*time.Second, mt.Target)
	} else {
		fmt.Println(mt.Name, "start, ", delay, " seconds later call target")
		mt.Timer = time.AfterFunc(time.Duration(delay)*time.Second, mt.Target)
	}
}

func (mt *MyTimer) Stop() bool {
	fmt.Println(mt.Name, " end.")

	mt.RunningMu.Lock()
	defer func() {
		mt.RunningMu.Unlock()
		mt.Running = false
	}()

	if mt.Running {
		return mt.Timer.Stop()
	} else {
		return false
	}
}

func (mt *MyTimer) Reset(delay uint64) {
	mt.Stop()
	mt.Start(delay)
}
