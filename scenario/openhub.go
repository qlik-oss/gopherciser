package scenario

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// OpenHubSettings settings for OpenHub
	OpenHubSettings struct{}
)

// Validate open app scenario item
func (openHub OpenHubSettings) Validate() error {
	return nil
}

// Execute execute the action
func (openHub OpenHubSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, setHubStart func()) {
	// New hub connection, clear any existing apps.
	sessionState.ArtifactMap = session.NewArtifactMap()

	host, err := connectionSettings.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	// Try one request first to minimize amount of errors when connection fails.
	_, _ = sessionState.Rest.GetSync(fmt.Sprintf("%s/api/v1/users/me", host), actionState, sessionState.LogEntry, nil)
	if actionState.Failed {
		return
	}

	// TODO - A lot!

	sessionState.Wait(actionState)
	if err := sessionState.ArtifactMap.LogMap(sessionState.LogEntry); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, err)
	}
}

// AppStructureAction implements AppStructureAction interface
func (openHub OpenHubSettings) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     true,
	}, nil
}
