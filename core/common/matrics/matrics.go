package matrics

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	// DefaultServerMetrics is the default instance of ServerMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultServerMetrics = NewServerMetrics()
)

var (
	rpcFlowIn = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "rpc_flow_in",
			Help: "Current flow in of rpc server",
		},
		[]string{"bcname", "type"})
	rpcFlowOut = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "rpc_flow_out",
			Help: "Current flow  out of rpc server",
		},
		[]string{"bcname", "type"})
	p2pFlowIn = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "p2p_flow_in",
			Help: "Current flow in of p2p server",
		},
		[]string{"bcname", "type"})
	p2pFlowOut = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "p2p_flow_out",
			Help: "Current flow out of p2p server",
		},
		[]string{"bcname", "type"})
)

// ServerMetrics is the collection metrics of xchain server to be registered on Prometheus.
// It's include rpc metrics, p2p metrics and others system metrics.
type ServerMetrics struct {
	RPCFlowIn  *prom.CounterVec
	RPCFlowOut *prom.CounterVec
	P2PFlowIn  *prom.CounterVec
	P2PFlowOut *prom.CounterVec
}

// NewServerMetrics return mertrics of server
func NewServerMetrics() *ServerMetrics {
	return &ServerMetrics{
		RPCFlowIn:  rpcFlowIn,
		RPCFlowOut: rpcFlowOut,
		P2PFlowIn:  p2pFlowIn,
		P2PFlowOut: p2pFlowOut,
	}
}
