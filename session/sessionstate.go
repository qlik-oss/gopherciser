package session

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	enigma "github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/randomizer"
	"github.com/qlik-oss/gopherciser/requestmetrics"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

type (
	// Rand currently used randomizer for connection
	rand struct {
		mu  sync.Mutex
		rnd *randomizer.Randomizer
	}

	// Event encapsulates an event channel and a function to be executed on events
	Event struct {
		// F is function to be executed on event
		F func(ctx context.Context, actionState *action.State) error
		// NoFailOnError -
		// False (Default) - Report error and abort action
		// True - Log warning instead of failing and error reporting
		NoFailOnError bool
		// Close executed on de-registering event
		Close func()
	}

	// IConnection interface for current
	IConnection interface {
		// Disconnect connection
		Disconnect() error

		// SetSense : set current sense environment
		SetSense(*enigmahandlers.SenseUplink)
		// Sense : connection to sense environment
		Sense() *enigmahandlers.SenseUplink
	}

	// State for user
	State struct {
		ctx       context.Context
		ctxCancel context.CancelFunc

		Cookies      http.CookieJar
		VirtualProxy string
		Connection   IConnection
		ArtifactMap  *ArtifactMap
		IDMap        IDMap
		HeaderJar    *HeaderJar
		LoggedIn     bool
		Timeout      time.Duration
		User         *users.User
		OutputsDir   string
		CurrentApp   *ArtifactEntry
		CurrentUser  *elasticstructs.User

		rand          *rand
		trafficLogger enigmahandlers.ITrafficLogger

		// CurrentActionState will contain the state of the latest action to be started
		CurrentActionState *action.State
		LogEntry           *logger.LogEntry
		EW                 statistics.ErrWarn
		Pending            PendingHandler
		Rest               *RestHandler
		RequestMetrics     *requestmetrics.RequestMetrics

		events  map[int]*Event // todo support multiple events per handle?
		eventMu sync.Mutex
	}

	// SessionVariables is used as a data carrier for session variables.
	SessionVariables struct {
		users.User
		Session uint64
		Thread  uint64
		Local   interface{}
	}
)

const (
	// DefaultTimeout per request timeout
	DefaultTimeout = 300 * time.Second
)

// New instance of session state
func New(ctx context.Context, outputsDir string, timeout time.Duration,
	user *users.User, session, instance uint64, virtualProxy string, onlyInstanceSeed bool) *State {
	sessionCtx, cancel := context.WithCancel(ctx)

	state := &State{
		ctx:          sessionCtx,
		ctxCancel:    cancel,
		Timeout:      timeout,
		ArtifactMap:  NewAppMap(),
		OutputsDir:   outputsDir,
		User:         user,
		HeaderJar:    NewHeaderJar(),
		VirtualProxy: virtualProxy,
		// Buffer size for the pendingHandler has been chosen after evaluation tests towards sense
		// with medium amount of objects in the sheets. Evaluation was done before introducing spinLoopPending
		// in pendingHandler and could possibly be lowered, this would however require re-evaluation.
		Pending:        NewPendingHandler(32),
		RequestMetrics: &requestmetrics.RequestMetrics{},
		events:         make(map[int]*Event),
	}

	if state.Timeout < time.Millisecond {
		state.Timeout = DefaultTimeout
	}

	if onlyInstanceSeed {
		// Use same random sequence for all users
		session = 1
	}
	rnd := randomizer.NewSeededRandomizer(randomizer.GetPredictableSeedUInt64(instance, session))
	state.SetRandomizer(rnd, false)

	return state
}

// Reset session, to be used when an existing session state enters a new "sequence"
func (state *State) Reset(ctx context.Context) {
	if state.ctxCancel != nil {
		state.ctxCancel()
	}
	state.ctx, state.ctxCancel = context.WithCancel(ctx)

	state.Connection = nil
	state.ArtifactMap = NewAppMap()
	state.IDMap = IDMap{}
	state.trafficLogger = nil
	state.HeaderJar = NewHeaderJar()
	state.LoggedIn = false
	state.CurrentActionState = nil
	state.EW = statistics.ErrWarn{}
	state.Rest = nil
	state.RequestMetrics = &requestmetrics.RequestMetrics{}
	state.events = make(map[int]*Event)
	state.CurrentApp = nil
	state.CurrentUser = nil
}

