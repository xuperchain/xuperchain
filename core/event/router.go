package event

import (
	"fmt"

	xchaincore "github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/pb"
)

// Router distribute events according to the event type and filter
type Router struct {
	topics map[pb.SubscribeType]Topic
}

// NewRounterFromChainMG instance Router from ChainManager
func NewRounterFromChainMG(chainmg ChainManager) *Router {
	blockTopic := NewBlockTopic(chainmg)
	r := &Router{
		topics: make(map[pb.SubscribeType]Topic),
	}
	r.topics[pb.SubscribeType_BLOCK] = blockTopic

	return r
}

// NewRounterFromChainMG instance Router from xchaincore.XChainMG
func NewRouter(chainmg *xchaincore.XChainMG) *Router {
	return NewRounterFromChainMG(NewChainManager(chainmg))
}

// EncodeFunc encodes event payload
type EncodeFunc func(x interface{}) ([]byte, error)

// Subscribe route events from pb.SubscribeType and filter buffer
func (r *Router) Subscribe(tp pb.SubscribeType, filterbuf []byte) (EncodeFunc, Iterator, error) {
	topic, ok := r.topics[tp]
	if !ok {
		return nil, nil, fmt.Errorf("subscribe type %s unsupported", tp)
	}
	filter, err := topic.ParseFilter(filterbuf)
	if err != nil {
		return nil, nil, fmt.Errorf("parse filter error: %s", err)
	}
	iter, err := topic.NewIterator(filter)
	return topic.MarshalEvent, iter, err
}

// RawSubscribe route events from pb.SubscribeType and filter struct
func (r *Router) RawSubscribe(tp pb.SubscribeType, filter interface{}) (Iterator, error) {
	topic, ok := r.topics[tp]
	if !ok {
		return nil, fmt.Errorf("subscribe type %s unsupported", tp)
	}
	iter, err := topic.NewIterator(filter)
	return iter, err
}
