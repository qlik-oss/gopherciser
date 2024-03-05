package session

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	enigma "github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/eventws"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	pending "github.com/qlik-oss/gopherciser/pending"
	"github.com/qlik-oss/gopherciser/randomizer"
	"github.com/qlik-oss/gopherciser/requestmetrics"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/structs"
	"github.com/qlik-oss/gopherciser/synced"
	"github.com/qlik-oss/gopherciser/users"
	"github.com/qlik-oss/gopherciser/wsdialer"
)

type (
	// Rand currently used randomizer for connection
	rand struct {
		mu  sync.Mutex
		rnd helpers.Randomizer
	}

	DefaultRandomizer struct {
		mu sync.Mutex
		*randomizer.Randomizer
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
		Cookies      http.CookieJar
		VirtualProxy string
		Connection   IConnection
		ArtifactMap  *ArtifactMap
		IDMap        IDMap
		HeaderJar    *HeaderJar
		Timeout      time.Duration
		User         *users.User
		OutputsDir   string
		CurrentApp   *ArtifactEntry
		CurrentUser  *structs.User
		Counters     *statistics.ExecutionCounters
		// CurrentActionState will contain the state of the latest action to be started
		CurrentActionState *action.State
		LogEntry           *logger.LogEntry
		EW                 statistics.ErrWarn
		Pending            pending.Handler
		Rest               *RestHandler
		RequestMetrics     *requestmetrics.RequestMetrics
		ReconnectSettings  ReconnectSettings
		Features           Features

		rand          *rand
		trafficLogger enigmahandlers.ITrafficLogger
		reconnect     ReconnectInfo
		targetEnv     string

		ctx       context.Context
		ctxCancel context.CancelFunc

		events  map[int]*Event // todo support multiple events per handle?
		eventMu sync.Mutex

		objects     map[string]ObjectHandlerInstance
		objectsLock sync.Mutex

		eventWs     *eventws.EventWebsocket
		eventWsLock sync.Mutex

		// customStates are used to register custom states when using gopherciser as a library
		customStates     map[string]interface{}
		customStatesLock sync.Mutex

		variables     map[string]interface{}
		variablesLock sync.RWMutex
	}

	// ReconnectSettings settings for re-connecting websocket on unexpected disconnect
	ReconnectSettings struct {
		// Reconnect set to true to attempt reconnecting websocket on disconnect
		Reconnect bool `json:"reconnect" doc-key:"reconnectsettings.reconnect" displayname:"Reconnect on disconnected websocket"`
		// Backoff pattern for reconnect, if empty defaults to defaultReconnectBackoff
		Backoff []float64 `json:"backoff" doc-key:"reconnectsettings.backoff" displayname:"Reconnect backoff pattern"`
	}

	ReconnectInfo struct {
		err                 error
		pendingReconnection pending.Handler
		reconnectFunc       func(bool) (string, error)
	}

	// SessionVariables is used as a data carrier for session variables.
	SessionVariables struct {
		users.User
		Session    uint64
		Thread     uint64
		ScriptVars map[string]interface{}
		Local      interface{}
		Artifacts  *TemplateArtifactMap
	}

	ObjectHandlerInstance interface {
		SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject)
		GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error)
	}

	NoActiveDocError struct {
		Msg string
		Err error
	}

	EventMetricsLogger struct {
		LogEntry *logger.LogEntry
	}
)

const (
	// DefaultTimeout per request timeout
	DefaultTimeout = 300 * time.Second
)

// Fake "events" for event websocket
const (
	EventWsReconnectStarted = "eventws.reconnect.started"
	EventWsReconnectEnded   = "eventws.reconnect.ended"
)

// Standard defined target environments
const (
	TargetEnvQlikSenseOnWindows = "qseow"
	TargetEnvStandaloneEngine   = "engine"

	TargetEnvDefault = TargetEnvStandaloneEngine
)

