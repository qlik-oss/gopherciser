package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	getAppStructureSettings struct {
		IncludeRaw bool        `json:"includeRaw,omitempty"`
		Summary    SummaryType `json:"summary,omitempty"`
	}
)

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

	innerAs := appstructure.AppStructure{
		AppMeta: appstructure.AppStructureAppMeta{
			Title: app.Layout.Title,
			Guid:  app.GUID,
		}}

	appStructure := &GeneratedAppStructure{
		innerAs,
		sessionState.LogEntry,
		AppStructureReport{},
		sync.Mutex{},
	}
	appStructure.logEntry = sessionState.LogEntry

	for _, info := range allInfos {
		if info == nil {
			continue
		}
		if err := appStructure.getStructureForObjectAsync(sessionState, actionState, app, info.Id, info.Type, settings.IncludeRaw); err != nil {
			actionState.AddErrors(err)
			return
		}
	}

	appStructure.getFieldListAsync(sessionState, actionState, app)

	if sessionState.Wait(actionState) {
		return // An error occurred
	}

	// todo clicking the "Selections" tab i sense would normally create a fieldlist and a dimensionlist object
	// to get dimensions and fields. We however already have master object dimensions in object list
	// should these be moved to a combined field+dimension list? Leave this as is now and to be decided
	// when implementing actions using fields and decide what works best for GUI.

	raw, err := json.MarshalIndent(appStructure, "", "  ")
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "error marshaling app structure"))
		return
	}

	outputDir := sessionState.OutputsDir
	if outputDir != "" && outputDir[len(outputDir)-1:] != "/" {
		outputDir += "/"
	}

	fileName := fmt.Sprintf("%s%s.structure", outputDir, app.GUID)
	structureFile, err := os.Create(fileName)
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

	appStructure.printSummary(settings.Summary, fileName)
}
