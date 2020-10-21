// This package was extracted from Tendermint
//
// Package pubsub implements a pub-sub model with a single publisher (Server)
// and multiple subscribers (clients).
//
// Though you can have multiple publishers by sharing a pointer to a server or
// by giving the same channel to each publisher and publishing messages from
// that channel (fan-in).
//
// Clients subscribe for messages, which could be of any type, using a query.
// When some message is published, we match it with all queries. If there is a
// match, this message will be pushed to all clients, subscribed to that query.
// See query subpackage for our implementation.
package pubsub

import (
	"context"
	"errors"
	"sync"

	"github.com/hyperledger/burrow/event/query"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/tendermint/tendermint/libs/service"
)

type operation int

const (
	sub operation = iota
	pub
	unsub
	shutdown
)

var (
	// ErrSubscriptionNotFound is returned when a client tries to unsubscribe
	// from not existing subscription.
	ErrSubscriptionNotFound = errors.New("subscription not found")

	// ErrAlreadySubscribed is returned when a client tries to subscribe twice or
	// more using the same query.
	ErrAlreadySubscribed = errors.New("already subscribed")
)

type cmd struct {
	op       operation
	query    query.Query
	ch       chan interface{}
	clientID string
	msg      interface{}
	tags     query.Tagged
}

// Server allows clients to subscribe/unsubscribe for messages, publishing
// messages with or without tags, and manages internal state.
type Server struct {
	service.BaseService

	cmds    chan cmd
	cmdsCap int

	mtx           sync.RWMutex
	subscriptions map[string]map[string]query.Query // subscriber -> query (string) -> query.Query
	logger        *logging.Logger
}

// Option sets a parameter for the server.
type Option func(*Server)

// NewServer returns a new server. See the commentary on the Option functions
// for a detailed description of how to configure buffering. If no options are
// provided, the resulting server's queue is unbuffered.
func NewServer(options ...Option) *Server {
	s := &Server{
		subscriptions: make(map[string]map[string]query.Query),
		logger:        logging.NewNoopLogger(),
	}
	s.BaseService = *service.NewBaseService(nil, "PubSub", s)

	for _, option := range options {
		option(s)
	}

	// if BufferCapacity option was not set, the channel is unbuffered
	s.cmds = make(chan cmd, s.cmdsCap)

	return s
}

// BufferCapacity allows you to specify capacity for the internal server's
// queue. Since the server, given Y subscribers, could only process X messages,
// this option could be used to survive spikes (e.g. high amount of
// transactions during peak hours).
func BufferCapacity(cap int) Option {
	return func(s *Server) {
		if cap > 0 {
			s.cmdsCap = cap
		}
	}
}

func WithLogger(logger *logging.Logger) Option {
	return func(s *Server) {
		s.logger = logger.WithScope("PubSub")
	}
}

// BufferCapacity returns capacity of the internal server's queue.
func (s *Server) BufferCapacity() int {
	return s.cmdsCap
}

