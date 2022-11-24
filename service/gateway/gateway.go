package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	scom "github.com/xuperchain/xuperchain/service/common"
	sconf "github.com/xuperchain/xuperchain/service/config"
	"github.com/xuperchain/xuperchain/service/pb"
	"github.com/xuperchain/xupercore/lib/logs"
)

type Gateway struct {
	scfg     *sconf.ServConf
	log      logs.Logger
	server   *http.Server
	isInit   bool
	exitOnce *sync.Once
}

func NewGateway(scfg *sconf.ServConf) (*Gateway, error) {
	if scfg == nil {
		return nil, fmt.Errorf("param error")
	}

	log, _ := logs.NewLogger("", scom.SubModName)
	obj := &Gateway{
		scfg:     scfg,
		log:      log,
		isInit:   true,
		exitOnce: &sync.Once{},
	}

	return obj, nil
}

// 启动gateway服务
func (t *Gateway) Run() error {
	if !t.isInit {
		return errors.New("gateway not init")
	}

	// 启动gateway，阻塞直到退出
	err := t.runGateway()
	if err != nil {
		t.log.Error("gateway abnormal exit.err:%v", err)
		return err
	}

	t.log.Trace("gateway exit")
	return nil
}

// 退出gateway服务，释放相关资源，需要幂等
func (t *Gateway) Exit() {
	if !t.isInit {
		return
	}

	t.exitOnce.Do(func() {
		t.stopGateway()
	})
}

func (t *Gateway) runGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithMaxMsgSize(t.scfg.MaxMsgSize),
		grpc.WithInitialWindowSize(t.scfg.InitWindowSize),
		grpc.WithWriteBufferSize(t.scfg.WriteBufSize),
		grpc.WithInitialConnWindowSize(t.scfg.InitConnWindowSize),
		grpc.WithReadBufferSize(t.scfg.ReadBufSize),
	}

	rpcEndpoint := fmt.Sprintf("127.0.0.1:%d", t.scfg.RpcPort)
	err := pb.RegisterXchainHandlerFromEndpoint(ctx, mux, rpcEndpoint, opts)
	if err != nil {
		return err
	}

	if t.scfg.EnableEndorser {
		err = pb.RegisterXendorserHandlerFromEndpoint(ctx, mux, rpcEndpoint, opts)
		if err != nil {
			return err
		}
	}

	addr := fmt.Sprintf(":%d", t.scfg.GWPort)
	t.server = &http.Server{
		Addr:    addr,
		Handler: t.interupt(mux),
	}
	err = t.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (t *Gateway) stopGateway() {
	if t.server != nil {
		t.server.Shutdown(context.Background())
	}
}

// interupt
func (t *Gateway) interupt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow CROS requests
		// Note: CROS is kind of dangerous in production environment
		// don't use this without consideration
		if t.scfg.AdapterAllowCROS {
			if origin := r.Header.Get("Origin"); origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
					t.preflightHandler(w, r)
					return
				}
			}
		}

		h.ServeHTTP(w, r)

		// Request log
		t.log.Trace("gateway access request", "ip", r.RemoteAddr, "method", r.Method, "url", r.URL.Path)
	})
}

func (t *Gateway) preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	return
}