// SetLogEntry set the log entry
func (state *State) SetLogEntry(entry *logger.LogEntry) {
	state.LogEntry = entry

	if entry.ShouldLogTraffic() {
		state.trafficLogger = enigmahandlers.NewTrafficLogger(entry)
	} else {
		state.trafficLogger = enigmahandlers.NewTrafficRequestCounter()
	}

	state.Rest = NewRestHandler(state.ctx, 64, state.trafficLogger, state.HeaderJar, state.VirtualProxy, state.Timeout)
}

// TrafficLogger returns the current trafficLogger
func (state *State) TrafficLogger() enigmahandlers.ITrafficLogger {
	return state.trafficLogger
}

// Randomizer get randomizer for session
func (state *State) Randomizer() *randomizer.Randomizer {
	state.rand.mu.Lock()
	defer state.rand.mu.Unlock()

	return state.rand.rnd
}

// SetRandomizer set randomizer for session, will not be set if already has a randomizer
// on concurrent sets, first instance to acquire lock will "win". When setting to nil, it
// will be automatically forced. Set force flag to have randomizer set even when a randomizer exists.
func (state *State) SetRandomizer(rnd *randomizer.Randomizer, force bool) {
	if state.rand == nil {
		state.rand = &rand{
			rnd: rnd,
		}
		return
	}

	state.rand.mu.Lock()
	defer state.rand.mu.Unlock()

	if rnd == nil {
		state.rand.rnd = nil
		return
	}

	if state.rand.rnd != nil && !force {
		return
	}

	state.rand.rnd = rnd
}

// IsAbortTriggered check if abort has been flagged
func (state *State) IsAbortTriggered() bool {
	return helpers.IsContextTriggered(state.ctx)
}

// Wait for all pending requests to finish, returns true if action state has been marked as failed
func (state *State) Wait(actionState *action.State) bool {
	state.Pending.WaitForPending(state.ctx)
	state.Rest.WaitForPending()
	return actionState.Failed
}

// ContextWithTimeout new context based on ctx with default timeout
func (state *State) ContextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = state.ctx
	}
	return context.WithTimeout(ctx, state.Timeout)
}

// ContextChangeList create changeList object and context to be used for synchronizing changes
func (state *State) ContextChangeList() (context.Context, *enigma.ChangeLists) {
	cl := enigma.ChangeLists{}
	ctxWithCL := context.WithValue(state.ctx, enigma.ChangeListsKey{}, &cl)
	return ctxWithCL, &cl
}

// ReqContext context to be used on request, includes timeout and changeList to be used for synchronizing changes.
// A request context is automatically added when using QueueRequest functions.
func (state *State) ReqContext() (context.Context, *enigma.ChangeLists, context.CancelFunc) {
	ctx, cl := state.ContextChangeList()
	ctx, cancel := state.ContextWithTimeout(ctx)
	return ctx, cl, cancel
}

// BaseContext for state. Normally shouldn't be used. Use ReqContext for for context on Sense actions, or methods handling the
// context such as QueueRequest and SendRequest
func (state *State) BaseContext() context.Context {
	return state.ctx
}

// QueueRequest Async request, add error to action state or log as warning depending on failOnError flag.
// This method adds timeout and ChangeList to ctx context and auto triggers changes. Thus ctx should not be used
// when having multiple request in a QueueRequest function, instead use SendRequest and SendRequestRaw internally in f.
// Changes can also be handled "manually" with the help of TriggerContextChanges.
func (state *State) QueueRequest(f func(ctx context.Context) error, actionState *action.State, failOnError bool, errMsg string) {
	state.QueueRequestWithCallback(f, actionState, failOnError, errMsg, nil)
}

