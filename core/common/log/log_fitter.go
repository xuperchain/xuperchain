package log

import (
	"fmt"
	"os"
	"sync"
)

// Reserve common key
const (
	CommFieldLogId    = "r_logid"
	CommFieldPid      = "r_pid"
	CommFieldCall     = "r_call"
	CommFieldIsNotice = "r_ntce"
)

const (
	DefaultCallDepth = 4
)

// Logger wrapper interface
type LogInterface interface {
	Crit(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
	Warn(msg string, ctx ...interface{})
	Info(msg string, ctx ...interface{})
	Trace(msg string, ctx ...interface{})
	Debug(msg string, ctx ...interface{})
}

// Logger Fitter
type LogFitter struct {
	logger       LogInterface
	logId        string
	pid          int
	commFields   []interface{}
	commFieldLck *sync.RWMutex
	ntceFields   []interface{}
	ntceFieldLck *sync.RWMutex
	callDepth    int
}

func NewLogger(logger LogInterface, logId string) (*LogFitter, error) {
	if logger == nil {
		return nil, fmt.Errorf("new logger param error")
	}
	if logId == "" {
		logId = GenLogId()
	}

	lf := &LogFitter{
		logger:       logger,
		logId:        logId,
		pid:          os.Getpid(),
		commFields:   make([]interface{}, 0),
		commFieldLck: &sync.RWMutex{},
		ntceFields:   make([]interface{}, 0),
		ntceFieldLck: &sync.RWMutex{},
		callDepth:    DefaultCallDepth,
	}

	return lf, nil
}

func (t *LogFitter) GetLogId() string {
	return t.logId
}

func (t *LogFitter) SetCommField(key string, value interface{}) {
	if !t.isInit() || key == "" || value == nil {
		return
	}

	t.commFieldLck.Lock()
	defer t.commFieldLck.Unlock()

	t.commFields = append(t.commFields, key, value)
}

func (t *LogFitter) SetNoticeField(key string, value interface{}) {
	if !t.isInit() || key == "" || value == nil {
		return
	}

	t.ntceFieldLck.Lock()
	defer t.ntceFieldLck.Unlock()

	t.ntceFields = append(t.ntceFields, key, value)
}

func (t *LogFitter) Crit(msg string, ctx ...interface{}) {
	if !t.isInit() {
		return
	}
	t.logger.Crit(msg, t.fmtCommLogger(ctx...)...)
}

func (t *LogFitter) Error(msg string, ctx ...interface{}) {
	if !t.isInit() {
		return
	}
	t.logger.Error(msg, t.fmtCommLogger(ctx...)...)
}

func (t *LogFitter) Warn(msg string, ctx ...interface{}) {
	if !t.isInit() {
		return
	}
	t.logger.Warn(msg, t.fmtCommLogger(ctx...)...)
}

func (t *LogFitter) Notice(msg string, ctx ...interface{}) {
	if !t.isInit() {
		return
	}
	t.logger.Info(msg, t.fmtNoticeLogger(ctx...)...)
}

func (t *LogFitter) Info(msg string, ctx ...interface{}) {
	if !t.isInit() {
		return
	}
	t.logger.Info(msg, t.fmtCommLogger(ctx...)...)
}

func (t *LogFitter) Trace(msg string, ctx ...interface{}) {
	if !t.isInit() {
		return
	}
	t.logger.Trace(msg, t.fmtCommLogger(ctx...)...)
}

func (t *LogFitter) Debug(msg string, ctx ...interface{}) {
	if !t.isInit() {
		return
	}

	t.logger.Debug(msg, t.fmtCommLogger(ctx...)...)
}

func (t *LogFitter) getCommField() []interface{} {
	t.commFieldLck.RLock()
	defer t.commFieldLck.RUnlock()

	return t.commFields
}

func (t *LogFitter) genBaseField(isNotice bool) []interface{} {
	comCtx := make([]interface{}, 0)
	fileLine, _ := GetFuncCall(t.callDepth)
	comCtx = append(comCtx, CommFieldCall, fileLine)
	comCtx = append(comCtx, CommFieldPid, t.pid)
	comCtx = append(comCtx, CommFieldLogId, t.logId)
	comCtx = append(comCtx, CommFieldIsNotice, isNotice)

	return comCtx
}

func (t *LogFitter) fmtCommLogger(ctx ...interface{}) []interface{} {
	if len(ctx)%2 != 0 {
		last := ctx[len(ctx)-1]
		ctx = ctx[:len(ctx)-1]
		ctx = append(ctx, "unknow", last)
	}

	// Ensure consistent output sequence
	comCtx := t.genBaseField(false)
	comCtx = append(comCtx, t.getCommField()...)
	comCtx = append(comCtx, ctx...)

	return comCtx
}

func (t *LogFitter) getNoticeField() []interface{} {
	t.ntceFieldLck.RLock()
	defer t.ntceFieldLck.RUnlock()

	return t.ntceFields
}

func (t *LogFitter) fmtNoticeLogger(ctx ...interface{}) []interface{} {
	if len(ctx)%2 != 0 {
		last := ctx[len(ctx)-1]
		ctx = ctx[:len(ctx)-1]
		ctx = append(ctx, "unknow", last)
	}

	comCtx := t.genBaseField(true)
	comCtx = append(comCtx, t.getCommField()...)
	comCtx = append(comCtx, t.getNoticeField()...)
	comCtx = append(comCtx, ctx...)

	t.clearNoticeFields()
	return comCtx
}

func (t *LogFitter) clearNoticeFields() {
	t.ntceFieldLck.RLock()
	defer t.ntceFieldLck.RUnlock()

	t.ntceFields = t.ntceFields[:0]
}

func (t *LogFitter) isInit() bool {
	if t.logger == nil || t.commFields == nil || t.ntceFields == nil ||
		t.commFieldLck == nil || t.ntceFieldLck == nil {
		return false
	}

	return true
}
