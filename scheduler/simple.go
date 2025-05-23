package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

type (
	// SimpleSchedSettings simple scheduler settings
	SimpleSchedSettings struct {
		ExecutionTime    int     `json:"executionTime" displayname:"Execution time" doc-key:"config.scheduler.settings.executiontime"` // in seconds
		Iterations       int     `json:"iterations" displayname:"Iterations" doc-key:"config.scheduler.settings.iterations"`
		RampupDelay      float64 `json:"rampupDelay" displayname:"Rampup delay" doc-key:"config.scheduler.settings.rampupdelay"` // in seconds
		ConcurrentUsers  int     `json:"concurrentUsers" displayname:"Concurrent users" doc-key:"config.scheduler.settings.concurrentusers"`
		ReuseUsers       bool    `json:"reuseUsers" displayname:"Reuse users" doc-key:"config.scheduler.settings.reuseusers"`
		OnlyInstanceSeed bool    `json:"onlyinstanceseed" displayname:"Only use instance seed" doc-key:"config.scheduler.settings.onlyinstanceseed"`
	}

	// SimpleScheduler simple scheduler
	SimpleScheduler struct {
		Scheduler
		Settings SimpleSchedSettings `json:"settings" doc-key:"config.scheduler.settings"`
	}
)

// Validate schedule
func (sched SimpleScheduler) Validate() ([]string, error) {
	// validate inherited settings
	if err := sched.Scheduler.Validate(); err != nil {
		return nil, err
	}

	errorMsg := "Invalid simple scheduler setting: "
	if sched.Settings.ExecutionTime < 1 && sched.Settings.ExecutionTime != -1 {
		return nil, errors.Errorf("%s ExecutionTime<%d>", errorMsg, sched.Settings.ExecutionTime)
	}
	if sched.Settings.Iterations == 0 {
		return nil, errors.Errorf("%s Iterations<%d>", errorMsg, sched.Settings.Iterations)
	}
	if sched.Settings.RampupDelay <= 0 {
		return nil, errors.Errorf("%s RampupDelay<%f>", errorMsg, sched.Settings.RampupDelay)
	}
	if sched.Settings.ConcurrentUsers < 1 && sched.Settings.ConcurrentUsers != -1 {
		return nil, errors.Errorf("%s ConcurrentUsers<%d>", errorMsg, sched.Settings.ConcurrentUsers)
	}
	return nil, nil
}

// Execute execute schedule
func (sched SimpleScheduler) Execute(ctx context.Context, log *logger.Log, timeout time.Duration, scenario []scenario.Action, outputsDir string,
	users users.UserGenerator, connectionSettings *connection.ConnectionSettings, counters *statistics.ExecutionCounters) (err error) {

	sched.ConnectionSettings = connectionSettings

	if sched.Settings.ExecutionTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(sched.Settings.ExecutionTime)*time.Second)
		defer cancel()
	}

	var (
		wg           sync.WaitGroup
		localThreads int

		mErr     *multierror.Error
		mErrLock sync.Mutex
	)

	// Create user, return true if more users should be created
	addUser := func() bool {
		if helpers.IsContextTriggered(ctx) {
			return false
		}

		localThreads++
		if sched.Settings.ConcurrentUsers > 0 && localThreads > sched.Settings.ConcurrentUsers {
			return false
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := sched.iterator(ctx, timeout, log, scenario, outputsDir, users, counters); err != nil {
				func() { // wrapped in function to minimize locking time
					mErrLock.Lock()
					defer mErrLock.Unlock()
					mErr = multierror.Append(mErr, err)
				}()
			}
		}()

		return true
	}

	ticker := time.NewTicker(time.Duration(sched.Settings.RampupDelay * float64(time.Second)))
	defer ticker.Stop()
	if addUser() {
		for range ticker.C {
			if !addUser() {
				break
			}
		}
	}

	wg.Wait()

	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

func (sched SimpleScheduler) iterator(ctx context.Context, timeout time.Duration, log *logger.Log,
	scenario []scenario.Action, outputsDir string, users users.UserGenerator, counters *statistics.ExecutionCounters) (err error) {

	if counters == nil {
		return errors.New("execution counters are nil")
	}

	thread := counters.Threads.Inc()

	var (
		mErr            *multierror.Error
		iteration       int
		outerIterations = sched.Settings.Iterations
		innerIterations = 1
	)

	if sched.Settings.ReuseUsers {
		outerIterations = 1
		innerIterations = sched.Settings.Iterations
	}

	for !helpers.IsContextTriggered(ctx) {
		iteration++
		if outerIterations > 0 && iteration > outerIterations {
			break
		}

		user := users.GetNext(counters)
		err = sched.StartNewUser(ctx, timeout, log, scenario, thread, outputsDir, user, innerIterations, sched.Settings.OnlyInstanceSeed, counters, nil)
		if err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}

	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

// RequireScenario report that scheduler requires a scenario
func (sched SimpleScheduler) RequireScenario() bool {
	return true
}

// PopulateHookData populate map with data to be used with hooks
func (sched SimpleScheduler) PopulateHookData(data map[string]interface{}) {
	data["ConcurrentUsers"] = sched.Settings.ConcurrentUsers
	data["ExecutionTime"] = sched.Settings.ExecutionTime
	data["Iterations"] = sched.Settings.Iterations
	data["OnlyInstanceSeed"] = sched.Settings.OnlyInstanceSeed
	data["RampupDelay"] = sched.Settings.RampupDelay
	data["ReuseUsers"] = sched.Settings.ReuseUsers
}
