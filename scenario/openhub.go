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

	// Try one request sync first to minimize amount of errors when connection fails.
	_, _ = sessionState.Rest.GetSync(fmt.Sprintf("%s/api/about/v1/language", host), actionState, sessionState.LogEntry, nil)
	if actionState.Failed {
		return
	}

	// TODO save feature flags and values
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/capability/v1/list", host), actionState, sessionState.LogEntry, nil)

	// TODO log versions?
	// GET api/hub/about

	// TODO Save provileges and values?
	// GET api/hub/v1/privileges

	// GET api/hub/v1/user/info

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
