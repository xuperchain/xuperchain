package notifications

import (
	"context"
	"encoding/json"
	"sync"

	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

type QueryEventType int

// Number of events to buffer.
var QueryEventBufferSize = 16

const (
	SendingQuery QueryEventType = iota
	PeerResponse
	FinalPeer
	QueryError
	Provider
	Value
	AddingPeer
	DialingPeer
)

type QueryEvent struct {
	ID        peer.ID
	Type      QueryEventType
	Responses []*pstore.PeerInfo
	Extra     string
}

type routingQueryKey struct{}
type eventChannel struct {
	mu  sync.Mutex
	ctx context.Context
	ch  chan<- *QueryEvent
}

// waitThenClose is spawned in a goroutine when the channel is registered. This
// safely cleans up the channel when the context has been canceled.
func (e *eventChannel) waitThenClose() {
	<-e.ctx.Done()
	e.mu.Lock()
	close(e.ch)
	// 1. Signals that we're done.
	// 2. Frees memory (in case we end up hanging on to this for a while).
	e.ch = nil
	e.mu.Unlock()
}

// send sends an event on the event channel, aborting if either the passed or
// the internal context expire.
func (e *eventChannel) send(ctx context.Context, ev *QueryEvent) {
	e.mu.Lock()
	// Closed.
	if e.ch == nil {
		e.mu.Unlock()
		return
	}
	// in case the passed context is unrelated, wait on both.
	select {
	case e.ch <- ev:
	case <-e.ctx.Done():
	case <-ctx.Done():
	}
	e.mu.Unlock()
}

func RegisterForQueryEvents(ctx context.Context) (context.Context, <-chan *QueryEvent) {
	ch := make(chan *QueryEvent, QueryEventBufferSize)
	ech := &eventChannel{ch: ch, ctx: ctx}
	go ech.waitThenClose()
	return context.WithValue(ctx, routingQueryKey{}, ech), ch
}

func PublishQueryEvent(ctx context.Context, ev *QueryEvent) {
	ich := ctx.Value(routingQueryKey{})
	if ich == nil {
		return
	}

	// We *want* to panic here.
	ech := ich.(*eventChannel)
	ech.send(ctx, ev)
}

func (qe *QueryEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"ID":        peer.IDB58Encode(qe.ID),
		"Type":      int(qe.Type),
		"Responses": qe.Responses,
		"Extra":     qe.Extra,
	})
}

func (qe *QueryEvent) UnmarshalJSON(b []byte) error {
	temp := struct {
		ID        string
		Type      int
		Responses []*pstore.PeerInfo
		Extra     string
	}{}
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}
	if len(temp.ID) > 0 {
		pid, err := peer.IDB58Decode(temp.ID)
		if err != nil {
			return err
		}
		qe.ID = pid
	}
	qe.Type = QueryEventType(temp.Type)
	qe.Responses = temp.Responses
	qe.Extra = temp.Extra
	return nil
}
