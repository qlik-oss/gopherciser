package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// DoSaveSettings issues a DoSave command to engine to save the currently opened app
	DoSaveSettings struct{}
)

// Validate DoSaveSettings action (Implements ActionSettings interface)
func (settings DoSaveSettings) Validate() error {
	return nil
}

// Execute DoSaveSettings action (Implements ActionSettings interface)
func (settings DoSaveSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	app, err := sessionState.CurrentSenseApp()
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
	}

	doc := app.Doc
	if doc == nil {
		actionState.AddErrors(errors.New("not connected to sense app"))
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		return errors.WithStack(app.Doc.DoSave(ctx, ""))
	}, actionState, true, "")

	sessionState.Wait(actionState)
}