var (
	defaultReconnectBackoff = wsdialer.DefaultBackoff // current set to same as event ws backoff, but keeping constant so it could be independently changed
)

// Error implements error interface
func (err NoActiveDocError) Error() string {
	if err.Msg == "" {
		err.Msg = "NoActiveDocError"
	}
	if err.Err != nil {
		return fmt.Sprintf("%s: %v", err.Msg, err.Err)
	}
	return err.Msg
}

// Reset child randomizer with new seed
func (rnd *DefaultRandomizer) Reset(instance, session uint64, onlyInstanceSeed bool) {
	rnd.mu.Lock()
	defer rnd.mu.Unlock()

	var seed int64
	if onlyInstanceSeed {
		// Use same random sequence for all users
		seed = randomizer.GetPredictableSeedUInt64(instance, 1)
	} else {
		seed = randomizer.GetPredictableSeedUInt64(instance, session)
	}
	rnd.Randomizer = randomizer.NewSeededRandomizer(seed)
}

func newSessionState(ctx context.Context, outputsDir string, timeout time.Duration, user *users.User, virtualProxy string, counters *statistics.ExecutionCounters) *State {
	sessionCtx, cancel := context.WithCancel(ctx)

	state := &State{
		Timeout:      timeout,
		ArtifactMap:  NewArtifactMap(),
		OutputsDir:   outputsDir,
		User:         user,
		HeaderJar:    NewHeaderJar(),
		VirtualProxy: virtualProxy,
		// Buffer size for the pendingHandler has been chosen after evaluation tests towards sense
		// with medium amount of objects in the sheets. Evaluation was done before introducing spinLoopPending
		// in pendingHandler and could possibly be lowered, this would however require re-evaluation.
		Pending:        pending.NewHandler(32),
		RequestMetrics: &requestmetrics.RequestMetrics{},
		Counters:       counters,
		customStates:   make(map[string]interface{}),
		variables:      make(map[string]interface{}),

		ctx:       sessionCtx,
		ctxCancel: cancel,
		events:    make(map[int]*Event),
		reconnect: ReconnectInfo{
			reconnectFunc:       nil,
			pendingReconnection: pending.NewHandler(32),
		},
	}

	if state.Timeout < time.Millisecond {
		state.Timeout = DefaultTimeout
	}

	return state
}

// New instance of session state
func New(ctx context.Context, outputsDir string, timeout time.Duration, user *users.User, session, instance uint64,
	virtualProxy string, onlyInstanceSeed bool, counters *statistics.ExecutionCounters) *State {

	state := newSessionState(ctx, outputsDir, timeout, user, virtualProxy, counters)

	rnd := &DefaultRandomizer{}
	rnd.Reset(instance, session, onlyInstanceSeed)
	state.SetRandomizer(rnd, false)

	return state
}

// NewWithRandomizer instance of session state with custom randomizer
func NewWithRandomizer(ctx context.Context, outputsDir string, timeout time.Duration, user *users.User, virtualProxy string,
	rnd helpers.Randomizer, counters *statistics.ExecutionCounters) *State {
	state := newSessionState(ctx, outputsDir, timeout, user, virtualProxy, counters)
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
	state.ArtifactMap = NewArtifactMap()
	state.IDMap = IDMap{}
	state.trafficLogger = nil
	state.HeaderJar = NewHeaderJar()
	state.CurrentActionState = nil
	state.EW = statistics.ErrWarn{}
	state.Rest = nil
	state.RequestMetrics = &requestmetrics.RequestMetrics{}
	state.events = make(map[int]*Event)
	state.CurrentApp = nil
	state.CurrentUser = nil
	state.objects = nil
	state.customStates = make(map[string]interface{})
}

