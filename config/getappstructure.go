package config

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	GetAppStructureSettings struct {
		IncludeRaw bool `json:"includeRaw,omitempty"`
		// TruncateStringsTo truncates non significant strings to size if set to > 0
		TruncateStringsTo int `json:"truncate,omitempty"`
		AppStructures     map[string]*GeneratedAppStructure
	}
)

// Validate implements ActionSettings interface
func (settings *GetAppStructureSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute implements ActionSettings interface
func (settings *GetAppStructureSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
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

	structure := &GeneratedAppStructure{
		innerAs,
		sessionState.LogEntry,
		AppStructureReport{},
		sync.Mutex{},
		settings.TruncateStringsTo,
	}
	structure.logEntry = sessionState.LogEntry

	for _, info := range allInfos {
		if info == nil {
			continue
		}
		if err := structure.getStructureForObjectAsync(sessionState, actionState, app, info.Id, info.Type, settings.IncludeRaw); err != nil {
			actionState.AddErrors(err)
			return
		}
	}

	structure.getFieldListAsync(sessionState, actionState, app)

	if sessionState.Wait(actionState) {
		return // An error occurred
	}

	// Get SheetList layout to determine meta for sheets
	sheetList, err := app.GetSheetList(sessionState, actionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "error getting sheetlist"))
		return
	}
	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return sheetList.UpdateLayout(ctx)
	}); err != nil {
		actionState.AddErrors(err)
		return
	}

	sheetListLayout := sheetList.Layout()
	err = structure.addSheetMeta(sheetListLayout)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	settings.AppStructures[app.GUID] = structure
}
