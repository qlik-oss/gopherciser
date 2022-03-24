package common

import (
	"context"
	"time"

	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

type DummyScheduler struct{}

func (dummy DummyScheduler) Validate() ([]string, error) { return nil, nil }
func (dummy DummyScheduler) Execute(context.Context, *logger.Log, time.Duration, []scenario.Action, string, users.UserGenerator, *connection.ConnectionSettings, *statistics.ExecutionCounters) error {
	return nil
}
func (dummy DummyScheduler) RequireScenario() bool                        { return false }
func (dummy DummyScheduler) PopulateHookData(data map[string]interface{}) {}
