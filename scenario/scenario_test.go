package scenario_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	PanicActionSettings struct{}
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

func (settings PanicActionSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute implements ActionSettings interface
func (settings PanicActionSettings) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	switch sessionState.LogEntry.Session.Session {
	case 1:
		panic("panic 1")
	case 2:
		sessionState.QueueRequest(func(ctx context.Context) error {
			panic("panic 2")
		}, actionState, true, "")
	case 3:
		actionState.AddErrors(errors.New("fail error 1"))
	case 4:
		sessionState.QueueRequest(func(ctx context.Context) error {
			return errors.New("fail error 2")
		}, actionState, true, "")
	default:
		actionState.AddErrors(errors.Errorf("unexpected session number<%d>", sessionState.LogEntry.Session.Session))
	}
	sessionState.Wait(actionState)
}

func TestPanicAndFailRecover(t *testing.T) {
	if err := scenario.RegisterAction("panic", PanicActionSettings{}); err != nil {
		t.Fatal("failed registering panic action 1", err)
	}

	// todo make sure test don't produce log file
	script := `{
	"settings": {
		"timeout": 300,
		"logs": {
			"format" : "no"
		}
	},
	"scheduler": {
		"type": "simple",
		"settings": {
			"executiontime": 10,
			"iterations": 4,
			"rampupdelay": 1.0,
			"concurrentusers": 1
		}
	},
	"loginSettings" : {
		"type" : "none"
	},
    "connectionSettings": {
        "mode": "ws",
        "server": "myserver.example"
	},
	"scenario": [
		{
			"action": "panic"
		}
	]
}`
	var cfg config.Config
	if err := jsonit.Unmarshal([]byte(script), &cfg); err != nil {
		t.Fatal("error unmarshaling config", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := errors.Cause(cfg.Execute(ctx, nil)); err != nil {
		switch err.(type) {
		case *multierror.Error:
			expectedErrs := []string{"PANIC: panic 1", "PANIC: panic 2", "fail error 1", "fail error 2"}

			mErr := err.(*multierror.Error)
			if len(mErr.Errors) != len(expectedErrs) {
				t.Log(err)
				t.Fatalf("multi error contains %d errors, expected %d", len(mErr.Errors), len(expectedErrs))
			}

			for i, v := range mErr.Errors {
				errMsg := strings.SplitN(v.Error(), " Stack", 2)
				if len(errMsg) < 1 || errMsg[0] != expectedErrs[i] {
					t.Errorf("[%d] Unexpected error<%s> expected<%s>", i, v, expectedErrs[i])
				}
			}
		default:
			t.Fatal("error executing script:", err)
		}
	} else {
		t.Fatal("unexpected error from config execution", err)
	}
}
