package p2pv2

type msgOptions struct {
	filters         []FilterStrategy
	bcname          string
	targetPeerAddrs []string
	percentage      float32 //percentage wait for return
	// compressed
	compressed bool
}

// MessageOption define single option function
type MessageOption func(*msgOptions)

// WithFilters add filter strategies to message option
func WithFilters(filter []FilterStrategy) MessageOption {
	return func(o *msgOptions) {
		o.filters = filter
	}
}

// WithBcName add bcname to message option
func WithBcName(bcname string) MessageOption {
	return func(o *msgOptions) {
		o.bcname = bcname
	}
}

// WithCompressed set compredded to message option
func WithCompressed(compressed bool) MessageOption {
	return func(o *msgOptions) {
		o.compressed = compressed
	}
}

// WithPercentage add percentage to message option
func WithPercentage(percentage float32) MessageOption {
	return func(o *msgOptions) {
		o.percentage = percentage
	}
}

// WithTargetPeerAddrs add target peer addresses to message option
func WithTargetPeerAddrs(peerAddrs []string) MessageOption {
	return func(o *msgOptions) {
		o.targetPeerAddrs = peerAddrs
	}
}

// getMessageOption create MessageOptions with given options
func getMessageOption(opts []MessageOption) *msgOptions {
	msgOpts := &msgOptions{
		percentage: 1,
		filters:    []FilterStrategy{DefaultStrategy},
	}
	for _, f := range opts {
		f(msgOpts)
	}
	if msgOpts.percentage > 1 {
		msgOpts.percentage = 1
	}
	return msgOpts
}
