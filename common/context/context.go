package context

import (
	"context"
	"fmt"
	"time"

	"github.com/xuperchain/xuperchain/common/def"
	"github.com/xuperchain/xupercore/kernel/engines/xuperos/common"
	"github.com/xuperchain/xupercore/lib/logs"
	"github.com/xuperchain/xupercore/lib/timer"
)

const (
	ReqCtxKeyName = "reqCtx"
)

// 请求级别上下文
type ReqCtx interface {
	context.Context
	GetEngine() common.Engine
	GetLog() logs.Logger
	GetTimer() *timer.XTimer
	GetClientIp() string
}

type ReqCtxImpl struct {
	engine   common.Engine
	log      logs.Logger
	timer    *timer.XTimer
	clientIp string
}

func NewReqCtx(engine common.Engine, reqId, clientIp string) (ReqCtx, error) {
	if engine == nil {
		return nil, fmt.Errorf("new request context failed because engine is nil")
	}

	log, err := logs.NewLogger(reqId, def.SubModName)
	if err != nil {
		return nil, fmt.Errorf("new request context failed because new logger failed.err:%s", err)
	}

	ctx := &ReqCtxImpl{
		engine:   engine,
		log:      log,
		timer:    timer.NewXTimer(),
		clientIp: clientIp,
	}

	return ctx, nil
}

func WithReqCtx(ctx context.Context, reqCtx ReqCtx) context.Context {
	return context.WithValue(ctx, ReqCtxKeyName, reqCtx)
}

func ValueReqCtx(ctx context.Context) ReqCtx {
	val := ctx.Value(ReqCtxKeyName)
	if reqCtx, ok := val.(ReqCtx); ok {
		return reqCtx
	}
	return nil
}

func (t *ReqCtxImpl) GetEngine() common.Engine {
	return t.engine
}

func (t *ReqCtxImpl) GetLog() logs.Logger {
	return t.log
}

func (t *ReqCtxImpl) GetTimer() *timer.XTimer {
	return t.timer
}

func (t *ReqCtxImpl) GetClientIp() string {
	return t.clientIp
}

func (t *ReqCtxImpl) Deadline() (deadline time.Time, ok bool) {
	return
}

func (t *ReqCtxImpl) Done() <-chan struct{} {
	return nil
}

func (t *ReqCtxImpl) Err() error {
	return nil
}

func (t *ReqCtxImpl) Value(key interface{}) interface{} {
	return nil
}