// QueueRequestWithCallback Async request, add error to action state or log as warning depending on failOnError flag.
// This method adds timeout and ChangeList to ctx context and auto triggers changes. Thus ctx should not be used
// when having multiple request in a QueueRequest function, instead use SendRequest and SendRequestRaw internally in f.
// Changes can also be handled "manually" with the help of TriggerContextChanges.
func (state *State) QueueRequestWithCallback(f func(ctx context.Context) error, actionState *action.State, failOnError bool, errMsg string, callback func(err error)) {
	ctx, cl := state.ContextChangeList()

	// When request has executed, report errors or trigger events
	onFinished := func(err error) {
		if err != nil {
			state.onRequestError(err, actionState, failOnError, errMsg)
		} else if len(cl.Changed) > 0 {
			state.TriggerEvents(actionState, cl.Changed, cl.Closed)
		}

		if callback != nil {
			callback(err)
		}
	}

	state.Pending.QueueRequest(ctx, state.Timeout, f, state.LogEntry, onFinished)
}

func (state *State) onRequestError(err error, actionState *action.State, failOnError bool, errMsg string) {
	if failOnError {
		if actionState == nil {
			state.LogEntry.LogErrorWithMsg("actionstate nil! error not reported correctly on action", err)
			return
		}
		actionState.AddErrors(err)
	} else {
		if helpers.IsContextTriggered(state.BaseContext()) {
			return // Don't log "error" warnings when cancelling
		}
		warning := err.Error()
		if errMsg != "" {
			warning = errMsg + ". " + warning
		}
		state.LogEntry.Log(logger.WarningLevel, warning)
	}
}

// Disconnect de-registers all events and disconnects current connection
func (state *State) Disconnect() {
	if state == nil {
		return
	}
	state.LogEntry.LogDebug("Disconnect session")

	state.DeregisterAllEvents()

	if state.Connection != nil {
		if err := state.Connection.Disconnect(); err != nil {
			if state.LogEntry != nil {
				state.LogEntry.LogErrorWithMsg("error disconnecting", err)
			}
		}
	}
}

// DeregisterAllEvents for session
func (state *State) DeregisterAllEvents() {
	state.eventMu.Lock()
	defer state.eventMu.Unlock()
	for _, event := range state.events {
		if event != nil && event.Close != nil {
			event.Close()
		}
	}
	state.events = make(map[int]*Event)
}

// RegisterEvent register function to be executed on object change
func (state *State) RegisterEvent(handle int,
	event func(ctx context.Context, actionState *action.State) error,
	onClose func(), failOnError bool) {
	state.registerEvent(handle, &Event{
		F:             event,
		NoFailOnError: !failOnError,
		Close:         onClose,
	})
}

// DeRegisterEvents for handles in list
func (state *State) DeRegisterEvents(handles []int) {
	state.eventMu.Lock()
	defer state.eventMu.Unlock()
	for _, handle := range handles {
		state.deRegisterEventNoLock(handle)
	}
}

// DeRegisterEvent for handle
func (state *State) DeRegisterEvent(handle int) {
	state.eventMu.Lock()
	defer state.eventMu.Unlock()
	state.deRegisterEventNoLock(handle)
}

func (state *State) deRegisterEventNoLock(handle int) {
	if event := state.events[handle]; event != nil {
		if event.Close != nil {
			event.Close()
		}
		delete(state.events, handle)
	}
}

func (state *State) registerEvent(handle int, event *Event) {
	if state == nil || event == nil {
		return
	}
	state.eventMu.Lock()
	defer state.eventMu.Unlock()
	// todo check if already existing event, handle how?
	state.events[handle] = event
}

