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

	// TODO log versions from about request?
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/about", host), actionState, sessionState.LogEntry, nil)

	// TODO Save privileges and values?
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/privileges", host), actionState, sessionState.LogEntry, nil)

	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/user/info", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/desktoplink", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/apps/user", host), actionState, sessionState.LogEntry, nil)

	// TODO fill artifactmap
	// GET api/hub/v1/streams
	// <- {"data":[{"type":"Stream","id":"aaec8d41-5201-43ab-809f-3063750dfafd","attributes":{"name":"Everyone","modifiedDate":"2021-03-03T16:22:04.588Z","privileges":["read","publish"]},"relationships":{"owner":{"data":{"type":"User","id":"93b3d9d6-4f00-42a7-8252-5b6cfbfac224"}}}}],"included":[{"type":"User","id":"93b3d9d6-4f00-42a7-8252-5b6cfbfac224","attributes":{"name":"sa_repository","userId":"sa_repository","userDirectory":"INTERNAL","privileges":null}}]}
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/streams", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		// TODO will there be more with more streams?
		// aaec8d41-5201-43ab-809f-3063750dfafd is Everyone stream
		// GET api/hub/v1/apps/stream/aaec8d41-5201-43ab-809f-3063750dfafd
	})

	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/reports", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/qvdocuments", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/properties", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/qps/user?targetUri=%s/header/hub/", host, host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/insight-bot/config", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/hub/qrsData?reloadUri=%s/header/hub/", host, host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/insight-advisor-chat/license", host), actionState, sessionState.LogEntry, nil)

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
