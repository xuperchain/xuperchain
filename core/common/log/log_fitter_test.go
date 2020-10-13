package log

import (
	"sync"
	"testing"

	"github.com/xuperchain/xuperchain/core/common/config"
)

var logCfg = &config.LogConfig{
	Module:         "test",
	Filepath:       "logs",
	Filename:       "test",
	Fmt:            "logfmt",
	Console:        true,
	Level:          "debug",
	Async:          true,
	RotateInterval: 3600,
	RotateBackups:  3,
}

func TestNotice(t *testing.T) {
	logger, err := OpenDefaultLog(logCfg)
	if err != nil {
		t.Errorf("open log fail.err:%v", err)
	}

	log, err := NewLogger(logger, GenLogId())
	if err != nil {
		t.Errorf("new logger fail.err:%v", err)
	}

	wg := &sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()

			log.Notice("test test test", "a", 1, "b", 2, "c", 3, "num", num)
			log.Debug("test", "a", 1, "b", 2, "c", 3, "num", num)
			log.Trace("test", "a", 1, "b", 2, "c", 3, "num", num)
			log.Info("test", "a", 1, "b", 2, "c", 3, "num", num)
			log.SetNoticeField("key1", num)
			log.SetNoticeField("key2", num)
			log.Notice("test test", "a", true, "b", 1, "num", num)
			log.SetNoticeField("key10", num)
		}(i)
	}

	log.Notice("test", "a", 1, "b", 2, "c", 3)
	log.Notice("test", "a", 1, "b", 2, "c", 3)
	log.SetNoticeField("key3", 3)
	log.SetNoticeField("key4", 4)
	log.SetNoticeField("key5", 5)
	log.SetNoticeField("key6", 6)
	log.Notice("test", "a", 1, "b", 2, "c", 3)
	log.Notice("test", "a", 1, "b", 2, "c", 3)
	log.Notice("test", "a", 1, "b", 2, "c", 3)
	log.Warn("test warn", 1)
	log.Warn("test warn", 1, 2)

	wg.Wait()
	log.Notice("test", "a", 1, "b", 2, "c", 3)
}

func TestNewLoggerForEVM(t *testing.T) {
	logger, _ := NewLoggerForEVM()

	logger.TraceMsg("EVM logger test", "result", 1)
}