// SetLogEntry set the log entry
func (state *State) SetLogEntry(entry *logger.LogEntry) {
	state.LogEntry = entry

	if entry.ShouldLogTraffic() {
		state.trafficLogger = enigmahandlers.NewTrafficLogger(entry, state.Counters)
	} else {
		state.trafficLogger = enigmahandlers.NewTrafficRequestCounter(state.Counters)
	}

	state.Rest = NewRestHandler(state.ctx, state.trafficLogger, state.HeaderJar, state.VirtualProxy, state.Timeout)
}

// TrafficLogger returns the current trafficLogger
func (state *State) TrafficLogger() enigmahandlers.ITrafficLogger {
	return state.trafficLogger
}

// Randomizer get randomizer for session
func (state *State) Randomizer() helpers.Randomizer {
	state.rand.mu.Lock()
	defer state.rand.mu.Unlock()

	return state.rand.rnd
}

// SetRandomizer set randomizer for session, will not be set if already has a randomizer
// on concurrent sets, first instance to acquire lock will "win". When setting to nil, it
// will be automatically forced. Set force flag to have randomizer set even when a randomizer exists.
func (state *State) SetRandomizer(rnd helpers.Randomizer, force bool) {
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
	if state.Rest != nil {
		state.Rest.WaitForPending()
	}
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

// QueueRequestRaw same as QueueRequestAsync but accepts function returning  (json.RawMessage, error), to be used when sending Raw functions
// where the response is not actually used
func (state *State) QueueRequestRaw(f func(ctx context.Context) (json.RawMessage, error), actionState *action.State, failOnError bool, errMsg string) {
	state.QueueRequestWithCallback(func(ctx context.Context) error {
		_, err := f(ctx)
		return err
	}, actionState, failOnError, errMsg, nil)
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
		if errMsg != "" {
			err = errors.Wrap(err, errMsg)
		} else {
			err = errors.WithStack(err)
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

// GetEvent for handle
func (state *State) GetEventFunc(handle int) func(ctx context.Context, actionState *action.State) error {
	state.eventMu.Lock()
	defer state.eventMu.Unlock()
	event, exist := state.events[handle]
	if exist && event != nil {
		return event.F
	}
	return nil
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

type (
	TemplateArtifactMap struct {
		artifactMap *ArtifactMap
	}
)

func (artifacts *TemplateArtifactMap) getArtifact(artifactType, lookup string, cmpType ArtifactEntryCompareType,
	valueLabel string, valueGetter func(*ArtifactEntry) string) (string, error) {
	if artifactType == "" {
		return "", errors.New("first argument artifactType is empty string")
	}
	if lookup == "" {
		return "", errors.Errorf("second argument %s is empty string", cmpType)
	}
	if artifacts == nil {
		return "", errors.New("templateArtifactMap is nil")
	}
	if artifacts.artifactMap == nil {
		return "", errors.New("artifactMap is nil")
	}
	artifact, err := artifacts.artifactMap.Lookup(artifactType, lookup, cmpType)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if artifact == nil {
		return "", errors.New("artifact is nil")
	}
	valueStr := valueGetter(artifact)
	if valueStr == "" {
		return "", errors.Errorf("%s is empty string", valueLabel)
	}
	return valueStr, nil
}

func (artifacts *TemplateArtifactMap) GetIDByTypeAndName(artifactType, name string) (string, error) {
	return artifacts.getArtifact(artifactType, name, ArtifactEntryCompareTypeName, "id", func(artifact *ArtifactEntry) string {
		return artifact.ID
	})
}

func (artifacts *TemplateArtifactMap) GetNameByTypeAndID(artifactType, id string) (string, error) {
	return artifacts.getArtifact(artifactType, id, ArtifactEntryCompareTypeID, "name", func(artifact *ArtifactEntry) string {
		return artifact.Name
	})
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
		Session:    session,
		Thread:     thread,
		Local:      localData,
		ScriptVars: state.variables,
		Artifacts:  &TemplateArtifactMap{state.ArtifactMap},
	}

	if state.User != nil {
		sessionVars.User = *state.User
	}

	return sessionVars
}

// ReplaceSessionVariables execute template and replace session variables, e.g. "my app ({{.UserName}})" -> "my app (user_1)"
func (state *State) ReplaceSessionVariables(input *synced.Template) (string, error) {
	state.variablesLock.RLock() // Lock variables map while replaceing session variables
	defer state.variablesLock.RUnlock()

	return state.ReplaceSessionVariablesWithLocalData(input, nil)
}

// ReplaceSessionVariablesWithLocalData execute template and replace session variables, e.g. "my app ({{.UserName}})" -> "my app (user_1)",
// extra "local" data can be added in addition to the session variables
func (state *State) ReplaceSessionVariablesWithLocalData(input *synced.Template, localData interface{}) (string, error) {
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

	buf := helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)
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
	if state != nil && state.ReconnectSettings.Reconnect {
		state.reconnect.pendingReconnection.IncPending()
		state.LogEntry.LogInfo("WSFailed", "Sense websocket unexpectedly disconnected")
		go func() { // failed executefuncs executed synchronously so this need to be async
			defer state.reconnect.pendingReconnection.DecPending()

			panicErr := helpers.RecoverWithErrorFunc(func() {
				if err := state.Reconnect(); err != nil {
					err = errors.Wrap(err, "failed to reconnect websocket and app")
					state.LogEntry.LogError(err) // report error directly since errors on "Cancel" will be igonored
					state.Cancel()
					return
				}
			})

			if panicErr != nil && state.LogEntry != nil {
				state.LogEntry.LogErrorWithMsg("panic triggering functions on failed connection", panicErr)
			}
		}()
	}
}

// GetObjectHandlerInstance for object ID and type
func (state *State) GetObjectHandlerInstance(id, typ string) ObjectHandlerInstance {
	state.objectsLock.Lock()
	defer state.objectsLock.Unlock()

	if state.objects == nil {
		state.objects = make(map[string]ObjectHandlerInstance)
	}

	instance, ok := state.objects[id]
	if ok {
		return instance
	}

	handler := GlobalObjectHandler.GetObjectHandler(typ)
	instance = handler.Instance(id)
	state.objects[id] = instance

	return instance
}

// AwaitReconnect awaits any reconnect lock to be released
func (state *State) AwaitReconnect() {
	if !state.ReconnectSettings.Reconnect {
		return
	}
	state.reconnect.pendingReconnection.WaitForPending(state.ctx)
}

// Reconnect attempts reconnecting to previously opened app
func (state *State) Reconnect() error {
	state.reconnect.pendingReconnection.IncPending()
	defer state.reconnect.pendingReconnection.DecPending()

	state.LogDebugf("session.state.Reconnect(): entered")
	defer state.LogDebugf("session.state.Reconnect(): exited")

	if !state.ReconnectSettings.Reconnect {
		return nil
	}

	if state.CurrentActionState != nil {
		state.CurrentActionState.AddErrors(wsdialer.DisconnectError{Type: enigmahandlers.SenseWsType})
	}

	reconnectStart := time.Now()
	var attempts int
	defer func() {
		if state.LogEntry != nil {
			state.LogEntry.LogInfo("WebsocketReconnect",
				fmt.Sprintf("success=%v;attempts=%d;TimeSpent=%d", state.reconnect.err == nil, attempts, time.Since(reconnectStart).Milliseconds()))
		}
	}()

	// Get currently subscribed objects
	var subscribedObjects []string
	var sheetObjects []string
	if state.Connection != nil && state.Connection.Sense() != nil {
		subscribedObjects = make([]string, 0, state.Connection.Sense().Objects.Len())
		_ = state.Connection.Sense().Objects.ForEach(func(obj *enigmahandlers.Object) error {
			if obj == nil || obj.ID == "" {
				return nil
			}

			switch obj.Type {
			case enigmahandlers.ObjTypeSheet:
				if sheetObjects == nil {
					sheetObjects = make([]string, 0, 1)
				}
				sheetObjects = append(sheetObjects, obj.ID)
			case enigmahandlers.ObjTypeApp:
				// will be set when re-attaching session
			default:
				subscribedObjects = append(subscribedObjects, obj.ID)
			}

			return nil
		})
	}

	if state == nil || state.reconnect.reconnectFunc == nil {
		state.reconnect.err = nil
		return nil // we should not try to reconnect
	}

	backOff := state.ReconnectSettings.Backoff
	if len(backOff) < 1 {
		backOff = defaultReconnectBackoff
	}

reconnectLoop:
	for i, waitTime := range backOff {
		helpers.WaitFor(state.BaseContext(), time.Duration(waitTime*float64(time.Second)))

		reConnectActionState := &action.State{}

		attempts = i + 1
		if _, err := state.reconnect.reconnectFunc(true); err != nil {
			switch helpers.TrueCause(err).(type) {
			case enigmahandlers.NoSessionOnConnectError:
				state.reconnect.err = errors.WithStack(err)
				break reconnectLoop
			default:
				state.reconnect.err = errors.Wrap(err, "Failed connecting to sense server")
				continue reconnectLoop
			}
		}

		if err := state.SetupChangeChan(); err != nil {
			state.reconnect.err = errors.Wrap(err, "failed to setup change channel")
			continue reconnectLoop
		}

		upLink := state.Connection.Sense()

		doc, err := state.ReOpenDoc(reConnectActionState, upLink)
		if err != nil {
			state.reconnect.err = errors.WithStack(err)
			break reconnectLoop // no active doc, don't try re connecting again
		}

		// set active doc as current app
		if err := upLink.SetCurrentApp(doc.GenericId, doc); err != nil {
			state.reconnect.err = errors.WithStack(err)
			break reconnectLoop
		}

		// Re add any "current" sheets
		for _, id := range sheetObjects {
			if _, _, err := state.GetSheet(reConnectActionState, state.Connection.Sense(), id); err != nil {
				state.reconnect.err = errors.WithStack(err)
				break reconnectLoop
			}
		}

		var wg sync.WaitGroup
		// re-subscribe to objects
		for _, id := range subscribedObjects {
			localId := id
			wg.Add(1)
			state.QueueRequest(func(ctx context.Context) error {
				defer wg.Done()
				GetAndAddObjectSync(state, reConnectActionState, localId)
				return nil
			}, reConnectActionState, true, "")
		}
		wg.Wait()

		state.reconnect.err = reConnectActionState.Errors()
		switch errors.Cause(state.reconnect.err).(type) {
		case nil:
			return nil // successful re-connect
		case NoActiveDocError:
			break reconnectLoop // invalid doc, don't try more re-connects
		}
	}

	return errors.Wrap(state.reconnect.err, "Reconnect error")
}

// SetReconnectFunc sets current app re-connect function
func (state *State) SetReconnectFunc(f func(bool) (string, error)) {
	if state == nil {
		return
	}
	state.reconnect.reconnectFunc = f
}

// GetReconnectError from latest finished reconnect attempt
func (state *State) GetReconnectError() error {
	return state.reconnect.err
}

// IsSenseWebsocketDisconnected checks if error is caused by websocket disconnect
func (state *State) IsSenseWebsocketDisconnected(err error) bool {
	disconnectErr, ok := helpers.TrueCause(err).(wsdialer.DisconnectError)
	if ok && disconnectErr.Type == enigmahandlers.SenseWsType {
		return true
	}

	return false
}

// CurrentSenseApp returns currently set sense app or error if none found
func (state *State) CurrentSenseApp() (*senseobjects.App, error) {
	uplink, err := state.CurrentSenseUplink()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if uplink.CurrentApp == nil {
		return nil, errors.New("no current sense app set")
	}

	return uplink.CurrentApp, nil
}

// CurrentSenseUplink return currently set sense uplink or error if none found
func (state *State) CurrentSenseUplink() (*enigmahandlers.SenseUplink, error) {
	if state == nil {
		return nil, errors.New("nil state")
	}

	if state.Connection == nil {
		return nil, errors.New("no current connection set")
	}

	if state.Connection.Sense() == nil {
		return nil, errors.New("no current sense uplink set")
	}

	return state.Connection.Sense(), nil
}

// SetupEventWebsocketAsync setup event websocket and listener
func (state *State) SetupEventWebsocketAsync(actionState *action.State, nurl neturl.URL, allowuntrusted bool) {
	// change scheme if set to http or https
	switch nurl.Scheme {
	case "http:":
		nurl.Scheme = "ws"
	case "https":
		nurl.Scheme = "wss"
	}
	state.eventWsLock.Lock()
	if state.eventWs != nil {
		if err := state.eventWs.Close(); err != nil && state.LogEntry != nil {
			state.LogEntry.Log(logger.WarningLevel, err)
		}
		state.eventWs = nil
	}

	state.QueueRequest(func(ctx context.Context) error {
		defer state.eventWsLock.Unlock()
		currentActionState := func() *action.State { return state.CurrentActionState }
		metricslogger := &EventMetricsLogger{LogEntry: state.LogEntry}
		var err error
		state.eventWs, err = eventws.SetupEventSocket(ctx, state.BaseContext(), state.Timeout, state.Cookies, state.trafficLogger, metricslogger, &nurl,
			state.HeaderJar.GetHeader(nurl.Hostname()), allowuntrusted, state.RequestMetrics, currentActionState)
		if err != nil {
			return errors.WithStack(err)
		}

		// enable re-connect of event websocket
		state.eventWs.Reconnect.GetContext = state.BaseContext
		state.eventWs.Reconnect.AutoReconnect = true
		state.eventWs.Reconnect.Backoff = state.ReconnectSettings.Backoff
		state.eventWs.Reconnect.SetPending = state.Pending.IncPending
		state.eventWs.Reconnect.UnsetPending = state.Pending.DecPending
		state.eventWs.Reconnect.OnReconnectStart = func() {
			if state == nil {
				return
			}
			if state.LogEntry != nil {
				state.LogEntry.LogDebug("Start re-connect of event websocket")
			}
			state.eventWs.FakeEvent(eventws.Event{
				Operation: EventWsReconnectStarted,
			})
		}
		state.eventWs.Reconnect.OnReconnectDone = func(err error, attempts int, timeSpent time.Duration) {
			if state == nil {
				return
			}
			if state.LogEntry != nil {
				state.LogEntry.LogDebug("End re-connect of event websocket")
				state.LogEntry.LogInfo("EventWsReconnect", fmt.Sprintf("success=%v;attempts=%d;TimeSpent=%d", err == nil, attempts, timeSpent))
			}
			if err != nil && !state.IsAbortTriggered() && state.CurrentActionState != nil {
				state.CurrentActionState.AddErrors(errors.Wrap(err, "error reconnecting event websocket"))
			}
			if state.eventWs != nil {
				state.eventWs.FakeEvent(eventws.Event{
					Operation: EventWsReconnectEnded,
					Success:   err == nil,
				})
			}
		}
		return nil
	}, actionState, true, "")
}

// EventWebsocket returns current established event websocket or nil
func (state *State) EventWebsocket() *eventws.EventWebsocket {
	state.eventWsLock.Lock()
	defer state.eventWsLock.Unlock()

	return state.eventWs
}

func (lgr *EventMetricsLogger) SocketOpenMetric(url *neturl.URL, duration time.Duration) {
	if lgr == nil || lgr.LogEntry == nil {
		return
	}

	path := ""
	if url != nil {
		path = url.Path
	}

	lgr.LogEntry.LogTrafficMetric(duration.Nanoseconds(), 0, 0, -1, path, "Open", "EventWS", "")
}

// ClearObjectSubscriptions and currently subscribed objects
func (state *State) ClearObjectSubscriptions() {
	upLink := state.Connection.Sense()
	// Clear subscribed objects
	clearedObjects, errClearObject := upLink.Objects.ClearObjectsOfType(enigmahandlers.ObjTypeGenericObject)
	if errClearObject != nil {
		state.LogEntry.Log(logger.WarningLevel, clearedObjects)
	}
	state.DeRegisterEvents(clearedObjects)

	// Clear any sheets set
	clearedObjects, errClearObject = upLink.Objects.ClearObjectsOfType(enigmahandlers.ObjTypeSheet)
	if errClearObject != nil {
		state.LogEntry.Log(logger.WarningLevel, clearedObjects)
	}
	state.DeRegisterEvents(clearedObjects)
}

// ClearSubscribedObjects unsubscribes from listed objects
func (state *State) ClearSubscribedObjects(IDs []string) error {
	upLink := state.Connection.Sense()
	for _, id := range IDs {
		obj, err := upLink.Objects.GetObjectByID(id)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := upLink.Objects.ClearObject(obj.Handle); err != nil {
			return errors.WithStack(err)
		}
		state.DeRegisterEvent(obj.Handle)
	}
	return nil
}

// UpdateFeatureMap request features from server and updates feature map
func (state *State) UpdateFeatureMap(host string, actionState *action.State) {
	state.Features.UpdateFeatureMap(state.Rest, host, actionState, state.LogEntry)
}

// AddCustomState add custom data handler to state object
func (state *State) AddCustomState(key string, value interface{}) {
	state.customStatesLock.Lock()
	defer state.customStatesLock.Unlock()

	state.customStates[key] = value
}

// GetCustomState data handler from state object
func (state *State) GetCustomState(key string) (interface{}, bool) {
	state.customStatesLock.Lock()
	defer state.customStatesLock.Unlock()

	value, exist := state.customStates[key]
	return value, exist
}

// Once executes the function with matching string once (per session)
func (state *State) Once(key string, f func()) bool {
	_, exists := state.GetCustomState(key)
	if exists {
		return false
	}
	state.AddCustomState(key, struct{}{})
	f()
	return true
}

// SetVariableValue to session variable map
func (state *State) SetVariableValue(variable string, value interface{}) {
	state.variablesLock.Lock()
	defer state.variablesLock.Unlock()

	state.variables[variable] = value
}

// LogDebugf debug log using current LogEntry
func (state *State) LogDebugf(format string, args ...interface{}) {
	if state.LogEntry != nil {
		state.LogEntry.LogDebugf(format, args...)
	}
}

// LogDetail detail entry using current LogEntry
func (state *State) LogDetail(level logger.LogLevel, msg, detail string) {
	if state.LogEntry != nil {
		state.LogEntry.LogDetail(level, msg, detail)
	}
}

// LogError error log using current LogEntry (remember normaly errors should be added to actionstate if available)
func (state *State) LogError(err error) {
	if state.LogEntry != nil {
		state.LogEntry.LogError(err)
	}
}

// LogInfo info log using current LogEntry
func (state *State) LogInfo(infoType, msg string) {
	if state.LogEntry != nil {
		state.LogEntry.LogInfo(infoType, msg)
	}
}

// LogTrafficMetric traffic metric log using current LogEntry
func (state *State) LogTrafficMetric(responseTime int64, sent, received uint64, requestID int, method, params, trafficType, msg string) {
	if state.LogEntry != nil {
		state.LogEntry.LogTrafficMetric(responseTime, sent, received, requestID, method, params, trafficType, msg)
	}
}

// TargetEnv currently set target enviroment
func (state *State) TargetEnv() string {
	return state.targetEnv
}

// SetTargetEnv set current target environment
func (state *State) SetTargetEnv(env string) {
	state.targetEnv = env
}
