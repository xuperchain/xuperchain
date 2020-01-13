// Package events is the event center for internal system events.
// All modules could register event callback for each event.
package events

import (
	"errors"
	"reflect"
	"sync"
)

// Event center errors
var (
	ErrInvalidParams    = errors.New("Invalid Params")
	ErrQueueFull        = errors.New("Event queue is full")
	ErrDuplicateHandler = errors.New("Duplicate handler")
	ErrPartialSuccess   = errors.New("Operations partial success")
)

// EventBus is the event hub for all registered events
type EventBus struct {
	handler map[EventType][]EventHandler
	lock    sync.RWMutex
}

var eventBus *EventBus

func newEventBus() *EventBus {
	return &EventBus{
		make(map[EventType][]EventHandler),
		sync.RWMutex{},
	}
}

func init() {
	eventBus = newEventBus()
}

// GetEventBus return the instance of EventBus
func GetEventBus() *EventBus {
	return eventBus
}

func (ec *EventBus) findHandlerIndex(eType EventType, handler EventHandler) int {
	if ehs, ok := ec.handler[eType]; ok {
		shValue := reflect.ValueOf(handler)
		for idx, eh := range ehs {
			ehValue := reflect.ValueOf(eh)
			if shValue.Pointer() == ehValue.Pointer() {
				return idx
			}
		}
	}
	return -1
}

// Subscribe specified event type with a handler
func (ec *EventBus) Subscribe(eType EventType, handler EventHandler) error {
	if handler == nil {
		return ErrInvalidParams
	}
	ec.lock.Lock()
	defer ec.lock.Unlock()
	if ec.findHandlerIndex(eType, handler) >= 0 {
		// handler already exists
		return ErrDuplicateHandler
	}

	if _, ok := ec.handler[eType]; !ok {
		ec.handler[eType] = make([]EventHandler, 0)
	}
	ec.handler[eType] = append(ec.handler[eType], handler)
	return nil
}

// SubscribeMulti subscribe multiple events with one handler
func (ec *EventBus) SubscribeMulti(eTypes []EventType, handler EventHandler) ([]EventType, error) {
	if len(eTypes) == 0 || handler == nil {
		return nil, ErrInvalidParams
	}
	failedTypes := []EventType{}
	for _, eType := range eTypes {
		err := ec.Subscribe(eType, handler)
		if err != nil {
			failedTypes = append(failedTypes, eType)
		}
	}
	if len(failedTypes) != 0 {
		return failedTypes, ErrPartialSuccess
	}
	return nil, nil
}

// Unsubscribe the given event handler of specified event type
func (ec *EventBus) Unsubscribe(eType EventType, handler EventHandler) error {
	if handler == nil {
		return ErrInvalidParams
	}
	ec.lock.Lock()
	defer ec.lock.Unlock()
	idx := ec.findHandlerIndex(eType, handler)
	if idx < 0 {
		// handler not found, treat as unsubscribe successfully
		return nil
	}

	ec.handler[eType] = append(ec.handler[eType][:idx], ec.handler[eType][idx+1:]...)
	return nil
}

// UnsubscribeMulti unsubscribe multiple events with given handler
func (ec *EventBus) UnsubscribeMulti(eTypes []EventType, handler EventHandler) ([]EventType, error) {
	if len(eTypes) == 0 || handler == nil {
		return nil, ErrInvalidParams
	}
	failedTypes := []EventType{}
	for _, eType := range eTypes {
		err := ec.Unsubscribe(eType, handler)
		if err != nil {
			failedTypes = append(failedTypes, eType)
		}
	}
	if len(failedTypes) != 0 {
		return failedTypes, ErrPartialSuccess
	}
	return nil, nil
}

// FireEvent trigger a new event message to registerred handlers
func (ec *EventBus) FireEvent(em *EventMessage) error {
	if em == nil || em.Message == nil {
		return ErrInvalidParams
	}
	ec.dispatchEvent(em, nil)
	return nil
}

// FireEventAsync trigger a new event message to registerred handlers asynchronous
func (ec *EventBus) FireEventAsync(em *EventMessage) (*sync.WaitGroup, error) {
	if em == nil || em.Message == nil {
		return nil, ErrInvalidParams
	}
	wg := &sync.WaitGroup{}
	ec.dispatchEvent(em, wg)
	return wg, nil
}

func (ec *EventBus) dispatchEvent(em *EventMessage, wg *sync.WaitGroup) {
	// read lock help keep subscribers unchanged while process event
	ec.lock.RLock()
	defer ec.lock.RUnlock()
	if eh, ok := ec.handler[em.Type]; ok {
		for _, h := range eh {
			if wg != nil {
				wg.Add(1)
				go ec.handlerEvent(h, em, wg)
			} else {
				ec.handlerEvent(h, em, nil)
			}
		}
	}
}

func (ec *EventBus) handlerEvent(handler EventHandler, em *EventMessage, wg *sync.WaitGroup) {
	if em != nil {
		handler(em)
	}
	if wg != nil {
		wg.Done()
	}
}
