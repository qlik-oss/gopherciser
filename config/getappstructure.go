package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
	"os"
	"sync"
	"time"
)

type (
	getAppStructureSettings struct{}

	// AppObjectDef title and ID of a Sense object
	AppObjectDef struct {
		// Id of object
		Id string
		// Type of Sense object
		Type string
	}

	// AppStructureObject sense object structure
	AppStructureObject struct {
		AppObjectDef
		// Properties of Sense object
		Properties json.RawMessage
		// Children to the sense object
		Children []AppObjectDef
	}

	// AppStructure of Sense app
	AppStructure struct {
		// Objects in Sense app
		Objects []AppStructureObject

		structureLock sync.Mutex
	}
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

// Validate implements ActionSettings interface
func (settings *getAppStructureSettings) Validate() error {
	return nil
}

// Execute implements ActionSettings interface
func (settings *getAppStructureSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	var allInfos []*enigma.NxInfo

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		allInfos, err = app.Doc.GetAllInfos(ctx)
		return err
	}); err != nil {
		actionState.AddErrors(err)
		return
	}

	appStructure := &AppStructure{}

	for _, info := range allInfos {
		if info == nil {
			continue
		}
		if err := getStructureForObjectAsync(sessionState, actionState, app, info.Id, info.Type, appStructure); err != nil {
			actionState.AddErrors(err)
			return
		}
	}

	if sessionState.Wait(actionState) {
		return // An error occurred
	}

	raw, err := json.MarshalIndent(appStructure, "", "  ")
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "error marshaling app structure"))
		return
	}

	outputDir := sessionState.OutputsDir
	if outputDir != "" && outputDir[len(outputDir)-1:] != "/" {
		outputDir += "/"
	}

	structureFile, err := os.Create(fmt.Sprintf("%s%s.structure", outputDir, app.GUID))
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to create structure file"))
		return
	}
	defer func() {
		if err := structureFile.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close file<%v> successfully: %v\n", structureFile, err)
		}
	}()
	if _, err = structureFile.Write(raw); err != nil {
		actionState.AddErrors(errors.Wrap(err, "error while writing to structure file"))
		return
	}
}

func getStructureForObjectAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, id, typ string, appStructure *AppStructure) error {
	// Todo check if visualization
	// Todo get dimensions and measures
	// Todo check for master item connections

	if appStructure == nil {
		return errors.New("appStructure object is nil")
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		obj := AppStructureObject{
			AppObjectDef: AppObjectDef{
				Id:   id,
				Type: typ,
			},
		}

		genObj, err := app.Doc.GetObject(ctx, id)
		if err != nil {
			return errors.WithStack(err)
		}

		obj.Properties, err = genObj.GetPropertiesRaw(ctx)
		if err != nil {
			return errors.WithStack(err)
		}

		childInfos, err := genObj.GetChildInfos(ctx)
		if err != nil {
			return errors.WithStack(err)
		}

		var children []AppObjectDef
		if len(childInfos) > 0 {
			children = make([]AppObjectDef, 0, len(childInfos))
		}

		for _, child := range childInfos {
			if child == nil {
				continue
			}
			children = append(children, AppObjectDef{
				Id:   child.Id,
				Type: child.Type,
			})
		}
		obj.Children = children

		appStructure.AddObject(obj)
		return nil
	}, actionState, true, "")

	return nil
}

func (structure *AppStructure) AddObject(obj AppStructureObject) {
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()
	if structure.Objects == nil {
		structure.Objects = make([]AppStructureObject, 0, 1)
	}
	structure.Objects = append(structure.Objects, obj)
}
