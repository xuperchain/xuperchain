package main

import (
	"context"
	"flag"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/xuperchain/xuperchain/core/pb"
	"google.golang.org/grpc"
	"net/http"
)

var (
	rpcEndpoint = flag.String("gateway_endpoint", "localhost:37101", "endpoint of grpc service forward to")
	// http port
	httpEndpoint = flag.String("http_endpoint", ":8098", "endpoint of http service")
	// InitialWindowSize window size
	InitialWindowSize int32 = 128 << 10
	// InitialConnWindowSize connection window size
	InitialConnWindowSize int32 = 64 << 10
	// ReadBufferSize buffer size
	ReadBufferSize = 32 << 10
	// WriteBufferSize write buffer size
	WriteBufferSize = 32 << 10
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithInitialWindowSize(InitialWindowSize), grpc.WithWriteBufferSize(WriteBufferSize), grpc.WithInitialConnWindowSize(InitialConnWindowSize), grpc.WithReadBufferSize(ReadBufferSize)}
	err := pb.RegisterXchainHandlerFromEndpoint(ctx, mux, *rpcEndpoint, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(*httpEndpoint, mux)
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}
