package scheduler

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
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
	// Type type of scheduler
	Type int
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

		Cancel(msg string)
		SetCancel(cancel func(msg string))
	}

	// Scheduler common core of schedulers
	Scheduler struct {
		// SchedType type of scheduler
		SchedType Type `json:"type" doc-key:"config.scheduler.type"`
		// TimeBuf add wait time in between iterations
		TimeBuf TimeBuffer `json:"iterationtimebuffer" doc-key:"config.scheduler.iterationtimebuffer"`
		// InstanceNumber used to ensure different randomizations when running script in multiple different instances
		InstanceNumber uint64 `json:"instance" doc-key:"config.scheduler.instance"`
		// ReconnectSettings settings for re-connecting websocket on unexpected disconnect
		ReconnectSettings session.ReconnectSettings `json:"reconnectsettings" doc-key:"config.scheduler.reconnectsettings"`
		// MaxErrorCount abort scheduler after error count is reached
		MaxErrorCount uint64 `json:"maxerrors,omitempty" doc-key:"config.scheduler.maxerrors"`

		connectionSettings *connection.ConnectionSettings

		continueOnErrors bool

		cancel func(msg string)
	}

	schedulerTmp struct {
		SchedType Type `json:"type"`
	}
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	// SchedUnknown unknown scheduler
	SchedUnknown Type = iota
	// SchedSimple simple scheduler
	SchedSimple
)

var (
	schedulerTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"simple": int(SchedSimple),
	})
)

func (value Type) GetEnumMap() *enummap.EnumMap {
	return schedulerTypeEnumMap
}

// SchedHandler get scheduler instance of type
// Todo change to logic allowing for external schedulers
func SchedHandler(scheduler Type) interface{} {
	switch scheduler {
	case SchedUnknown:
		return nil
	case SchedSimple:
		return &SimpleScheduler{}
	default:
		return nil
	}
}

// UnmarshalJSON unmarshal scheduler type from json
func (value *Type) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal Type")
	}

	*value = Type(i)

	return nil
}

// MarshalJSON marshal scheduler type to json
func (value Type) MarshalJSON() ([]byte, error) {
	str, err := (*value.GetEnumMap()).String(int(value))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get scheduler type")
	}

	if str == "" {
		return nil, errors.Errorf("Unknown scheduler type<%v>", value)
	}

	return []byte(fmt.Sprintf(`"%s"`, str)), nil
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
	sched.continueOnErrors = enabled
	return nil
}

func (sched *Scheduler) startNewUser(ctx context.Context, timeout time.Duration, log *logger.Log, userScenario []scenario.Action, thread uint64,
	outputsDir string, user *users.User, iterations int, onlyInstanceSeed bool, counters *statistics.ExecutionCounters) error {

	sessionID := counters.Sessions.Inc()
	instanceID := sched.InstanceNumber
	if instanceID < 1 {
		instanceID = 1
	}
	var iteration int

	sessionState := session.New(ctx, outputsDir, timeout, user, sessionID, instanceID, sched.connectionSettings.VirtualProxy, onlyInstanceSeed, counters)
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

		if err := setupRESTHandler(sessionState, sched.connectionSettings); err != nil {
			return errors.WithStack(err)
		}

		err := sched.runIteration(userScenario, sessionState, ctx)
		if err != nil {
			mErr = multierror.Append(mErr, err)
			if sched.MaxErrorCount > 0 && counters.Errors.Current() > sched.MaxErrorCount {
				globalLogEntry := log.NewLogEntry()
				msg := fmt.Sprintf("Max error count of %d surpassed, aborting execution!", sched.MaxErrorCount)
				globalLogEntry.Log(logger.ErrorLevel, msg)

				sched.Cancel(msg)
				break
			}
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
	headers, err := connectionSettings.GetHeaders(sessionState)
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
		err := v.Execute(sessionState, sched.connectionSettings)
		if isAborted, _ := scenario.CheckActionError(err); isAborted {
			return nil
		}
		if err != nil {
			if errTimeBuf := sched.TimeBuf.Wait(ctx, true); errTimeBuf != nil {
				logEntry := sessionState.LogEntry.ShallowCopy()
				logEntry.Action = nil
				logEntry.LogError(errors.Wrap(errTimeBuf, "time buffer in-between sequences failed"))
			}
			if !sched.continueOnErrors {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

// SetCancel function to execute to cancel all executions
func (sched *Scheduler) SetCancel(cancel func(msg string)) {
	if sched == nil {
		return
	}
	sched.cancel = cancel
}

// Cancel all executions
func (sched *Scheduler) Cancel(msg string) {
	if sched == nil {
		return
	}
	if sched.cancel != nil {
		sched.cancel(msg)
	}
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
	if err := jsonit.Unmarshal(raw, &tmp); err != nil {
		return nil, errors.Wrap(err, "failed unmarshaling scheduler type")
	}

	schedType := SchedHandler(tmp.SchedType)
	if err := jsonit.Unmarshal(raw, schedType); err != nil {
		return nil, errors.Wrap(err, "failed unmarshaling scheduler ")
	}

	if sched, ok := schedType.(IScheduler); ok {
		return sched, nil
	}

	return nil, errors.Errorf("Failed casting scheduler<%v> to IScheduler", tmp.SchedType)
}
