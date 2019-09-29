package dht

import "github.com/libp2p/go-libp2p-core/routing"

type quorumOptionKey struct{}

const defaultQuorum = 16

// Quorum is a DHT option that tells the DHT how many peers it needs to get
// values from before returning the best one.
//
// Default: 16
func Quorum(n int) routing.Option {
	return func(opts *routing.Options) error {
		if opts.Other == nil {
			opts.Other = make(map[interface{}]interface{}, 1)
		}
		opts.Other[quorumOptionKey{}] = n
		return nil
	}
}

func getQuorum(opts *routing.Options, ndefault int) int {
	responsesNeeded, ok := opts.Other[quorumOptionKey{}].(int)
	if !ok {
		responsesNeeded = ndefault
	}
	return responsesNeeded
}
