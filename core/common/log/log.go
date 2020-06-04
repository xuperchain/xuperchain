package log

import (
	"fmt"
	"os"

	log "github.com/xuperchain/log15"
	"github.com/xuperchain/xuperchain/core/common/config"
)

var (
	// DefaultLogger is the default logger used by Info, Warn ... methods
	DefaultLogger = Logger{log.Root()}

	// Trace print trace level log
	Trace = DefaultLogger.Trace
	// Debug print debug level log
	Debug = DefaultLogger.Debug
	// Info print info level log
	Info = DefaultLogger.Info
	// Warn print warn level log
	Warn = DefaultLogger.Warn
	// Error print error level log
	Error = DefaultLogger.Error
)

// LogBufSize define log buffer channel size
const LogBufSize = 102400

// Logger wrapper
type Logger struct {
	log.Logger
}

// OpenLog create and open log stream using LogConfig
func OpenLog(lc *config.LogConfig) (Logger, error) {

	infoFile := lc.Filepath + "/" + lc.Filename + ".log"
	wfFile := lc.Filepath + "/" + lc.Filename + ".log.wf"
	os.MkdirAll(lc.Filepath, os.ModePerm)

	lfmt := log.LogfmtFormat()
	switch lc.Fmt {
	case "json":
		lfmt = log.JsonFormat()
	}

	xlog := log.New("module", lc.Module)

	lvLevel, err := log.LvlFromString(lc.Level)
	if nil != err {
		fmt.Printf("log level error%v\n", err)
	}

	// set lowest level as level limit, this may improve performance
	xlog.SetLevelLimit(lvLevel)

	// init normal and warn/fault log file handler, RotateFileHandler
	// only valid if `RotateInterval` and `RotateBackups` greater than 0
	var (
		nmHandler log.Handler
		wfHandler log.Handler
	)
	if lc.RotateInterval > 0 && lc.RotateBackups > 0 {
		nmHandler = log.Must.RotateFileHandler(
			infoFile, lfmt, lc.RotateInterval, lc.RotateBackups)
		wfHandler = log.Must.RotateFileHandler(
			wfFile, lfmt, lc.RotateInterval, lc.RotateBackups)
	} else {
		nmHandler = log.Must.FileHandler(infoFile, lfmt)
		wfHandler = log.Must.FileHandler(wfFile, lfmt)
	}

	if lc.Async {
		nmHandler = log.BufferedHandler(LogBufSize, nmHandler)
		wfHandler = log.BufferedHandler(LogBufSize, wfHandler)
	}

	// prints log level between `lvLevel` to Info to common log
	nmfileh := log.BoundLvlFilterHandler(lvLevel, log.LvlError, nmHandler)

	// prints log level greater or equal to Warn to wf log
	wffileh := log.LvlFilterHandler(log.LvlWarn, wfHandler)

	var lhd log.Handler
	if lc.Console {
		hstd := log.StreamHandler(os.Stderr, lfmt)
		lhd = log.SyncHandler(log.MultiHandler(hstd, nmfileh, wffileh))
	} else {
		lhd = log.SyncHandler(log.MultiHandler(nmfileh, wffileh))
	}

	xlog.SetHandler(lhd)
	l := Logger{xlog}
	return l, err
}

// OpenDefaultLog create and open log stream using LogConfig and assign DefaultLogger
func OpenDefaultLog(lc *config.LogConfig) (Logger, error) {
	logger, err := OpenLog(lc)
	if err != nil {
		return logger, err
	}
	DefaultLogger = logger
	Trace = DefaultLogger.Trace
	Debug = DefaultLogger.Debug
	Info = DefaultLogger.Info
	Warn = DefaultLogger.Warn
	Error = DefaultLogger.Error
	return logger, nil
}

func init() {
	DefaultLogger.SetHandler(log.StreamHandler(os.Stderr, log.LogfmtFormat()))
}
