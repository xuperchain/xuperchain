package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/xuperchain/xuperchain/core/pb"
	"google.golang.org/grpc"
	"net/http"
)

var (
	rpcEndpoint = flag.String("gateway_endpoint", "localhost:37101", "endpoint of grpc service forward to")
	// http port
	httpEndpoint = flag.String("http_endpoint", ":8098", "endpoint of http service")
	// enable default xendorser
	enableEndorser = flag.Bool("enable_endorser", false, "is enable xendorser")
	// enable CROS
	allowCROS = flag.Bool("allow_cros", false, "is allow Cross-origin resource sharing requests")

	// InitialWindowSize window size
	InitialWindowSize int32 = 128 << 10
	// InitialConnWindowSize connection window size
	InitialConnWindowSize int32 = 64 << 10
	// ReadBufferSize buffer size
	ReadBufferSize = 32 << 10
	// WriteBufferSize write buffer size
	WriteBufferSize = 32 << 10
)

func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	return
}

// interupt
func interupt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow CROS requests
		// Note: CROS is kind of dangerous in production environment
		//       don't use this without consideration
		if *allowCROS {
			if origin := r.Header.Get("Origin"); origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
					preflightHandler(w, r)
					return
				}
			}
		}

		h.ServeHTTP(w, r)

		// Request log
		log.Printf("ip=%s method=%s URL=%s\n", r.RemoteAddr, r.Method, r.URL.Path)
	})
}

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
	if *enableEndorser {
		err = pb.RegisterXendorserHandlerFromEndpoint(ctx, mux, *rpcEndpoint, opts)
		if err != nil {
			return err
		}
	}

	return http.ListenAndServe(*httpEndpoint, interupt(mux))
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		panic(err)
	}
}
