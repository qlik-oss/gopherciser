package scheduler

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/buildmetrics"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

type (
	// IScheduler interface of scheduler
	IScheduler interface {
		Validate() ([]string, error)
		Execute(
			context.Context,
			*logger.Log,
			time.Duration,
			[]scenario.Action,
			string, // outputsDir
			users.UserGenerator,
			*connection.ConnectionSettings,
			*statistics.ExecutionCounters,
		) error
		RequireScenario() bool
		// PopulateHookData populate map with data which can be used by go template in hooks
		PopulateHookData(data map[string]interface{})
	}

	SchedulerType string

	// Scheduler common core of schedulers
	Scheduler struct {
		// SchedType type of scheduler
		SchedType SchedulerType `json:"type" doc-key:"config.scheduler.type" displayname:"Type"`
		// TimeBuf add wait time in between iterations
		TimeBuf TimeBuffer `json:"iterationtimebuffer" doc-key:"config.scheduler.iterationtimebuffer" displayname:"Iteration time buffer"`
		// InstanceNumber used to ensure different randomizations when running script in multiple different instances
		InstanceNumber uint64 `json:"instance" doc-key:"config.scheduler.instance"  displayname:"Instance"`
		// ReconnectSettings settings for re-connecting websocket on unexpected disconnect
		ReconnectSettings session.ReconnectSettings `json:"reconnectsettings" doc-key:"config.scheduler.reconnectsettings" displayname:"Reconnection settings"`

		ConnectionSettings *connection.ConnectionSettings `json:"-"`
		ContinueOnErrors   bool                           `json:"-"`
	}

	schedulerTmp struct {
		SchedType string `json:"type"`
	}
)

// Core schedulers
const (
	SchedSimple = "simple"
)

// Schedulers need an entry in schedulerHandler
var (
	schedulerHandler map[string]IScheduler
	shLock           sync.Mutex
)

func init() {
	if err := RegisterScheduler(SchedSimple, SimpleScheduler{}); err != nil {
		panic(fmt.Sprint("failed to register simple scheduler", err))
	}
}

// GetEnumMap fakes a scheduler enum for GUI
func (typ SchedulerType) GetEnumMap() *enummap.EnumMap {
	m, err := cpSchedulerHandlerToEnumMap()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprint("failed to convery scheduler handler to enum map:", err))
		return nil
	}
	return m
}

func cpSchedulerHandlerToEnumMap() (*enummap.EnumMap, error) {
	shLock.Lock()
	defer shLock.Unlock()

	schedEnum := enummap.New()
	i := -1
	for schedulerName := range schedulerHandler {
		i++
		if err := schedEnum.Add(schedulerName, i); err != nil {
			return nil, err
		}
	}
	return schedEnum, nil
}

// RegisterScheduler register a custom scheduler, this will fail if scheduler with same name exists
// This should be done as early as possible and must be done before unmarshaling config
func RegisterScheduler(sched string, scheduler IScheduler) error {
	return errors.WithStack(registerScheduler(false, sched, scheduler))
}

// RegisterSchedulerOverride register a custom scheduler and override any existing with same name
// This should be done as early as possible and must be done before unmarshaling config
func RegisterSchedulerOverride(sched string, scheduler IScheduler) error {
	return errors.WithStack(registerScheduler(true, sched, scheduler))
}

func registerScheduler(override bool, sched string, scheduler IScheduler) error {
	shLock.Lock()
	defer shLock.Unlock()

	if schedulerHandler == nil {
		schedulerHandler = make(map[string]IScheduler)
	}

	sched = strings.ToLower(sched)

	if !override && schedulerHandler[sched] != nil {
		return errors.Errorf("scheduler<%s> already registered as type<%T>", sched, schedulerHandler[sched])
	}
	schedulerHandler[sched] = scheduler
	return nil
}

// SchedHandler get scheduler instance of type
func SchedHandler(scheduler string) interface{} {
	shLock.Lock()
	defer shLock.Unlock()

	schedType := schedulerHandler[scheduler]
	if schedType == nil {
		return nil
	}
	return reflect.New(reflect.TypeOf(schedType)).Interface()
}

func setLogEntry(sessionState *session.State, log *logger.Log, session, thread uint64, user string) {
	sessionState.SetLogEntry(log.NewLogEntry())
	sessionState.LogEntry.AddInterceptor(logger.ErrorLevel, onError(sessionState))
	sessionState.LogEntry.AddInterceptor(logger.WarningLevel, onWarning(sessionState))

	// Set user values
	sessionState.LogEntry.SetSessionEntry(&logger.SessionEntry{
		Thread:  thread,
		Session: session,
		User:    user,
	})
}

func onError(sessionState *session.State) func(entry *logger.LogEntry) bool {
	return func(entry *logger.LogEntry) bool {
		sessionState.EW.IncErr()
		sessionState.Counters.Errors.Inc()
		if sessionState.LogEntry != nil && sessionState.LogEntry.Action != nil {
			buildmetrics.ReportError(sessionState.LogEntry.Action.Action, sessionState.LogEntry.Action.Label)
		}
		return true
	}
}

func onWarning(sessionState *session.State) func(entry *logger.LogEntry) bool {
	return func(entry *logger.LogEntry) bool {
		sessionState.EW.IncWarn()
		sessionState.Counters.Warnings.Inc()
		if sessionState.LogEntry != nil && sessionState.LogEntry.Action != nil {
			buildmetrics.ReportWarning(sessionState.LogEntry.Action.Action, sessionState.LogEntry.Action.Label)
		}
		return true
	}
}

