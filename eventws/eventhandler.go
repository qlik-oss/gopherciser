package eventws

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/atomichandlers"
)

type (
	// EventHandler handles events received on event websocket
	EventHandler struct {
		onOperation     map[string]map[uint64]*EventFunc
		onOperationLock sync.Mutex
		buffer          []Event
		bufferLock      sync.RWMutex
	}

	EventFunc struct {
		ID uint64

		f         func(Event)
		operation string
	}
)

var funcId = &atomichandlers.AtomicCounter{}

const BufferSize = 20

// NewEventHandler creates a new EventHandler which can be registered to and EventWs,
func NewEventHandler() *EventHandler {
	return &EventHandler{
		onOperation: make(map[string]map[uint64]*EventFunc),
		buffer:      make([]Event, 0, BufferSize),
	}
}

func (handler *EventHandler) event(actionState *action.State, message []byte) {
	var event Event
	if err := jsonit.Unmarshal(message, &event); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	// update event buffer
	handler.addToBuffer(event)

	handler.triggerFunctions(event)
}

func (handler *EventHandler) addToBuffer(event Event) {
	handler.bufferLock.Lock()
	defer handler.bufferLock.Unlock()
	if len(handler.buffer) > BufferSize-1 {
		handler.buffer = append(handler.buffer[1:], event)
	} else {
		handler.buffer = append(handler.buffer, event)
	}
}

func (handler *EventHandler) triggerFunctions(event Event) {
	handler.onOperationLock.Lock()
	defer handler.onOperationLock.Unlock()

	funcs := handler.onOperation[event.Operation]
	for _, eventFunc := range funcs {
		go eventFunc.f(event)
	}
}

// RegisterFunc to be executed on event triggering
// replay: triggers the latest events from buffer upon creation
func (handler *EventHandler) RegisterFunc(operation string, f func(event Event), replay bool) *EventFunc {
	eventFunc := &EventFunc{
		ID:        funcId.Inc(),
		f:         f,
		operation: operation,
	}

	handler.addEventFunc(eventFunc)

	if replay {
		// replay buffered events
		handler.replayEvents(operation, f)
	}

	return eventFunc
}

func (handler *EventHandler) replayEvents(operation string, f func(event Event)) {
	handler.bufferLock.RLock()
	defer handler.bufferLock.RUnlock()

	for _, event := range handler.buffer {
		if operation == event.Operation {
			go f(event)
		}
	}
}

// DeRegisterFunc from execution on event triggering
func (handler *EventHandler) DeRegisterFunc(eventFunc *EventFunc) {
	if eventFunc == nil || eventFunc.operation == "" {
		return
	}

	handler.onOperationLock.Lock()
	defer handler.onOperationLock.Unlock()

	eventFuncs, ok := handler.onOperation[eventFunc.operation]
	if ok {
		delete(eventFuncs, eventFunc.ID)
	}
}

func (handler *EventHandler) addEventFunc(eventFunc *EventFunc) {
	if eventFunc == nil || eventFunc.operation == "" {
		return
	}

	handler.onOperationLock.Lock()
	defer handler.onOperationLock.Unlock()

	if handler.onOperation == nil {
		handler.onOperation = make(map[string]map[uint64]*EventFunc)
	}

	eventFuncs := handler.onOperation[eventFunc.operation]
	if eventFuncs == nil {
		eventFuncs = make(map[uint64]*EventFunc)
		handler.onOperation[eventFunc.operation] = eventFuncs
	}

	eventFuncs[eventFunc.ID] = eventFunc
}

// FakeEvent fake event being received
func (handler *EventHandler) FakeEvent(event Event) {
	handler.addToBuffer(event)
	handler.triggerFunctions(event)
}