// Subscribe creates a subscription for the given client. It accepts a channel
// on which messages matching the given query can be received. An error will be
// returned to the caller if the context is canceled or if subscription already
// exist for pair clientID and query.
func (s *Server) Subscribe(ctx context.Context, clientID string, qry query.Query, outBuffer int) (<-chan interface{}, error) {
	s.mtx.RLock()
	clientSubscriptions, ok := s.subscriptions[clientID]
	if ok {
		_, ok = clientSubscriptions[qry.String()]
	}
	s.mtx.RUnlock()
	if ok {
		return nil, ErrAlreadySubscribed
	}
	// We are responsible for closing this channel so we create it
	out := make(chan interface{}, outBuffer)
	select {
	case s.cmds <- cmd{op: sub, clientID: clientID, query: qry, ch: out}:
		s.mtx.Lock()
		if _, ok = s.subscriptions[clientID]; !ok {
			s.subscriptions[clientID] = make(map[string]query.Query)
		}
		// preserve original query
		// see Unsubscribe
		s.subscriptions[clientID][qry.String()] = qry
		s.mtx.Unlock()
		return out, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Unsubscribe removes the subscription on the given query. An error will be
// returned to the caller if the context is canceled or if subscription does
// not exist.
func (s *Server) Unsubscribe(ctx context.Context, clientID string, qry query.Query) error {
	var origQuery query.Query
	s.mtx.RLock()
	clientSubscriptions, ok := s.subscriptions[clientID]
	if ok {
		origQuery, ok = clientSubscriptions[qry.String()]
	}
	s.mtx.RUnlock()
	if !ok {
		return ErrSubscriptionNotFound
	}

	// original query is used here because we're using pointers as map keys
	select {
	case s.cmds <- cmd{op: unsub, clientID: clientID, query: origQuery}:
		s.mtx.Lock()
		delete(clientSubscriptions, qry.String())
		s.mtx.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// UnsubscribeAll removes all client subscriptions. An error will be returned
// to the caller if the context is canceled or if subscription does not exist.
func (s *Server) UnsubscribeAll(ctx context.Context, clientID string) error {
	s.mtx.RLock()
	_, ok := s.subscriptions[clientID]
	s.mtx.RUnlock()
	if !ok {
		return ErrSubscriptionNotFound
	}

	select {
	case s.cmds <- cmd{op: unsub, clientID: clientID}:
		s.mtx.Lock()
		delete(s.subscriptions, clientID)
		s.mtx.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Publish publishes the given message. An error will be returned to the caller
// if the context is canceled.
func (s *Server) Publish(ctx context.Context, msg interface{}) error {
	return s.PublishWithTags(ctx, msg, query.TagMap(make(map[string]interface{})))
}

// PublishWithTags publishes the given message with the set of tags. The set is
// matched with clients queries. If there is a match, the message is sent to
// the client.
func (s *Server) PublishWithTags(ctx context.Context, msg interface{}, tags query.Tagged) error {
	select {
	case s.cmds <- cmd{op: pub, msg: msg, tags: tags}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// OnStop implements Service.OnStop by shutting down the server.
func (s *Server) OnStop() {
	s.cmds <- cmd{op: shutdown}
}

// NOTE: not goroutine safe
type state struct {
	// query -> client -> ch
	queries map[query.Query]map[string]chan interface{}
	// client -> query -> struct{}
	clients map[string]map[query.Query]struct{}
	logger  *logging.Logger
}

// OnStart implements Service.OnStart by starting the server.
func (s *Server) OnStart() error {
	go s.loop(state{
		queries: make(map[query.Query]map[string]chan interface{}),
		clients: make(map[string]map[query.Query]struct{}),
		logger:  s.logger,
	})
	return nil
}

// OnReset implements Service.OnReset
func (s *Server) OnReset() error {
	return nil
}

func (s *Server) loop(state state) {
loop:
	for cmd := range s.cmds {
		switch cmd.op {
		case unsub:
			if cmd.query != nil {
				state.remove(cmd.clientID, cmd.query)
			} else {
				state.removeAll(cmd.clientID)
			}
		case shutdown:
			for clientID := range state.clients {
				state.removeAll(clientID)
			}
			break loop
		case sub:
			state.add(cmd.clientID, cmd.query, cmd.ch)
		case pub:
			state.send(cmd.msg, cmd.tags)
		}
	}
}

func (state *state) add(clientID string, q query.Query, ch chan interface{}) {
	// add query if needed
	if _, ok := state.queries[q]; !ok {
		state.queries[q] = make(map[string]chan interface{})
	}

	// create subscription
	state.queries[q][clientID] = ch

	// add client if needed
	if _, ok := state.clients[clientID]; !ok {
		state.clients[clientID] = make(map[query.Query]struct{})
	}
	state.clients[clientID][q] = struct{}{}
}

func (state *state) remove(clientID string, q query.Query) {
	clientToChannelMap, ok := state.queries[q]
	if !ok {
		return
	}

	ch, ok := clientToChannelMap[clientID]
	if ok {
		closeAndDrain(ch)

		delete(state.clients[clientID], q)

		// if it not subscribed to anything else, remove the client
		if len(state.clients[clientID]) == 0 {
			delete(state.clients, clientID)
		}

		delete(state.queries[q], clientID)
		if len(state.queries[q]) == 0 {
			delete(state.queries, q)
		}
	}
}

func (state *state) removeAll(clientID string) {
	queryMap, ok := state.clients[clientID]
	if !ok {
		return
	}

	for q := range queryMap {
		ch := state.queries[q][clientID]
		closeAndDrain(ch)

		delete(state.queries[q], clientID)
		if len(state.queries[q]) == 0 {
			delete(state.queries, q)
		}
	}
	delete(state.clients, clientID)
}

func closeAndDrain(ch chan interface{}) {
	close(ch)
	for range ch {
	}
}

func (state *state) send(msg interface{}, tags query.Tagged) {
	for q, clientToChannelMap := range state.queries {
		if q.Matches(tags) {
			for _, ch := range clientToChannelMap {
				select {
				case ch <- msg:
				default:
					// It's difficult to do anything sensible here with retries/times outs since we may reorder a client's
					// view of events by sending a later message before an earlier message we retry. If per-client order
					// matters then we need a queue per client. Possible for us it does not...
				}
			}
		}
		err := q.MatchError()
		if err != nil {
			state.logger.InfoMsg("pubsub Server could not execute query", structure.ErrorKey, err)
		}
	}
}
