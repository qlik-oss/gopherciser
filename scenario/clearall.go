package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//ClearAllSettings clear all selections action
	ClearAllSettings struct{}
)

// Validate ClearAll action (Implements ActionSettings interface)
func (settings ClearAllSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute ClearAll action (Implements ActionSettings interface)
func (settings ClearAllSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		if err := app.Doc.ClearAll(ctx, false, ""); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}, actionState, true, "Failed to clear all")

	// Send GetApplayout request
	sessionState.QueueRequest(func(ctx context.Context) error {
		_, err := app.Doc.GetAppLayout(ctx)
		return errors.WithStack(err)
	}, actionState, false, "GetAppLayout request failed")

	sessionState.Wait(actionState)
}
