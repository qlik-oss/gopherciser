package scenario

import (
	"context"
	"fmt"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
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

	var wg sync.WaitGroup
	var oldChildInfos []*enigma.NxInfo

	// Clone sheet
	var sheetID string
	cloneObject := func(ctx context.Context) error {
		var err error
		origSheetId := sessionState.IDMap.Get(settings.ID)
		sheetID, err = app.Doc.CloneObject(ctx, origSheetId)

		if err != nil {
			return errors.WithStack(err)
		}

		wg.Add(1)
		// Send GetChildInfos request for original sheet for api compliance
		getSheetChildInfosAsync(sessionState, actionState, app, origSheetId, func(infos []*enigma.NxInfo, err error) {
			defer wg.Done()
			oldChildInfos = infos
		})

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

	sessionState.QueueRequest(func(ctx context.Context) error {
		return nil
	}, actionState, true, fmt.Sprintf("failed to get child infos for sheet<%s>", sheet.ID))

	// change title
	var metaDef *senseobjects.SheetMetaDef
	if sheet.Properties == nil {
		actionState.AddErrors(errors.Errorf("sheet properties are nil"))
		return
	}
	if err := mapstructure.Decode((*sheet.Properties)["qMetaDef"], &metaDef); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to decode metaDef"))
		return
	}

	if metaDef != nil {
		metaDef.Title = fmt.Sprintf("%s (Cloned by %s)", metaDef.Title, sessionState.LogEntry.Session.User)
	} else {
		sessionState.LogEntry.Log(logger.WarningLevel, "sheet metaDef was nil")
	}

	(*sheet.Properties)["qMetaDef"] = metaDef

	wg.Wait()
	var newChildInfos []*enigma.NxInfo
	err = sessionState.SendRequest(actionState, func(ctx context.Context) error {
		newChildInfos, err = sheet.GetChildInfos(ctx)
		return err
	})
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	if len(oldChildInfos) != len(newChildInfos) {
		actionState.AddErrors(
			errors.Errorf("source sheet<%s> child count<%d> not same as destination sheet<%s> child count<%d>", sheetID, len(oldChildInfos), sheet.ID, len(newChildInfos)),
		)
		return
	}

	cellsAny := (*sheet.Properties)["cells"]
	if cellsAny == nil {
		actionState.AddErrors(errors.Errorf("sheet<%s> has no cells", sheet.ID))
		return
	}
	cells, ok := cellsAny.([]any)
	if !ok {
		actionState.AddErrors(errors.Errorf("failed to cast sheet cells type<%T> to []any", cellsAny))
		return
	}

	newCells := make([]map[string]any, len(cells))
	for i, cellAny := range cells {
		cell, ok := cellAny.(map[string]any)
		if !ok {
			actionState.AddErrors(errors.Errorf("failed to cast cell<%d> type<%T> to map[string]any", i, cellAny))
			return
		}
		oldNameAny := cell["name"]
		if oldNameAny == nil {
			actionState.AddErrors(errors.Errorf("item<%d> cell<%v> does not have a name", i, cell))
			return
		}
		oldName, ok := cell["name"].(string)
		if !ok {
			actionState.AddErrors(errors.Errorf("could not cast %v type<%T> to string", oldNameAny, oldNameAny))
			return
		}

		// Find position in old infos
		for j, oldInfo := range oldChildInfos {
			if oldInfo.Id == oldName {
				// replace cell name with that position in newinfo
				cell["name"] = newChildInfos[j].Id
				newCells[i] = cell
				break
			}
		}
	}

	// replace cells with new cells
	(*sheet.Properties)["cells"] = newCells

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

func getSheetChildInfosAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, id string,
	callback func(infos []*enigma.NxInfo, err error)) {
	sessionState.QueueRequestWithCallback(func(ctx context.Context) error {
		sheetObject, err := senseobjects.GetSheet(ctx, app, id)
		if err != nil {
			callback(nil, err)
			return errors.WithStack(err)
		}

		infos, err := sheetObject.GetChildInfos(ctx)
		callback(infos, err)
		return errors.WithStack(err)
	}, actionState, true, fmt.Sprintf("failed to get child infos for sheet<%s>", id), func(err error) {})
}
