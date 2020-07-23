package utxo

import (
	"testing"
	"time"
)

func TestNotify(t *testing.T) {
	notifier := NewBlockHeightNotifier()
	closedch := make(chan struct{})

	var height int64
	go func() {
		height = notifier.WaitHeight(10)
		close(closedch)
	}()

	notifier.UpdateHeight(10)
	select {
	case <-time.After(2 * time.Second):
		t.Fatal("wait timeout")
	case <-closedch:
	}

	if height != 10 {
		t.Fatalf("expect 10 got %d", height)
	}
}
