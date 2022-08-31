package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// SetScriptSettings set script
	SetScriptSettings struct {
		Script string `json:"script" displayname:"Script" displayelement:"textarea" doc-key:"setscript.script"`
	}
)

// Validate implements ActionSettings interface
func (settings SetScriptSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute implements ActionSettings interface
func (settings SetScriptSettings) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}
	connection := sessionState.Connection.Sense()

	app := connection.CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	doSet := func(ctx context.Context) error {
		err := app.Doc.SetScript(ctx, settings.Script)
		return err
	}
	if err := sessionState.SendRequest(actionState, doSet); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to set script"))
		return
	}

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error { return app.Doc.DoSave(ctx, "") }); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to save app"))
	}

	sessionState.Wait(actionState)
}
