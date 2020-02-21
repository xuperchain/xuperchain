package base

type MsgOptions struct {
	Filters         []FilterStrategy
	Bcname          string
	TargetPeerAddrs []string
	TargetPeerIDs   []string
	Percentage      float32 //percentage wait for return
	// compress
	Compress  bool
	WhiteList map[string]bool
}

// MessageOption define single option function
type MessageOption func(*MsgOptions)

func WithWhiteList(whiteList map[string]bool) MessageOption {
	return func(o *MsgOptions) {
		o.WhiteList = whiteList
	}
}

// WithFilters add filter strategies to message option
func WithFilters(filter []FilterStrategy) MessageOption {
	return func(o *MsgOptions) {
		o.Filters = filter
	}
}

// WithBcName add bcname to message option
func WithBcName(bcname string) MessageOption {
	return func(o *MsgOptions) {
		o.Bcname = bcname
	}
}

// WithCompress set compredded to message option
func WithCompress(compress bool) MessageOption {
	return func(o *MsgOptions) {
		o.Compress = compress
	}
}

// WithPercentage add percentage to message option
func WithPercentage(percentage float32) MessageOption {
	return func(o *MsgOptions) {
		o.Percentage = percentage
	}
}

// WithTargetPeerAddrs add target peer addresses to message option
func WithTargetPeerAddrs(peerAddrs []string) MessageOption {
	return func(o *MsgOptions) {
		o.TargetPeerAddrs = peerAddrs
	}
}

// WithTargetPeerIDs add target peer IDs to message option
func WithTargetPeerIDs(pid []string) MessageOption {
	return func(o *MsgOptions) {
		o.TargetPeerIDs = pid
	}
}

// GetMessageOption create MessageOptions with given options
func GetMessageOption(opts []MessageOption) *MsgOptions {
	msgOpts := &MsgOptions{
		Percentage: 1,
	}
	for _, f := range opts {
		f(msgOpts)
	}
	if msgOpts.Percentage > 1 {
		msgOpts.Percentage = 1
	}
	return msgOpts
}
