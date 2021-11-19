package scenario

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/structs"
)

type (
	// OpenHubSettings settings for OpenHub
	OpenHubSettings struct{}
)

// Validate open app scenario item
func (openHub OpenHubSettings) Validate() ([]string, error) {
	return nil, nil
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

	reqNoError := session.DefaultReqOptions()
	reqNoError.FailOnError = false

	sessionState.Features.UpdateCapabilities(sessionState.Rest, host, actionState, sessionState.LogEntry)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/about", host), actionState, sessionState.LogEntry, nil)
	getPrivilegesAsync(sessionState, actionState, host)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/user/info", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/desktoplink", host), actionState, sessionState.LogEntry, nil)
	fillArtifactsFromStreamsAsync(sessionState, actionState, host)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/reports", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/qvdocuments", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/properties", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/qps/user?targetUri=%s/hub/", host, host), actionState, sessionState.LogEntry, nil) // TODO check requests with header

	xrfkey := helpers.GenerateXrfKey(sessionState.Randomizer())
	sessionState.Rest.GetWithHeadersAsync(fmt.Sprintf("%s/qrs/datacollection/settings?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, map[string]string{
		"X-Qlik-XrfKey": xrfkey,
	}, nil, nil)

	// These requests will warn only instead of error in case of failure

	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/insight-bot/config", host), actionState, sessionState.LogEntry, reqNoError)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/insight-advisor-chat/license", host), actionState, sessionState.LogEntry, reqNoError)

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

func fillArtifactsFromStreamsAsync(sessionState *session.State, actionState *action.State, host string) {
	// Get all apps in "Work" and "Published" sections
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/apps/user", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		var stream structs.Stream
		if err := jsonit.Unmarshal(req.ResponseBody, &stream); err != nil {
			actionState.AddErrors(err)
			return
		}
		if err := sessionState.ArtifactMap.FillAppsUsingStream(stream); err != nil {
			actionState.AddErrors(err)
			return
		}
	})

	// Get all apps from other streams
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/streams", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		var streams structs.Streams
		if err := jsonit.Unmarshal(req.ResponseBody, &streams); err != nil {
			actionState.AddErrors(err)
			return
		}

		for _, data := range streams.Data {
			if data.Type != structs.StreamsTypeStream {
				continue
			}

			sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/apps/stream/%s", host, data.ID), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
				if err != nil {
					return
				}
				var stream structs.Stream
				if err = jsonit.Unmarshal(req.ResponseBody, &stream); err != nil {
					actionState.AddErrors(err)
					return
				}

				if err := sessionState.ArtifactMap.FillAppsUsingStream(stream); err != nil {
					actionState.AddErrors(err)
					return
				}
			})
		}
	})
}

func getPrivilegesAsync(sessionState *session.State, actionState *action.State, host string) {
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/privileges", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil || !sessionState.LogEntry.ShouldLogDebug() {
			return
		}
		var privileges structs.Privileges
		if err := jsonit.Unmarshal(req.ResponseBody, &privileges); err != nil {
			sessionState.LogEntry.Logf(logger.WarningLevel, "failed to unmarshal privileges response: %s", err)
		}
		sessionState.LogEntry.LogDebugf("privileges: %v", privileges)
	})
}
