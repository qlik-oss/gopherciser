package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// DuplicateSheetSettings clone object settings
	DuplicateSheetSettings struct {
		// ID of object to clone
		ID string `json:"id" displayname:"Sheet ID" doc-key:"duplicatesheet.id" appstructure:"active:sheet"`
		// ChangeSheet after cloning
		ChangeSheet bool `json:"changesheet" displayname:"Change to sheet after creation" doc-key:"duplicatesheet.changesheet"`
		// Save object changes after clone
		Save bool `json:"save" displayname:"Save sheet" doc-key:"duplicatesheet.save"`
		// CloneID
		CloneID string `json:"cloneid" displayname:"Cloned sheet ID" doc-key:"duplicatesheet.cloneid"`
	}
)

// Execute clone object
func (settings DuplicateSheetSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}

	uplink := sessionState.Connection.Sense()
	app := uplink.CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("not connected to a Sense app"))
		return
	}

	// Clone sheet
	var sheetID string
	cloneObject := func(ctx context.Context) error {
		var err error
		origSheetId := sessionState.IDMap.Get(settings.ID)
		sheetID, err = app.Doc.CloneObject(ctx, origSheetId)

		if err != nil {
			return errors.WithStack(err)
		}

		// Send GetChildInfos request for original sheet for api compliance
		getSheetChildInfosAsync(sessionState, actionState, app, origSheetId)

		return nil
	}
	if err := sessionState.SendRequest(actionState, cloneObject); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to clone object"))
		return
	}
	actionState.Details = sheetID
	if settings.CloneID != "" {
		if err := sessionState.IDMap.Add(settings.CloneID, sheetID, sessionState.LogEntry); err != nil {
			actionState.AddErrors(errors.Wrapf(err, "failed to add key<%s> value<%s> to id map", settings.CloneID, sheetID))
			return
		}
	}

	// Get new sheet
	_, sheet, err := sessionState.GetSheet(actionState, uplink, sheetID)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	// change title
	sheet.Properties.MetaDef.Title = fmt.Sprintf("%s (Cloned by %s)", sheet.Properties.MetaDef.Title, sessionState.LogEntry.Session.User)

	// update sheet properties
	if err := sessionState.SendRequest(actionState, sheet.SetProperties); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	// save new object
	if settings.Save {
		if err := sessionState.SendRequest(actionState, app.Doc.SaveObjects); err != nil {
			actionState.AddErrors(errors.Wrap(err, "Do Save failed"))
			return
		}
	}

	// Send GetChildInfos request for cloned sheet for api compliance
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := sheet.GetChildInfos(ctx)
		return errors.WithStack(err)
	}, actionState, false, fmt.Sprintf("failed to get child infos for sheet<%s>", sheet.ID))

	// Set new sheet as the "active" sheet
	if settings.ChangeSheet {
		sessionState.Wait(actionState) // wait until sheetList has been updated

		// clear current subscribed objects
		clearedObjects, errClearObject := uplink.Objects.ClearObjectsOfType(enigmahandlers.ObjTypeGenericObject)
		if errClearObject != nil {
			sessionState.LogEntry.Log(logger.WarningLevel, clearedObjects)
		}
		sessionState.DeRegisterEvents(clearedObjects)

		// "change" sheet
		if err := subscribeSheetObjectsAsync(sessionState, actionState, app, sheetID); err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}
	}

	sessionState.Wait(actionState)
}

// Validate clone object settings
func (settings DuplicateSheetSettings) Validate() ([]string, error) {
	if settings.ID == "" {
		return nil, errors.New("Duplicate sheet needs an id of a sheet to duplicate")
	}

	return nil, nil
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
func (settings DuplicateSheetSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	if !settings.ChangeSheet {
		return nil, nil, false // Do nothing
	} else {
		return nil, nil, true // Remove previous sheet objects
	}
}

func getSheetChildInfosAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, id string) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sheetObject, err := senseobjects.GetSheet(ctx, app, id)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = sheetObject.GetChildInfos(ctx)
		return errors.WithStack(err)
	}, actionState, false, fmt.Sprintf("failed to get child infos for sheet<%s>", id))
}
