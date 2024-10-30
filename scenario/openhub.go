package scenario

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
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

	StreamsState map[string]string
)

const StreamsStateKey = "streamsState"

// Validate open app scenario item
func (openHub OpenHubSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute execute the action
func (openHub OpenHubSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, setHubStart func()) {
	sessionState.SetTargetEnv(session.TargetEnvQlikSenseOnWindows)

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

	// Create and set a XRF key to headers
	xrfkey := helpers.GenerateXrfKey(sessionState.Randomizer())
	hostUrl, err := url.Parse(host)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to parse hostname from URL"))
		return
	}
	headers := sessionState.HeaderJar.GetHeader(hostUrl.Host)
	headers.Add("X-Qlik-XrfKey", xrfkey)

	// Request done synced
	sessionState.Features.UpdateCapabilities(sessionState.Rest, host, actionState, sessionState.LogEntry)
	if actionState.Failed {
		return
	}
	_, _ = sessionState.Rest.GetSyncWithCallback(fmt.Sprintf("%s/api/hub/about?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		getPrivilegesAsync(sessionState, actionState, host, xrfkey)
	})
	if actionState.Failed {
		return
	}

	// async requests
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/user/info?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/desktoplink?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil)
	fillArtifactsFromStreamsAsync(sessionState, actionState, host, xrfkey)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/reports?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/qvdocuments?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/properties?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/externalproductsignons?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, reqNoError)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/streams?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil) // Send second streams request because client does
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/apps/favorites?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil)

	virtualProxy := ""
	if connectionSettings.VirtualProxy != "" {
		virtualProxy = fmt.Sprintf("/%s", connectionSettings.VirtualProxy)
	}
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/qps/user?targetUri=%s%s/hub/", host, host, virtualProxy), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/qrs/datacollection/settings?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil)

	// These requests will warn only instead of error in case of failure
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/insight-advisor-chat/license?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, reqNoError)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/hub/v1/custombannermessages?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, reqNoError)
	// var QlikCSRFToken string
	noContentOptions := session.DefaultReqOptions()
	noContentOptions.ExpectedStatusCode = []int{http.StatusNoContent}
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/qps/csrftoken", host), actionState, sessionState.LogEntry, noContentOptions, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		// QlikCSRFToken = req.ResponseHeaders.Get("qlik-csrf-token")
	})

	// Client requests features twice, so we do it twice
	sessionState.Features.UpdateCapabilities(sessionState.Rest, host, actionState, sessionState.LogEntry)

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

func fillArtifactsFromStreamsAsync(sessionState *session.State, actionState *action.State, host, xrfkey string) {
	// Get all apps in "Work" and "Published" sections
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/apps/user?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		var stream structs.Stream
		if err := json.Unmarshal(req.ResponseBody, &stream); err != nil {
			actionState.AddErrors(err)
			return
		}
		if err := sessionState.ArtifactMap.FillAppsUsingStream(stream); err != nil {
			actionState.AddErrors(err)
			return
		}
	})

	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/apps/stream/myspace?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		var stream structs.Stream
		if err := json.Unmarshal(req.ResponseBody, &stream); err != nil {
			actionState.AddErrors(err)
			return
		}
		if err := sessionState.ArtifactMap.FillAppsUsingStream(stream); err != nil {
			actionState.AddErrors(err)
			return
		}
	})

	// Get all apps from other streams
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/hub/v1/streams?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		var streams structs.Streams
		if err := json.Unmarshal(req.ResponseBody, &streams); err != nil {
			actionState.AddErrors(err)
			return
		}

		streamsState := make(StreamsState)

		for _, data := range streams.Data {
			if data.Type != structs.StreamsTypeStream {
				continue
			}
			streamsState[data.Attributes.Name] = data.ID
		}

		sessionState.AddCustomState(StreamsStateKey, &streamsState)
	})
}

func getPrivilegesAsync(sessionState *session.State, actionState *action.State, host, xrfkey string) {
	_, _ = sessionState.Rest.GetSyncWithCallback(fmt.Sprintf("%s/api/hub/v1/privileges?xrfkey=%s", host, xrfkey), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil || !sessionState.LogEntry.ShouldLogDebug() {
			return
		}
		var privileges structs.Privileges
		if err := json.Unmarshal(req.ResponseBody, &privileges); err != nil {
			sessionState.LogEntry.Logf(logger.WarningLevel, "failed to unmarshal privileges response: %s", err)
		}
		sessionState.LogEntry.LogDebugf("privileges: %v", privileges)
	})
}
