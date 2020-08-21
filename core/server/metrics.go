package server

import prom "github.com/prometheus/client_golang/prometheus"

var (
	// DefaultServerMetrics is the default instance of server metrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultServerMetrics = newServerMetrics()
)

func init() {
	prom.MustRegister(DefaultServerMetrics.rpcFlowIn)
	prom.MustRegister(DefaultServerMetrics.rpcFlowOut)
}

type serverMetrics struct {
	rpcFlowIn  *prom.CounterVec
	rpcFlowOut *prom.CounterVec
}

func newServerMetrics() *serverMetrics {
	return &serverMetrics{
		rpcFlowIn: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "rpc_flow_in",
				Help: "Current flow in of rpc server",
			},
			[]string{"bcname", "type"}),
		rpcFlowOut: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "rpc_flow_out",
				Help: "Current flow  out of rpc server",
			},
			[]string{"bcname", "type"}),
	}
}
