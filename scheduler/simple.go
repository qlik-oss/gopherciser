package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/users"
)

type (
	// SimpleSchedSettings simple scheduler settings
	SimpleSchedSettings struct {
		ExecutionTime    int     `json:"executionTime" displayname:"Execution time" doc-key:"config.scheduler.settings.executiontime"` // in seconds
		Iterations       int     `json:"iterations" displayname:"Iterations" doc-key:"config.scheduler.settings.iterations"`
		RampupDelay      float64 `json:"rampupDelay" displayname:"Rampup delay" doc-key:"config.scheduler.settings.rampupdelay"` // in seconds
		ConcurrentUsers  int     `json:"concurrentUsers" displayname:"Concurrent users" displayname:"Rampup delay" doc-key:"config.scheduler.settings.concurrentusers"`
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
func (sched SimpleScheduler) Validate() error {
	// validate inherited settings
	if err := sched.Scheduler.Validate(); err != nil {
		return err
	}

	errorMsg := "Invalid simple scheduler setting: "
	if sched.Settings.ExecutionTime < 1 && sched.Settings.ExecutionTime != -1 {
		return errors.Errorf("%s ExecutionTime<%d>", errorMsg, sched.Settings.ExecutionTime)
	}
	if sched.Settings.Iterations == 0 {
		return errors.Errorf("%s Iterations<%d>", errorMsg, sched.Settings.Iterations)
	}
	if sched.Settings.RampupDelay <= 0 {
		return errors.Errorf("%s RampupDelay<%f>", errorMsg, sched.Settings.RampupDelay)
	}
	if sched.Settings.ConcurrentUsers < 1 && sched.Settings.ConcurrentUsers != -1 {
		return errors.Errorf("%s ConcurrentUsers<%d>", errorMsg, sched.Settings.ConcurrentUsers)
	}
	return nil
}

// Execute execute schedule
func (sched SimpleScheduler) Execute(ctx context.Context, log *logger.Log, timeout time.Duration,
	scenario []scenario.Action, outputsDir string, users users.UserGenerator, connectionSettings *connection.ConnectionSettings) (err error) {

	sched.connectionSettings = connectionSettings

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

	for {
		if helpers.IsContextTriggered(ctx) {
			break
		}

		localThreads++
		if sched.Settings.ConcurrentUsers > 0 && localThreads > sched.Settings.ConcurrentUsers {
			break
		}

		if localThreads != 1 {
			helpers.WaitFor(ctx, time.Duration(sched.Settings.RampupDelay*float64(time.Second)))
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			var err error
			if sched.Settings.ReuseUsers {
				err = sched.iteratorReuseUsers(ctx, timeout, log, scenario, outputsDir, users)
			} else {
				err = sched.iteratorNewUsers(ctx, timeout, log, scenario, outputsDir, users)
			}

			if err != nil {
				func() { // wrapped in function to minimize locking time
					mErrLock.Lock()
					defer mErrLock.Unlock()
					mErr = multierror.Append(mErr, err)

				}()
			}
		}()
	}

	wg.Wait()

	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

func (sched SimpleScheduler) iteratorNewUsers(ctx context.Context, timeout time.Duration, log *logger.Log,
	scenario []scenario.Action, outputsDir string, users users.UserGenerator) (err error) {

	thread := globals.Threads.Inc()

	var (
		mErr      *multierror.Error
		iteration int
	)

	for {
		if helpers.IsContextTriggered(ctx) {
			break
		}

		iteration++
		if sched.Settings.Iterations > 0 && iteration > sched.Settings.Iterations {
			break
		}

		user := users.GetNext()
		err = sched.startNewUser(ctx, timeout, log, scenario, thread, outputsDir, user, sched.connectionSettings, 1, sched.Settings.OnlyInstanceSeed)
		if err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}

	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

func (sched SimpleScheduler) iteratorReuseUsers(ctx context.Context, timeout time.Duration, log *logger.Log,
	scenario []scenario.Action, outputsDir string, users users.UserGenerator) (err error) {

	thread := globals.Threads.Inc()

	var mErr *multierror.Error

	user := users.GetNext()
	err = sched.startNewUser(ctx, timeout, log, scenario, thread, outputsDir, user, sched.connectionSettings, sched.Settings.Iterations, sched.Settings.OnlyInstanceSeed)
	if err != nil {
		mErr = multierror.Append(mErr, err)
	}

	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

// RequireScenario report that scheduler requires a scenario
func (sched *SimpleScheduler) RequireScenario() bool {
	return true
}
