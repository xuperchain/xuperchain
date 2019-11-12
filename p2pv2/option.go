package p2pv2

type msgOptions struct {
	filters         []FilterStrategy
	bcname          string
	targetPeerAddrs []string
	targetPeerIDs   []string
	percentage      float32 //percentage wait for return
	// compress
	compress bool
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

// WithCompress set compredded to message option
func WithCompress(compress bool) MessageOption {
	return func(o *msgOptions) {
		o.compress = compress
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

// WithTargetPeerIDs add target peer IDs to message option
func WithTargetPeerIDs(pid []string) MessageOption {
	return func(o *msgOptions) {
		o.targetPeerIDs = pid
	}
}

// getMessageOption create MessageOptions with given options
func getMessageOption(opts []MessageOption) *msgOptions {
	msgOpts := &msgOptions{
		percentage: 1,
	}
	for _, f := range opts {
		f(msgOpts)
	}
	if msgOpts.percentage > 1 {
		msgOpts.percentage = 1
	}
	return msgOpts
}
