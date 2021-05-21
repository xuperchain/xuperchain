package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"reflect"
	"runtime"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	sctx "github.com/xuperchain/xupercore/example/xchain/common/context"
	ecom "github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/lib/logs"
	"github.com/xuperchain/xupercore/lib/utils"

	"github.com/xuperchain/xuperchain/common/xupospb/pb"
	acom "github.com/xuperchain/xuperchain/service/common"
)

type RpcServ struct {
	engine ecom.Engine
	log    logs.Logger
}

func NewRpcServ(engine ecom.Engine, log logs.Logger) *RpcServ {
	return &RpcServ{
		engine: engine,
		log:    log,
	}
}

// UnaryInterceptor provides a hook to intercept the execution of a unary RPC on the server.
func (t *RpcServ) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (respRes interface{}, err error) {
		// set request header
		type HeaderInterface interface {
			GetHeader() *pb.Header
		}
		if req.(HeaderInterface).GetHeader() == nil {
			header := reflect.ValueOf(req).Elem().FieldByName("Header")
			if header.IsValid() && header.IsNil() && header.CanSet() {
				header.Set(reflect.ValueOf(t.defReqHeader()))
			}
		}
		if req.(HeaderInterface).GetHeader().GetLogid() == "" {
			req.(HeaderInterface).GetHeader().Logid = utils.GenLogId()
		}
		reqHeader := req.(HeaderInterface).GetHeader()

		// set request context
		reqCtx, _ := t.createReqCtx(ctx, reqHeader)
		ctx = sctx.WithReqCtx(ctx, reqCtx)

		// output access log
		logFields := make([]interface{}, 0)
		logFields = append(logFields, "from", reqHeader.GetFromNode(),
			"client_ip", reqCtx.GetClientIp(), "rpc_method", info.FullMethod)

		// panic recover
		defer func() {
			reqCtx.GetLog().Info("access", logFields...)
			if e := recover(); e != nil {
				err = fmt.Errorf("%s log_id = %s", ecom.ErrInternal, reqCtx.GetLog().GetLogId())
				reqCtx.GetLog().Error("Rpc server happen panic", "error", e)

				// stack
				stack := make([]byte, 8192)
				n := runtime.Stack(stack[:], false)
				log.Printf("%s Rpc server happen panic: %s", reqCtx.GetLog().GetLogId(), stack[:n])
			}
		}()

		// handle request
		// 根据err自动设置响应错误码，err需要是ecom.Error类型的标准err，否则会响应为未知错误
		stdErr := ecom.ErrSuccess
		respRes, err = handler(ctx, req)
		if err != nil {
			stdErr = ecom.CastError(err)
		}
		// 根据错误统一设置header，对外统一响应err=nil，通过Header.ErrCode判断
		respHeader := &pb.Header{
			Logid:    reqHeader.GetLogid(),
			FromNode: t.genTraceId(),
			Error:    t.convertErr(stdErr),
		}
		// 通过反射设置header到response
		header := reflect.ValueOf(respRes).Elem().FieldByName("Header")
		if header.IsValid() && header.IsNil() && header.CanSet() {
			header.Set(reflect.ValueOf(respHeader))
		}

		// output ending log
		// 可以通过log库提供的SetInfoField方法附加输出到ending log
		logFields = append(logFields, "status", stdErr.Status, "err_code", stdErr.Code,
			"err_msg", stdErr.Msg, "cost_time", reqCtx.GetTimer().Print())
		return respRes, err
	}
}

func (t *RpcServ) defReqHeader() *pb.Header {
	return &pb.Header{
		Logid:    utils.GenLogId(),
		FromNode: "",
		Error:    pb.XChainErrorEnum_UNKNOW_ERROR,
	}
}

func (t *RpcServ) createReqCtx(gctx context.Context, reqHeader *pb.Header) (sctx.ReqCtx, error) {
	// 获取客户端ip
	clientIp, err := t.getClietIP(gctx)
	if err != nil {
		t.log.Error("access proc failed because get client ip failed", "error", err)
		return nil, fmt.Errorf("get client ip failed")
	}

	// 创建请求上下文
	rctx, err := sctx.NewReqCtx(t.engine, reqHeader.GetLogid(), clientIp)
	if err != nil {
		t.log.Error("access proc failed because create request context failed", "error", err)
		return nil, fmt.Errorf("create request context failed")
	}

	return rctx, nil
}

func (t *RpcServ) getClietIP(gctx context.Context) (string, error) {
	pr, ok := peer.FromContext(gctx)
	if !ok {
		return "", fmt.Errorf("create peer form context failed")
	}

	if pr.Addr == nil || pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("get client_ip failed because peer.Addr is nil")
	}

	addrSlice := strings.Split(pr.Addr.String(), ":")
	return addrSlice[0], nil
}

// 生成包含机器host和请求时间的AES加密字符串，方便问题定位
func (t *RpcServ) genTraceId() string {
	return utils.GetHostName()
}

// 转化错误类型为原接口错误
func (t *RpcServ) convertErr(stdErr *ecom.Error) pb.XChainErrorEnum {
	if stdErr == nil {
		return pb.XChainErrorEnum_UNKNOW_ERROR
	}

	if errCode, ok := acom.StdErrToXchainErrMap[stdErr.Code]; ok {
		return errCode
	}

	return pb.XChainErrorEnum_UNKNOW_ERROR
}
