package scenario

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// DisconnectElastic environment, use after actions in an elastic environment and before connecting to an windows environment,
	// or custom actions towards other types of environments. The action will disconnect open websockets, including towards sense, and thus
	// doesn't need to be combined with a disconnectapp action.
	DisconnectElastic struct{}
)

// Validate DisconnectElastic action (Implements ActionSettings interface)
func (settings DisconnectElastic) Validate() error {
	return nil
}

// Execute DisconnectElastic action (Implements ActionSettings interface)
func (settings DisconnectElastic) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection != nil {
		if err := sessionState.Connection.Disconnect(); err != nil {
			sessionState.LogEntry.Logf(logger.WarningLevel, "error disconnecting: %v", err)
		}
	}

	// remove any set re-connect function
	sessionState.SetReconnectFunc(nil)

	// close event websocket if open
	if eventWS := sessionState.EventWebsocket(); eventWS != nil {
		if err := eventWS.Close(); err != nil {
			sessionState.LogEntry.Logf(logger.WarningLevel, "error disconnecting event websocket: %v", err)
		}
	}

	sessionState.Wait(actionState)
}

// AppStructureAction Implements AppStructureAction interface. It returns if this action should be included
// when doing an "get app structure" from script, IsAppAction tells the scenario
// to insert a "getappstructure" action after that action using data from
// sessionState.CurrentApp. A list of sub actions to be evaluated can also be included.
// AppStructureAction returns if this action should be included when getting app structure
// and any additional sub actions which should also be included.
func (settings *DisconnectElastic) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     true,
	}, nil
}
