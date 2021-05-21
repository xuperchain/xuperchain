package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"sync"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	gpromeus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/xuperchain/xupercore/kernel/engines"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/lib/logs"

	sconf "github.com/xuperchain/xuperchain/common/config"
	"github.com/xuperchain/xuperchain/common/def"
	"github.com/xuperchain/xuperchain/common/xupospb/pb"
)

// rpc server启停控制管理
type RpcServMG struct {
	scfg     *sconf.ServConf
	engine   ecom.Engine
	log      logs.Logger
	rpcServ  *RpcServ
	servHD   *grpc.Server
	isInit   bool
	exitOnce *sync.Once
}

func NewRpcServMG(scfg *sconf.ServConf, engine engines.BCEngine) (*RpcServMG, error) {
	if scfg == nil || engine == nil {
		return nil, fmt.Errorf("param error")
	}
	xosEngine, err := xuperos.EngineConvert(engine)
	if err != nil {
		return nil, fmt.Errorf("not xuperos engine")
	}

	log, _ := logs.NewLogger("", def.SubModName)
	obj := &RpcServMG{
		scfg:     scfg,
		engine:   xosEngine,
		log:      log,
		rpcServ:  NewRpcServ(engine.(ecom.Engine), log),
		isInit:   true,
		exitOnce: &sync.Once{},
	}

	return obj, nil
}

// 启动rpc服务，阻塞运行
func (t *RpcServMG) Run() error {
	if !t.isInit {
		return errors.New("RpcServMG not init")
	}

	t.log.Trace("run grpc server", "isTls", t.scfg.EnableTls)

	// 启动rpc server，阻塞直到退出
	err := t.runRpcServ()
	if err != nil {
		t.log.Error("grpc server abnormal exit", "err", err)
		return err
	}

	t.log.Trace("grpc server exit")
	return nil
}

// 退出rpc服务，释放相关资源，需要幂等
func (t *RpcServMG) Exit() {
	if !t.isInit {
		return
	}

	t.exitOnce.Do(func() {
		t.stopRpcServ()
	})
}

// 启动rpc服务，阻塞直到退出
func (t *RpcServMG) runRpcServ() error {
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		t.rpcServ.UnaryInterceptor(),
		gpromeus.UnaryServerInterceptor,
	}

	rpcOptions := []grpc.ServerOption{
		middleware.WithUnaryServerChain(unaryInterceptors...),
		grpc.MaxRecvMsgSize(t.scfg.MaxMsgSize),
		grpc.ReadBufferSize(t.scfg.ReadBufSize),
		grpc.InitialWindowSize(t.scfg.InitWindowSize),
		grpc.InitialConnWindowSize(t.scfg.InitConnWindowSize),
		grpc.WriteBufferSize(t.scfg.WriteBufSize),
	}

	if t.scfg.EnableTls {
		creds, err := t.newTls()
		if err != nil {
			return err
		}
		rpcOptions = append(rpcOptions, grpc.Creds(creds))
	}

	t.servHD = grpc.NewServer(rpcOptions...)
	pb.RegisterXchainServer(t.servHD, t.rpcServ)

	// event involved rpc
	eventService := newEventService(t.scfg, t.engine)
	pb.RegisterEventServiceServer(t.servHD, eventService)

	if t.scfg.EnableEndorser {
		endorserService, err := newEndorserService(t.scfg, t.engine, t.rpcServ)
		if err != nil {
			t.log.Error("failed to register endorser", "err", err)
			return fmt.Errorf("failed to register endorser")
		}
		pb.RegisterXendorserServer(t.servHD, endorserService)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", t.scfg.RpcPort))
	if err != nil {
		t.log.Error("failed to listen", "err", err)
		return fmt.Errorf("failed to listen")
	}

	reflection.Register(t.servHD)
	if err := t.servHD.Serve(lis); err != nil {
		t.log.Error("failed to serve", "err", err)
		return err
	}

	t.log.Trace("rpc server exit")
	return nil
}

func (t *RpcServMG) newTls() (credentials.TransportCredentials, error) {
	envConf := t.engine.Context().EnvCfg
	tlsPath := envConf.GenDataAbsPath(envConf.TlsDir)
	bs, err := ioutil.ReadFile(filepath.Join(tlsPath, "cert.crt"))
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		return nil, err
	}
	certificate, err := tls.LoadX509KeyPair(filepath.Join(tlsPath, "key.pem"),
		filepath.Join(tlsPath, "private.key"))
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName:   t.scfg.TlsServerName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	return creds, nil
}

// 需要幂等
func (t *RpcServMG) stopRpcServ() {
	if t.servHD != nil {
		// 优雅关闭grpc server
		t.servHD.GracefulStop()
	}
}
