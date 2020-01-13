package eventbus

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/libp2p/go-libp2p-core/event"
)

///////////////////////
// BUS

// basicBus is a type-based event delivery system
type basicBus struct {
	lk    sync.Mutex
	nodes map[reflect.Type]*node
}

var _ event.Bus = (*basicBus)(nil)

type emitter struct {
	n       *node
	typ     reflect.Type
	closed  int32
	dropper func(reflect.Type)
}

func (e *emitter) Emit(evt interface{}) error {
	if atomic.LoadInt32(&e.closed) != 0 {
		return fmt.Errorf("emitter is closed")
	}
	e.n.emit(evt)
	return nil
}

func (e *emitter) Close() error {
	if !atomic.CompareAndSwapInt32(&e.closed, 0, 1) {
		return fmt.Errorf("closed an emitter more than once")
	}
	if atomic.AddInt32(&e.n.nEmitters, -1) == 0 {
		e.dropper(e.typ)
	}
	return nil
}

func NewBus() event.Bus {
	return &basicBus{
		nodes: map[reflect.Type]*node{},
	}
}

func (b *basicBus) withNode(typ reflect.Type, cb func(*node), async func(*node)) {
	b.lk.Lock()

	n, ok := b.nodes[typ]
	if !ok {
		n = newNode(typ)
		b.nodes[typ] = n
	}

	n.lk.Lock()
	b.lk.Unlock()

	cb(n)

	if async == nil {
		n.lk.Unlock()
	} else {
		go func() {
			defer n.lk.Unlock()
			async(n)
		}()
	}
}

func (b *basicBus) tryDropNode(typ reflect.Type) {
	b.lk.Lock()
	n, ok := b.nodes[typ]
	if !ok { // already dropped
		b.lk.Unlock()
		return
	}

	n.lk.Lock()
	if atomic.LoadInt32(&n.nEmitters) > 0 || len(n.sinks) > 0 {
		n.lk.Unlock()
		b.lk.Unlock()
		return // still in use
	}
	n.lk.Unlock()

	delete(b.nodes, typ)
	b.lk.Unlock()
}

type sub struct {
	ch      chan interface{}
	nodes   []*node
	dropper func(reflect.Type)
}

func (s *sub) Out() <-chan interface{} {
	return s.ch
}

func (s *sub) Close() error {
	go func() {
		// drain the event channel, will return when closed and drained.
		// this is necessary to unblock publishes to this channel.
		for range s.ch {
		}
	}()

	for _, n := range s.nodes {
		n.lk.Lock()

		for i := 0; i < len(n.sinks); i++ {
			if n.sinks[i] == s.ch {
				n.sinks[i], n.sinks[len(n.sinks)-1] = n.sinks[len(n.sinks)-1], nil
				n.sinks = n.sinks[:len(n.sinks)-1]
				break
			}
		}

		tryDrop := len(n.sinks) == 0 && atomic.LoadInt32(&n.nEmitters) == 0

		n.lk.Unlock()

		if tryDrop {
			s.dropper(n.typ)
		}
	}
	close(s.ch)
	return nil
}

var _ event.Subscription = (*sub)(nil)

// Subscribe creates new subscription. Failing to drain the channel will cause
// publishers to get blocked. CancelFunc is guaranteed to return after last send
// to the channel
func (b *basicBus) Subscribe(evtTypes interface{}, opts ...event.SubscriptionOpt) (_ event.Subscription, err error) {
	settings := subSettings(subSettingsDefault)
	for _, opt := range opts {
		if err := opt(&settings); err != nil {
			return nil, err
		}
	}

	types, ok := evtTypes.([]interface{})
	if !ok {
		types = []interface{}{evtTypes}
	}

	out := &sub{
		ch:    make(chan interface{}, settings.buffer),
		nodes: make([]*node, len(types)),

		dropper: b.tryDropNode,
	}

	for _, etyp := range types {
		if reflect.TypeOf(etyp).Kind() != reflect.Ptr {
			return nil, errors.New("subscribe called with non-pointer type")
		}
	}

	for i, etyp := range types {
		typ := reflect.TypeOf(etyp)

		b.withNode(typ.Elem(), func(n *node) {
			n.sinks = append(n.sinks, out.ch)
			out.nodes[i] = n
		}, func(n *node) {
			if n.keepLast {
				l := n.last
				if l == nil {
					return
				}
				out.ch <- l
			}
		})
	}

	return out, nil
}

// Emitter creates new emitter
//
// eventType accepts typed nil pointers, and uses the type information to
// select output type
//
// Example:
// emit, err := eventbus.Emitter(new(EventT))
// defer emit.Close() // MUST call this after being done with the emitter
//
// emit(EventT{})
func (b *basicBus) Emitter(evtType interface{}, opts ...event.EmitterOpt) (e event.Emitter, err error) {
	var settings emitterSettings

	for _, opt := range opts {
		if err := opt(&settings); err != nil {
			return nil, err
		}
	}

	typ := reflect.TypeOf(evtType)
	if typ.Kind() != reflect.Ptr {
		return nil, errors.New("emitter called with non-pointer type")
	}
	typ = typ.Elem()

	b.withNode(typ, func(n *node) {
		atomic.AddInt32(&n.nEmitters, 1)
		n.keepLast = n.keepLast || settings.makeStateful
		e = &emitter{n: n, typ: typ, dropper: b.tryDropNode}
	}, nil)
	return
}

///////////////////////
// NODE

type node struct {
	// Note: make sure to NEVER lock basicBus.lk when this lock is held
	lk sync.Mutex

	typ reflect.Type

	// emitter ref count
	nEmitters int32

	keepLast bool
	last     interface{}

	sinks []chan interface{}
}

func newNode(typ reflect.Type) *node {
	return &node{
		typ: typ,
	}
}

func (n *node) emit(event interface{}) {
	typ := reflect.TypeOf(event)
	if typ != n.typ {
		panic(fmt.Sprintf("Emit called with wrong type. expected: %s, got: %s", n.typ, typ))
	}

	n.lk.Lock()
	if n.keepLast {
		n.last = event
	}

	for _, ch := range n.sinks {
		ch <- event
	}
	n.lk.Unlock()
}
