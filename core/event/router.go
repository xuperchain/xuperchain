package event

import (
	"fmt"

	xchaincore "github.com/xuperchain/xuperchain/core/core"
	"github.com/xuperchain/xuperchain/core/pb"
)

type Router struct {
	topics map[pb.SubscribeType]Topic
}

func NewRounterFromChainMG(chainmg ChainManager) *Router {
	blockTopic := NewBlockTopic(chainmg)
	r := &Router{
		topics: make(map[pb.SubscribeType]Topic),
	}
	r.topics[pb.SubscribeType_BLOCK] = blockTopic

	return r
}

func NewRouter(chainmg *xchaincore.XChainMG) *Router {
	return NewRounterFromChainMG(NewChainManager(chainmg))
}

func (r *Router) Subscribe(tp pb.SubscribeType, filterbuf []byte) (Iterator, error) {
	topic, ok := r.topics[tp]
	if !ok {
		return nil, fmt.Errorf("subscribe type %s unsupported", tp)
	}
	filter, err := topic.ParseFilter(filterbuf)
	if err != nil {
		return nil, fmt.Errorf("parse filter error: %s", err)
	}
	return topic.NewIterator(filter)
}
