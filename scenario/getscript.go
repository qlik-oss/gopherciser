package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// GetscriptSettings Getscript gets load script
	GetscriptSettings struct {
		SaveLog bool `json:"savelog" doc-key:"getscript.savelog"`
	}
)

// Validate GetscriptSettings action (Implements ActionSettings interface)
func (settings GetscriptSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute GetscriptSettings action (Implements ActionSettings interface)
func (settings GetscriptSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {

	if sessionState.Connection == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment (no connection)"))
		return
	}

	if sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment (no uplink)"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}
	var script string
	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		script, err = app.Doc.GetScript(ctx)
		return err
	}); err != nil {
		actionState.AddErrors(err)
		return
	}

	if script != "" && settings.SaveLog {
		sessionState.LogEntry.LogInfo("LoadScript", script)
	}

	sessionState.Wait(actionState)
}
