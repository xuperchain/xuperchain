package loggers

import (
	"fmt"
	"io"
	"text/template"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/term"
	"github.com/hyperledger/burrow/logging/structure"
)

type Syncable interface {
	Sync() error
}

func NewStreamLogger(writer io.Writer, format string) (log.Logger, error) {
	switch format {
	case "":
		return NewStreamLogger(writer, DefaultFormat)
	case JSONFormat:
		return NewJSONLogger(writer), nil
	case LogfmtFormat:
		return NewLogfmtLogger(writer), nil
	case TerminalFormat:
		return NewTerminalLogger(writer), nil
	default:
		return NewTemplateLogger(writer, format, []byte{})
	}
}

func NewJSONLogger(writer io.Writer) log.Logger {
	return interceptSync(writer, log.NewJSONLogger(writer))
}

func NewLogfmtLogger(writer io.Writer) log.Logger {
	return interceptSync(writer, log.NewLogfmtLogger(writer))
}

func NewTerminalLogger(writer io.Writer) log.Logger {
	logger := term.NewLogger(writer, log.NewLogfmtLogger, func(keyvals ...interface{}) term.FgBgColor {
		switch structure.Value(keyvals, structure.ChannelKey) {
		case structure.TraceChannelName:
			return term.FgBgColor{Fg: term.DarkGreen}
		default:
			return term.FgBgColor{Fg: term.Yellow}
		}
	})
	return interceptSync(writer, NewBurrowFormatLogger(logger, StringifyValues))
}

func NewTemplateLogger(writer io.Writer, textTemplate string, recordSeparator []byte) (log.Logger, error) {
	tmpl, err := template.New("template-logger").Parse(textTemplate)
	if err != nil {
		return nil, fmt.Errorf("could not parse '%s' as a text template: %v", textTemplate, err)
	}
	logger := log.LoggerFunc(func(keyvals ...interface{}) error {
		err := tmpl.Execute(writer, structure.KeyValuesMap(keyvals))
		if err == nil {
			_, err = writer.Write(recordSeparator)
		}
		return err
	})
	return interceptSync(writer, logger), nil
}

func interceptSync(writer io.Writer, logger log.Logger) log.Logger {
	return log.LoggerFunc(func(keyvals ...interface{}) error {
		switch structure.Signal(keyvals) {
		case structure.SyncSignal:
			if s, ok := writer.(Syncable); ok {
				return s.Sync()
			}
			// Don't log signals
			return nil
		default:
			return logger.Log(keyvals...)
		}
	})
}
