package scenario

import (
	"context"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ChangeSheetSettings settings for change sheet action
	ChangeSheetSettings struct {
		ID string `json:"id" displayname:"Sheet ID" doc-key:"changesheet.id"`
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

	uplink := sessionState.Connection.Sense()

	ClearCurrentSheet(uplink, sessionState)

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

	// Get sheet
	if _, _, err := getSheet(sessionState, actionState, uplink, settings.ID); err != nil {
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
