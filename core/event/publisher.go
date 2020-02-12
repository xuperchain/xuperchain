// import "github.com/moby/moby/pkg/pubsub"
// Use of this source code is governed under the Apache License, Version 2.0
//
// ~~ 2020.02.07
// Modified by Baidu.Inc. inorder to support customized event mechanism
package event

import (
	"sync"
	"time"

	"github.com/xuperchain/xuperchain/core/pb"
)

var wgPool = sync.Pool{New: func() interface{} { return new(sync.WaitGroup) }}

// NewPublisher creates a new pub/sub publisher to broadcast messages.
// The duration is used as the send timeout as to not block the publisher publishing
// messages to other clients if one client is slow or unresponsive.
// The buffer is used when creating new channels for subscribers.
func NewPublisher(publishTimeout time.Duration, buffer int) *Publisher {
	return &Publisher{
		buffer:      buffer,
		timeout:     publishTimeout,
		subscribers: make(map[subscriber]topicFunc),
	}
}

type subscriber chan *pb.Event
type topicFunc func(v *pb.Event) bool

// Publisher is basic pub/sub structure. Allows to send events and subscribe
// to them. Can be safely used from multiple goroutines.
type Publisher struct {
	m           sync.RWMutex
	buffer      int
	timeout     time.Duration
	subscribers map[subscriber]topicFunc
}

// Len returns the number of subscribers for the publisher
func (p *Publisher) Len() int {
	p.m.RLock()
	defer p.m.RUnlock()
	i := len(p.subscribers)
	return i
}

// Subscribe adds a new subscriber to the publisher returning the channel.
func (p *Publisher) Subscribe() chan *pb.Event {
	return p.SubscribeTopic(nil)
}

// SubscribeTopic adds a new subscriber that filters messages sent by a topic.
func (p *Publisher) SubscribeTopic(topic topicFunc) chan *pb.Event {
	ch := make(chan *pb.Event, p.buffer)
	p.m.Lock()
	defer p.m.Unlock()
	p.subscribers[ch] = topic
	return ch
}

// SubscribeTopicWithBuffer adds a new subscriber that filters messages sent by a topic.
// The returned channel has a buffer of the specified size.
func (p *Publisher) SubscribeTopicWithBuffer(topic topicFunc, buffer int) chan *pb.Event {
	ch := make(chan *pb.Event, buffer)
	p.m.Lock()
	defer p.m.Unlock()
	p.subscribers[ch] = topic
	return ch
}

// Evict removes the specified subscriber from receiving any more messages.
func (p *Publisher) Evict(sub chan *pb.Event) bool {
	p.m.Lock()
	defer p.m.Unlock()
	if _, exist := p.subscribers[sub]; !exist {
		return false
	}
	delete(p.subscribers, sub)
	close(sub)

	return true
}

// Publish sends the data in v to all subscribers currently registered with the publisher.
func (p *Publisher) Publish(v *pb.Event) {
	p.m.RLock()
	defer p.m.RUnlock()
	if len(p.subscribers) == 0 {
		return
	}

	wg := wgPool.Get().(*sync.WaitGroup)
	for sub, topic := range p.subscribers {
		wg.Add(1)
		go p.sendTopic(sub, topic, v, wg)
	}
	wg.Wait()
	wgPool.Put(wg)
}

// Close closes the channels to all subscribers registered with the publisher.
func (p *Publisher) Close() {
	p.m.Lock()
	defer p.m.Unlock()
	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}

func (p *Publisher) sendTopic(sub subscriber, topic topicFunc, v *pb.Event, wg *sync.WaitGroup) {
	defer wg.Done()
	if topic != nil && !topic(v) {
		return
	}

	// send under a select as to not block if the receiver is unavailable
	if p.timeout > 0 {
		timeout := time.NewTimer(p.timeout)
		defer timeout.Stop()

		select {
		case sub <- v:
		case <-timeout.C:
		}
		return
	}

	select {
	case sub <- v:
	default:
	}
}
