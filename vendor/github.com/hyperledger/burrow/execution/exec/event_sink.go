package exec

import (
	"github.com/hyperledger/burrow/execution/errors"
)

type EventSink interface {
	Call(call *CallEvent, exception *errors.Exception) error
	Log(log *LogEvent) error
}

type noopEventSink struct {
}

func NewNoopEventSink() *noopEventSink {
	return &noopEventSink{}
}

func (es *noopEventSink) Call(call *CallEvent, exception *errors.Exception) error {
	return nil
}

func (es *noopEventSink) Log(log *LogEvent) error {
	return nil
}

type logFreeEventSink struct {
	EventSink
}

func NewLogFreeEventSink(eventSink EventSink) *logFreeEventSink {
	return &logFreeEventSink{
		EventSink: eventSink,
	}
}

func (esc *logFreeEventSink) Log(log *LogEvent) error {
	return errors.Errorf(errors.Codes.IllegalWrite,
		"Log emitted from contract %v, but current call should be log-free", log.Address)
}
