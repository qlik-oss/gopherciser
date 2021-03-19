package elastic

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//ElasticDeleteOdagSettings settings for ElasticDeleteOdag
	ElasticDeleteOdagSettings struct {
		scenario.DeleteOdagSettings
	}
)

// Validate ElasticDeleteOdagSettings action (Implements ActionSettings interface)
func (settings ElasticDeleteOdagSettings) Validate() error {
	return settings.DeleteOdagSettings.Validate()
}

// Execute ElasticDeleteOdagSettings action (Implements ActionSettings interface)
func (settings ElasticDeleteOdagSettings) Execute(sessionState *session.State, actionState *action.State,
	connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	odagEndpoint := ElasticOdagEndpointConfiguration
	err := scenario.DeleteOdag(sessionState, settings.DeleteOdagSettings, actionState, connectionSettings, odagEndpoint)
	if err != nil {
		actionState.AddErrors(err)
	}
}
