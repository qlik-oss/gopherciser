package scenario

import (
	"context"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ProductVersionSettings settings for ProductVersion action
	ProductVersionSettings struct {
		// Log product version to log file as InfoType ProductVersion
		Log bool `json:"log" displayname:"Log product version" doc-key:"productversion.log"`
	}
)

// Validate ProductVersion action (Implements ActionSettings interface)
func (settings ProductVersionSettings) Validate() error {
	return nil
}

// Execute ProductVersion action (Implements ActionSettings interface)
func (settings ProductVersionSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		version, err := sessionState.Connection.Sense().Global.ProductVersion(ctx)
		if err != nil {
			return errors.WithStack(err)
		}

		if settings.Log {
			sessionState.LogEntry.LogInfo("ProductVersion", version)
		}

		return nil
	}, actionState, true, "Failed to get product version")

	sessionState.Wait(actionState)
}
