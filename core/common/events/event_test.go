package events

import (
	"testing"
)

var counter int

func Add(e *EventMessage) {
	switch e.Message.(type) {
	case int:
		delta := e.Message.(int)
		counter += delta
	}
}

type TestStruct struct {
	Counter int
}

func (ts *TestStruct) Add(e *EventMessage) {
	switch e.Message.(type) {
	case int:
		delta := e.Message.(int)
		ts.Counter += delta
	}
}

var ts TestStruct

func TestSubscribe(t *testing.T) {
	ec := GetEventBus()
	counter = 0

	err := ec.Subscribe(ProposerReady, Add)
	if err != nil {
		t.Fatalf("subscribe failed, err=%v\n", err)
	}

	err = ec.Subscribe(ProposerReady, ts.Add)
	if err != nil {
		t.Fatalf("subscribe failed, err=%v\n", err)
	}

	// duplicate subscribe
	err = ec.Subscribe(ProposerReady, Add)
	if err == nil || err != ErrDuplicateHandler {
		t.Fatalf("dup subscribe failed, err=%v\n", err)
	}

	em := &EventMessage{
		BcName:   "test",
		Type:     ProposerReady,
		Message:  6,
		Priority: 0,
		Sender:   nil,
	}

	err = ec.FireEvent(em)
	if err != nil {
		t.Fatalf("subscribe failed, err=%v\n", err)
	}

	if counter != 6 || ts.Counter != 6 {
		t.Fatalf("event fired but result is wrong, expect=6, actual=%d\n", counter)
	}
}

func TestUnsubscribe(t *testing.T) {
	ec := GetEventBus()
	counter = 0
	ts.Counter = 0
	em := &EventMessage{
		BcName:   "test",
		Type:     ProposerReady,
		Message:  6,
		Priority: 0,
		Sender:   nil,
	}

	err := ec.FireEvent(em)
	if err != nil {
		t.Fatalf("FireEvent failed, err=%v\n", err)
	}

	if counter != 6 {
		t.Fatalf("event fired but result is wrong, expect=6, actual=%d\n", counter)
	}

	if ts.Counter != 6 {
		t.Fatalf("event fired but result is wrong, expect=6, actual=%d\n", ts.Counter)
	}

	err = ec.Unsubscribe(ProposerReady, Add)
	if err != nil {
		t.Fatalf("unsubscribe failed, err=%v\n", err)
	}

	err = ec.FireEvent(em)
	if err != nil {
		t.Fatalf("FireEvent failed, err=%v\n", err)
	}

	if counter != 6 {
		t.Fatalf("event fired but result is wrong, expect=6, actual=%d\n", counter)
	}

	if ts.Counter != 12 {
		t.Fatalf("event fired but result is wrong, expect=12, actual=%d\n", ts.Counter)
	}
}

func TestFireEventAsync(t *testing.T) {
	counter = 0
	ts.Counter = 0

	ec := GetEventBus()
	em := &EventMessage{
		BcName:   "test",
		Type:     ProposerReady,
		Message:  6,
		Priority: 0,
		Sender:   nil,
	}

	err := ec.Subscribe(ProposerReady, Add)
	if err != nil {
		t.Fatalf("subscribe failed, err=%v\n", err)
	}

	wg, err := ec.FireEventAsync(em)
	if err != nil {
		t.Fatalf("FireEventAsync failed, err=%v\n", err)
	}

	wg.Wait()
	if counter != 6 {
		t.Fatalf("event fired but result is wrong, expect=6, actual=%d\n", counter)
	}

	if ts.Counter != 6 {
		t.Fatalf("event fired but result is wrong, expect=6, actual=%d\n", ts.Counter)
	}
}
