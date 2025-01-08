package scenario

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// DisconnectEnvironment use after actions in an environment and before connecting to a different environment,
	//The action will disconnect open websockets, including towards sense and event listeners, and thus
	// doesn't need to be combined with a disconnectapp action.
	DisconnectEnvironment struct{}
)

// Validate DisconnectEnvironment action (Implements ActionSettings interface)
func (settings DisconnectEnvironment) Validate() ([]string, error) {
	return nil, nil
}

// Execute DisconnectEnvironment action (Implements ActionSettings interface)
func (settings DisconnectEnvironment) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection != nil {
		if err := sessionState.Connection.Disconnect(); err != nil {
			sessionState.LogEntry.Logf(logger.WarningLevel, "error disconnecting: %v", err)
		}
	}

	// remove any set re-connect function
	sessionState.SetReconnectFunc(nil)

	sessionState.Wait(actionState)
}

// AppStructureAction Implements AppStructureAction interface. It returns if this action should be included
// when doing an "get app structure" from script, IsAppAction tells the scenario
// to insert a "getappstructure" action after that action using data from
// sessionState.CurrentApp. A list of sub actions to be evaluated can also be included.
// AppStructureAction returns if this action should be included when getting app structure
// and any additional sub actions which should also be included.
func (settings *DisconnectEnvironment) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     true,
	}, nil
}