// Validate Scheduler settings
func (sched *Scheduler) Validate() error {
	if sched == nil {
		return errors.New("scheduler is nil")
	}

	return sched.TimeBuf.Validate()
}

// SetContinueOnErrors toggles whether to continue on action errors
func (sched *Scheduler) SetContinueOnErrors(enabled bool) error {
	if sched == nil {
		return errors.New("scheduler is nil")
	}
	sched.ContinueOnErrors = enabled
	return nil
}

func (sched *Scheduler) StartNewUser(ctx context.Context, timeout time.Duration, log *logger.Log, userScenario []scenario.Action, thread uint64,
	outputsDir string, user *users.User, iterations int, onlyInstanceSeed bool, counters *statistics.ExecutionCounters) error {

	sessionID := counters.Sessions.Inc()
	instanceID := sched.InstanceNumber
	if instanceID < 1 {
		instanceID = 1
	}
	var iteration int

	sessionState := session.New(ctx, outputsDir, timeout, user, sessionID, instanceID, sched.ConnectionSettings.VirtualProxy, onlyInstanceSeed, counters)
	sessionState.ReconnectSettings = sched.ReconnectSettings

	userName := ""
	if user != nil {
		userName = user.UserName
	}

	counters.ActiveUsers.Inc()
	defer counters.ActiveUsers.Dec()

	buildmetrics.AddUser()
	defer buildmetrics.RemoveUser()

	var mErr *multierror.Error
	for {
		sched.TimeBuf.SetDurationStart(time.Now())

		if helpers.IsContextTriggered(ctx) {
			break
		}

		iteration++
		if iterations > 0 && iteration > iterations {
			break
		}

		if iteration > 1 {
			sessionID := counters.Sessions.Inc()
			sessionState.Randomizer().Reset(instanceID, sessionID, onlyInstanceSeed)
		}

		setLogEntry(sessionState, log, sessionID, thread, userName)

		if err := setupRESTHandler(sessionState, sched.ConnectionSettings); err != nil {
			return errors.WithStack(err)
		}

		err := sched.runIteration(userScenario, sessionState, ctx)
		if err != nil {
			mErr = multierror.Append(mErr, err)
		}

		if err := sched.TimeBuf.Wait(ctx, false); err != nil {
			logEntry := log.NewLogEntry()
			logEntry.Session = sessionState.LogEntry.Session
			logEntry.LogError(errors.Wrap(err, "time buffer in-between sequences failed"))
		}
	}

	return helpers.FlattenMultiError(mErr)
}

func setupRESTHandler(sessionState *session.State, connectionSettings *connection.ConnectionSettings) error {
	headers, err := connectionSettings.GetHeaders(sessionState, "")
	if err != nil {
		return errors.Wrap(err, "failed to get connection settings headers")
	}
	host, err := connectionSettings.GetHost()
	if err != nil {
		return errors.Wrap(err, "failed to extract hostname")
	}
	sessionState.HeaderJar.SetHeader(host, headers)

	client, err := session.DefaultClient(connectionSettings.Allowuntrusted, sessionState)
	if err != nil {
		return errors.WithStack(err)
	}

	sessionState.Rest.SetClient(client)
	return nil
}

func (sched *Scheduler) runIteration(userScenario []scenario.Action, sessionState *session.State, ctx context.Context) error {
	defer sessionState.Reset(ctx)
	defer sessionState.Disconnect() // make sure to disconnect connections at end of iteration
	defer logErrReport(sessionState)

	for _, v := range userScenario {
		err := v.Execute(sessionState, sched.ConnectionSettings)
		if isAborted, _ := scenario.CheckActionError(err); isAborted {
			return nil
		}
		if err != nil {
			if errTimeBuf := sched.TimeBuf.Wait(ctx, true); errTimeBuf != nil {
				logEntry := sessionState.LogEntry.ShallowCopy()
				logEntry.Action = nil
				logEntry.LogError(errors.Wrap(errTimeBuf, "time buffer in-between sequences failed"))
			}
			if !sched.ContinueOnErrors {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func logErrReport(sessionState *session.State) {
	if sessionState == nil {
		_, _ = os.Stderr.WriteString("Failed logging SequenceSummary, session state is nil\n")
		return
	}
	if sessionState.LogEntry == nil {
		_, _ = os.Stderr.WriteString("Failed logging SequenceSummary, LogEntry is nil\n")
		return
	}
	sessionState.LogEntry.LogErrorReport("SequenceSummary", sessionState.EW.TotErrors(), sessionState.EW.TotWarnings())
}

// UnmarshalScheduler unmarshal IScheduler
func UnmarshalScheduler(raw []byte) (IScheduler, error) {
	// use json parser instead for fetching type?
	var tmp schedulerTmp
	if err := json.Unmarshal(raw, &tmp); err != nil {
		return nil, errors.Wrap(err, "failed unmarshaling scheduler type")
	}

	schedType := SchedHandler(tmp.SchedType)
	if err := json.Unmarshal(raw, schedType); err != nil {
		return nil, errors.Wrapf(err, "failed unmarshaling scheduler ")
	}

	if sched, ok := schedType.(IScheduler); ok {
		return sched, nil
	}

	return nil, errors.Errorf("Failed casting scheduler<%T><%v> to IScheduler", schedType, tmp.SchedType)
}
