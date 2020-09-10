package base

import prom "github.com/prometheus/client_golang/prometheus"

var (
	// DefaultP2pMetrics is the default instance of p2pMetrics. It is
	// intended to be used in conjunction the default Prometheus metrics
	// registry.
	DefaultP2pMetrics = newP2pMetrics()
)

var (
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

// p2pMetrics is the metrics of p2p server
type p2pMetrics struct {
	P2PFlowIn  *prom.CounterVec
	P2PFlowOut *prom.CounterVec
}

// newP2pMetrics return
func newP2pMetrics() *p2pMetrics {
	return &p2pMetrics{
		P2PFlowIn:  p2pFlowIn,
		P2PFlowOut: p2pFlowOut,
	}
}

func init() {
	prom.MustRegister(p2pFlowIn)
	prom.MustRegister(p2pFlowOut)
}
