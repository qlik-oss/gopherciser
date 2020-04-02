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
	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/users"
)

type (
	// Type type of scheduler
	Type int
	// IScheduler interface of scheduler
	IScheduler interface {
		Validate() error
		Execute(
			context.Context,
			*logger.Log,
			time.Duration,
			[]scenario.Action,
			string, // outputsDir
			users.UserGenerator,
			*connection.ConnectionSettings,
		) error
		RequireScenario() bool
	}

	// Scheduler common core of schedulers
	Scheduler struct {
		// SchedType type of scheduler
		SchedType Type `json:"type" doc-key:"config.scheduler.type"`
		// TimeBuf add wait time in between iterations
		TimeBuf TimeBuffer `json:"iterationtimebuffer" doc-key:"config.scheduler.iterationtimebuffer"`
		// InstanceNumber used to ensure different randomizations when running script in multiple different instances
		InstanceNumber uint64 `json:"instance" doc-key:"config.scheduler.instance"`

		connectionSettings *connection.ConnectionSettings
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

var schedulerTypeEnumMap, _ = enummap.NewEnumMap(map[string]int{
	"simple": int(SchedSimple),
})

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
		globals.Errors.Inc()
		if sessionState.LogEntry != nil && sessionState.LogEntry.Action != nil {
			buildmetrics.ReportError(sessionState.LogEntry.Action.Action, sessionState.LogEntry.Action.Label)
		}
		return true
	}
}

func onWarning(sessionState *session.State) func(entry *logger.LogEntry) bool {
	return func(entry *logger.LogEntry) bool {
		sessionState.EW.IncWarn()
		globals.Warnings.Inc()
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

func (sched *Scheduler) startNewUser(ctx context.Context, timeout time.Duration, log *logger.Log,
	userScenario []scenario.Action, thread uint64, outputsDir string, user *users.User, iterations int, onlyInstanceSeed bool) error {

	sessionID := globals.Sessions.Inc()
	instanceID := sched.InstanceNumber
	if instanceID < 1 {
		instanceID = 1
	}
	var iteration int
	var mErr *multierror.Error

	sessionState := session.New(ctx, outputsDir, timeout, user, sessionID, instanceID, sched.connectionSettings.VirtualProxy, onlyInstanceSeed)

	userName := ""
	if user != nil {
		userName = user.UserName
	}

	globals.ActiveUsers.Inc()
	defer globals.ActiveUsers.Dec()

	buildmetrics.AddUser()
	defer buildmetrics.RemoveUser()

	for {
		sched.TimeBuf.SetDurationStart(time.Now())

		if helpers.IsContextTriggered(ctx) {
			break
		}

		iteration++
		if iterations > 0 && iteration > iterations {
			break
		}

		setLogEntry(sessionState, log, sessionID, thread, userName)

		err := sched.runIteration(userScenario, sessionState, mErr, ctx)
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

func (sched *Scheduler) runIteration(userScenario []scenario.Action, sessionState *session.State, mErr *multierror.Error, ctx context.Context) error {
	defer sessionState.Reset(ctx)
	defer sessionState.Disconnect() // make sure to disconnect connections at end of iteration
	defer logErrReport(sessionState)

	for _, v := range userScenario {
		if err := v.Execute(sessionState, sched.connectionSettings); err != nil {
			if isAborted, _ := scenario.CheckActionError(err); isAborted {
				return nil
			} else {
				if err := sched.TimeBuf.Wait(ctx, true); err != nil {
					logEntry := sessionState.LogEntry.ShallowCopy()
					logEntry.Action = nil
					logEntry.LogError(errors.Wrap(err, "time buffer in-between sequences failed"))
				}
				return errors.WithStack(err)
			}
		}
	}
	return mErr
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