// TriggerEvents from change and close lists
func (state *State) TriggerEvents(actionState *action.State, chHandles, clHandles []int) {
	// handle close events
	if len(clHandles) > 0 {
		state.LogEntry.LogDebugf("Trigger close for handles %v", clHandles)
		state.Pending.IncPending()
		go func() {
			defer state.Pending.DecPending()
			state.DeRegisterEvents(clHandles)
			if state.Connection != nil && state.Connection.Sense() != nil {
				if err := state.Connection.Sense().Objects.ClearObjects(clHandles); err != nil {
					actionState.AddErrors(err)
				}
			}
		}()
	}

	// handle change events
	if len(chHandles) < 1 {
		return
	}

	state.eventMu.Lock()
	defer state.eventMu.Unlock()

	state.LogEntry.LogDebugf("Trigger events for handles %v", chHandles)
	for _, handle := range chHandles {
		if event := state.events[handle]; event != nil {
			state.Pending.IncPending()
			go func() {
				defer state.Pending.DecPending()

				if event.F != nil {
					state.QueueRequest(func(ctx context.Context) error {
						return event.F(ctx, actionState)
					}, actionState, true, "")
				}
			}()
		}
	}
}

// SendRequest and trigger any object changes synchronously
func (state *State) SendRequest(actionState *action.State, f func(ctx context.Context) error) error {
	ctx, cl, cancel := state.ReqContext()
	defer cancel()

	if err := f(ctx); err != nil {
		return errors.WithStack(err)
	}

	state.TriggerEvents(actionState, cl.Changed, cl.Closed)

	return nil
}

// SendRequestRaw send request, trigger any object changes synchronously and return raw json response.
func (state *State) SendRequestRaw(actionState *action.State, f func(ctx context.Context) (json.RawMessage, error)) (json.RawMessage, error) {
	ctx, cl, cancel := state.ReqContext()
	defer cancel()

	raw, err := f(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	state.TriggerEvents(actionState, cl.Changed, cl.Closed)

	return raw, nil
}

// TriggerContextChanges gets ChangeList from context and triggers changes
func (state *State) TriggerContextChanges(ctx context.Context, actionState *action.State) {
	icl := ctx.Value(enigma.ChangeListsKey{})
	if icl == nil {
		return
	}

	cl, ok := icl.(*enigma.ChangeLists)
	if !ok {
		return
	}

	state.TriggerEvents(actionState, cl.Changed, cl.Closed)
}

// GetSessionVariable populates and returns session variables struct
func (state *State) GetSessionVariable(localData interface{}) SessionVariables {
	if state == nil {
		return SessionVariables{Local: localData}
	}

	var session uint64
	var thread uint64

	if state.LogEntry != nil && state.LogEntry.Session != nil {
		session = state.LogEntry.Session.Session
		thread = state.LogEntry.Session.Thread
	}

	sessionVars := SessionVariables{
		Session: session,
		Thread:  thread,
		Local:   localData,
	}

	if state.User != nil {
		sessionVars.User = *state.User
	}

	return sessionVars
}

// ReplaceSessionVariables execute template and replace session variables, e.g. "my app ({{.UserName}})" -> "my app (user_1)"
func (state *State) ReplaceSessionVariables(input *SyncedTemplate) (string, error) {
	return state.ReplaceSessionVariablesWithLocalData(input, nil)
}

// ReplaceSessionVariablesWithLocalData execute template and replace session variables, e.g. "my app ({{.UserName}})" -> "my app (user_1)",
// extra "local" data can be added in addition to the session variables
func (state *State) ReplaceSessionVariablesWithLocalData(input *SyncedTemplate, localData interface{}) (string, error) {
	if input == nil {
		return "", nil
	}

	if state == nil {
		return "", errors.New("nil state")
	}

	if state.LogEntry == nil {
		return "", errors.New("nil LogEntry on state")
	}

	if state.User == nil {
		return "", errors.New("nil User on state")
	}

	if state.LogEntry.Session == nil {
		return "", errors.New("nil Session on LogEntry")
	}

	buf := bytes.NewBuffer(nil)
	if err := input.Execute(buf, state.GetSessionVariable(localData)); err != nil {
		return "", errors.Wrap(err, "failed to execute variables template")
	}
	return buf.String(), nil
}

// Cancel state context
func (state *State) Cancel() {
	if state != nil {
		state.ctxCancel()
	}
}

// WSFailed Should be executed on websocket unexpectedly failing
func (state *State) WSFailed() {
	if state != nil {
		state.LogEntry.LogError(errors.New("websocket unexpectedly closed"))
		state.Cancel()
	}
}
