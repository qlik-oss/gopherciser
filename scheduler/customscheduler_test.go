package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/statistics"
	"github.com/qlik-oss/gopherciser/users"
)

// CustomScheduler for test
type CustomScheduler struct {
	scheduler.Scheduler
	CustomSetting string `json:"customsetting"`
}

func (sched CustomScheduler) Validate() ([]string, error) {
	return nil, nil
}
func (sched CustomScheduler) Execute(context.Context, *logger.Log, time.Duration, []scenario.Action, string, users.UserGenerator, *connection.ConnectionSettings, *statistics.ExecutionCounters) error {
	return nil
}
func (sched CustomScheduler) RequireScenario() bool {
	return false
}

// PopulateHookData populate map with data which can be used by go template in hooks
func (sched CustomScheduler) PopulateHookData(data map[string]interface{}) {}

func TestCustomScheduler(t *testing.T) {
	scheduler.RegisterScheduler("custom", CustomScheduler{})

	rawJson := `{
		"type" : "custom",
		"customsetting" : "MyValue"
	}`
	sched, err := scheduler.UnmarshalScheduler([]byte(rawJson))
	if err != nil {
		t.Fatal(err)
	}
	customSched, ok := sched.(*CustomScheduler)
	if !ok {
		t.Fatalf("scheduler of type %T, expected CustomScheduler", customSched)
	}

	if customSched.CustomSetting != "MyValue" {
		t.Errorf("CustomSetting<%s> expected<MyValue>", customSched.CustomSetting)
	}

	if customSched.SchedType != "custom" {
		t.Errorf("type<%s> expected<custom>", customSched.SchedType)
	}
}
