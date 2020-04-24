package config

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/session"
	"time"
)

type (
	getAppStructureSettings struct{}
)

func (cfg *Config) getAppStructureScenario() []scenario.Action {
	// TODO handle "injectable auth actions" using interface returning action array
	// TODO new action type with parameter showing "app" connect or not
	actionsToKeep := map[string]interface{}{
		"openapp": nil,
	}

	appStructureScenario := make([]scenario.Action, 0, 1)
	for _, act := range cfg.Scenario {
		if _, ok := actionsToKeep[act.Type]; ok {
			appStructureScenario = append(appStructureScenario, act)
			// todo inject GetAppStructureAction if action is action set to handle app connect
			appStructureScenario = append(appStructureScenario, scenario.Action{
				ActionCore: scenario.ActionCore{
					Type:  "GetAppStructure",
					Label: "Get app structure",
				},
				Settings: &getAppStructureSettings{},
			})
		}
	}

	return appStructureScenario
}

// GetAppStructures for all apps in scenario
func (cfg *Config) GetAppStructures(ctx context.Context) error {
	// find all auth and actions
	appStructureScenario := cfg.getAppStructureScenario()
	if len(appStructureScenario) < 1 {
		return errors.New("no applicable actions in scenario")
	}

	// Replace scheduler with 1 iteration 1 user simple scheduler
	cfg.Scheduler = &scheduler.SimpleScheduler{
		Scheduler: scheduler.Scheduler{
			SchedType: scheduler.SchedSimple,
		},
		Settings: scheduler.SimpleSchedSettings{
			ExecutionTime:   -1,
			Iterations:      1,
			ConcurrentUsers: 1,
			RampupDelay:     1.0,
		},
	}

	if err := cfg.Scheduler.Validate(); err != nil {
		return errors.WithStack(err)
	}

	// Setup outputs folder
	// Todo use outputs dir for structure result
	outputsDir, err := setupOutputs(cfg.Settings.OutputsSettings)
	if err != nil {
		return errors.WithStack(err)
	}

	// Todo where to log? Override for now during development
	stmpl, err := session.NewSyncedTemplate("./logs/appstructure.tsv")
	if err != nil {
		return errors.WithStack(err)
	}
	logSettings := LogSettings{
		Traffic:        false,
		Debug:          false,
		TrafficMetrics: false,
		FileName:       *stmpl,
		Format:         0,
		Summary:        0,
	}

	log, err := setupLogging(ctx, logSettings, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	timeout := time.Duration(cfg.Settings.Timeout) * time.Second
	if err := cfg.Scheduler.Execute(ctx, log, timeout, appStructureScenario, outputsDir, cfg.LoginSettings, &cfg.ConnectionSettings); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Execute implements ActionSettings interface
func (settings *getAppStructureSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	fmt.Println("Getting app structure...")
}

// Validate implements ActionSettings interface
func (settings *getAppStructureSettings) Validate() error {
	return nil
}
