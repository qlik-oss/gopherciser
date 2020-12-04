package scenario

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//GenerateOdagSettings settings for GenerateOdag
	ElasticGenerateOdagSettings struct {
		GenerateOdagSettings
	}
)

// Validate ElasticGenerateOdagSettings action (Implements ActionSettings interface)
func (settings ElasticGenerateOdagSettings) Validate() error {
	return settings.GenerateOdagSettings.Validate()
}

// Execute ElasticGenerateOdagSettings action (Implements ActionSettings interface)
func (settings ElasticGenerateOdagSettings) Execute(sessionState *session.State, actionState *action.State,
	connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	odagEndpoint := ElasticOdagEndpointConfiguration
	appGuid := sessionState.CurrentApp.ID
	err := generateOdag(sessionState, settings.GenerateOdagSettings, actionState, connectionSettings, odagEndpoint, appGuid)
	if err != nil {
		actionState.AddErrors(err)
	}
}
