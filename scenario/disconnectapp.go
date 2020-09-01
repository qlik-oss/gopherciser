package scenario

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// DisconnectAppSettings
	DisconnectAppSettings struct{}
)

// Validate
func (settings DisconnectAppSettings) Validate() error {
	return nil
}

// Execute
func (settings DisconnectAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection != nil {
		if err := sessionState.Connection.Disconnect(); err != nil {
			actionState.AddErrors(err)
		}
	}

	// remove re-connect function
	sessionState.SetReconnectFunc(nil)

	sessionState.Wait(actionState)
}
