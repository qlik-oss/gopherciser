package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ChangeSheetSettings settings for change sheet action
	ChangeSheetSettings struct {
		ID string `json:"id" displayname:"Sheet ID" doc-key:"changesheet.id" appstructure:"active:sheet"`
	}
)

// Validate change sheet action
func (settings ChangeSheetSettings) Validate() error {
	if settings.ID == "" {
		return errors.Errorf("Change sheet ID is blank")
	}
	return nil
}

// Execute change sheet action
func (settings ChangeSheetSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	actionState.Details = settings.ID

	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}

	uplink := sessionState.Connection.Sense()
	if uplink.CurrentApp == nil {
		actionState.AddErrors(errors.New("not connected to app"))
	}

	ClearObjectSubscriptions(sessionState)

	// Get or create current selection object
	sessionState.QueueRequest(func(ctx context.Context) error {
		if _, err := uplink.CurrentApp.GetCurrentSelections(sessionState, actionState); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}, actionState, true, "failed to create CurrentSelection object")

	// Get locale info
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.GetLocaleInfo(ctx)
		return errors.WithStack(err)
	}, actionState, false, "error getting locale info")

	// Send GetConfiguration request
	sessionState.QueueRequest(func(ctx context.Context) error {
		return errors.WithStack(uplink.Global.RPC(ctx, "GetConfiguration", nil))
	}, actionState, false, "GetConfiguration request failed")

	// Send GetApplayout request
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := uplink.CurrentApp.Doc.GetAppLayout(ctx)
		return errors.WithStack(err)
	}, actionState, false, "GetAppLayout request failed")

	// Get sheet
	if _, _, err := sessionState.GetSheet(actionState, uplink, settings.ID); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to get sheet"))
		return
	}

	// get all objects on sheet
	if err := subscribeSheetObjectsAsync(sessionState, actionState, uplink.CurrentApp, settings.ID); err != nil {
		actionState.AddErrors(err)
		return
	}

	sessionState.Wait(actionState)
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
func (settings ChangeSheetSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	selectables, err := structure.GetSelectables(settings.ID)
	if err != nil {
		return nil, nil, false
	}
	newObjs := appstructure.AppStructurePopulatedObjects{
		Parent:    settings.ID,
		Objects:   make([]appstructure.AppStructureObject, 0),
		Bookmarks: nil,
	}
	newObjs.Objects = append(newObjs.Objects, selectables...)
	return []*appstructure.AppStructurePopulatedObjects{&newObjs}, nil, true
}
