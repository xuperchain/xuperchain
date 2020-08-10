package server

import prom "github.com/prometheus/client_golang/prometheus"

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
)

func init() {
	prom.MustRegister(rpcFlowIn)
	prom.MustRegister(rpcFlowOut)
}

