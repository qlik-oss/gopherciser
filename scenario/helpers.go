package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

func getSheet(sessionState *session.State, actionState *action.State, upLink *enigmahandlers.SenseUplink, id string) (*enigmahandlers.Object, *senseobjects.Sheet, error) {
	app := upLink.CurrentApp
	if app == nil {
		err := errors.New("Not connected to a Sense app")
		return nil, nil, err
	}

	var sheet *senseobjects.Sheet
	getSheet := func(ctx context.Context) error {
		var err error
		id = sessionState.IDMap.Get(id)
		sheet, err = senseobjects.GetSheet(ctx, app, id)
		return err
	}
	if err := sessionState.SendRequest(actionState, getSheet); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	if sheet == nil {
		return nil, nil, errors.New("sheet is nil")
	}
	sessionState.LogEntry.LogDebugf("Fetched sheet<%s> successfully", id)

	getProperties := func(ctx context.Context) error {
		_, err := sheet.GetProperties(ctx)
		return err
	}
	if err := sessionState.SendRequest(actionState, getProperties); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	sheetObject := enigmahandlers.NewObject(sheet.Handle, enigmahandlers.ObjTypeSheet, id, sheet)
	if err := upLink.Objects.AddObject(sheetObject); err != nil {
		return nil, nil, errors.Wrap(err, "failed to add object to object list")
	}

	return sheetObject, sheet, nil
}

func subscribeSheetObjectsAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, sheetID string) error {
	sheetID = sessionState.IDMap.Get(sheetID)
	sheetEntry, err := getSheetEntry(sessionState, actionState, app, sheetID)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to objects")
	}

	for _, v := range sheetEntry.Data.Cells {
		sessionState.LogEntry.LogDebugf("subscribe to object<%s> type<%s>", v.Name, v.Type)
		GetAndAddObject(sessionState, actionState, v.Name, v.Type)
	}

	return nil
}

func getSheetEntry(sessionState *session.State, actionState *action.State, app *senseobjects.App, sheetid string) (*senseobjects.SheetNxContainerEntry, error) {
	sheetList, err := app.GetSheetList(sessionState, actionState)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return sheetList.GetSheetEntry(sheetid)
}

// GetAndAddObject get and add object to object handling
func GetAndAddObject(sessionState *session.State, actionState *action.State, name, oType string) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("object<%s> type<%s> found", name, oType)
		sense := sessionState.Connection.Sense()

		var genObj *enigma.GenericObject
		getObject := func(ctx context.Context) error {
			var err error
			genObj, err = sense.CurrentApp.Doc.GetObject(ctx, name)
			return err
		}
		if err := sessionState.SendRequest(actionState, getObject); err != nil {
			return errors.Wrapf(err, "Failed go get object<%s>", name)
		}

		obj, err := sense.AddNewObject(genObj.Handle, enigmahandlers.ObjTypeSheetObject, name, genObj)
		if err != nil {
			return errors.Wrapf(err, "Failed to add object<%s> to object list", name)
		}

		if genObj.GenericType == "auto-chart" {
			sessionState.QueueRequest(func(ctx context.Context) error {
				return getObjectLayout(sessionState, actionState, obj)
			}, actionState, true, "")

			handleAutoChart(sessionState, actionState, genObj, obj)
			return nil
		}

		setObjectDataAndEvents(sessionState, actionState, obj, genObj)

		children := obj.ChildList()
		if children != nil && children.Items != nil {
			sessionState.LogEntry.LogDebugf("object<%s> type<%s> has children", genObj.GenericId, genObj.GenericType)
			for _, child := range children.Items {
				GetAndAddObject(sessionState, actionState, child.Info.Id, child.Info.Type)
			}
		}

		return nil
	}, actionState, true, fmt.Sprintf("Failed to get object<%s>", name))
}

// ResolveAppName return guid or appname with replaced session variables
func ResolveAppName(sessionState *session.State, appguid string, appName *session.SyncedTemplate) (string, error) {
	if appName.String() != "" {
		appName, err := sessionState.ReplaceSessionVariables(appName)
		if err != nil {
			return "", errors.WithStack(err)
		}

		return appName, nil
	}
	return appguid, nil

}

// GetCurrentSheet from objects
func GetCurrentSheet(uplink *enigmahandlers.SenseUplink) (*senseobjects.Sheet, error) {
	sheets := uplink.Objects.GetObjectsOfType(enigmahandlers.ObjTypeSheet)
	if len(sheets) < 1 {
		return nil, errors.New("no current sheet found")
	}
	if len(sheets) > 1 {
		return nil, errors.Errorf("%d current sheets found", len(sheets))
	}
	sheetObj, ok := sheets[0].EnigmaObject.(*senseobjects.Sheet)
	if !ok {
		return nil, errors.Errorf("failed to cast object id<%s> to sheet object", sheetObj.GenericId)
	}
	return sheetObj, nil
}

// ClearCurrentSheet and currently subscribed objects
func ClearCurrentSheet(uplink *enigmahandlers.SenseUplink, sessionState *session.State) {
	clearedObjects, errClearObject := uplink.Objects.ClearObjectsOfType(enigmahandlers.ObjTypeSheetObject)
	if errClearObject != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, clearedObjects)
	}
	sessionState.DeRegisterEvents(clearedObjects)

	clearedObjects, errClearObject = uplink.Objects.ClearObjectsOfType(enigmahandlers.ObjTypeSheet)
	if errClearObject != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, clearedObjects)
	}
	sessionState.DeRegisterEvents(clearedObjects)
}

// Contains check whether any element in the supplied list matches (match func(s string) bool)
func Contains(list []string, match func(s string) bool) bool {
	for _, item := range list {
		if match(item) {
			return true
		}
	}
	return false
}
