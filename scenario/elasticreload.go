package scenario

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/eventws"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ElasticReloadCore Currently used ElasticReloadCore (as opposed to deprecated settings)
	ElasticReloadCore struct{}

	//ElasticReloadSettings specify app to reload
	ElasticReloadSettings struct {
		session.AppSelection
		ElasticReloadCore
	}

	// Older settings no longer used, if exist in JSON, an error will be thrown
	deprecatedElasticReloadSettings struct {
		AppGUID string `json:"appguid"`
		AppName string `json:"appname"`

		PollInterval helpers.TimeDuration `json:"pollinterval" displayname:"Poll interval" doc-key:"elasticreload.pollinterval"`
		SaveLog      bool                 `json:"log" displayname:"Save log" doc-key:"elasticreload.log"`
	}
)

const (
	postReloadEndpoint = "api/v1/reloads"
	getReloadEndpoint  = "api/v1/reloads"
)

const (
	statusCreated   = "CREATED"
	statusQueued    = "QUEUED"
	statusReloading = "RELOADING"
	statusSuccess   = "SUCCEEDED"
)

// UnmarshalJSON unmarshals reload settings from JSON
func (settings *ElasticReloadSettings) UnmarshalJSON(arg []byte) error {
	var deprecated deprecatedElasticReloadSettings
	if err := jsonit.Unmarshal(arg, &deprecated); err == nil { // skip check if error
		hasSettings := make([]string, 0, 2)
		if deprecated.AppGUID != "" {
			hasSettings = append(hasSettings, "appguid")
		}
		if deprecated.AppName != "" {
			hasSettings = append(hasSettings, "appname")
		}
		if deprecated.PollInterval > 0 {
			hasSettings = append(hasSettings, "pollinterval")
		}
		if deprecated.SaveLog {
			hasSettings = append(hasSettings, "log")
		}
		if len(hasSettings) > 0 {
			return errors.Errorf("%s settings<%s> are no longer used, remove this setting/-s from script", ActionElasticReload, strings.Join(hasSettings, ","))
		}
	}
	var core ElasticReloadCore
	if err := jsonit.Unmarshal(arg, &core); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticReload)
	}
	var appSelection session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelection); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionOpenApp)
	}
	*settings = ElasticReloadSettings{appSelection, core}
	return nil
}

// Validate EfeReload action (Implements ActionSettings interface)
func (settings ElasticReloadSettings) Validate() error {
	return nil
}

// Execute EfeReload action (Implements ActionSettings interface)
func (settings ElasticReloadSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	entry, err := settings.AppSelection.Select(sessionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed to perform app selection"))
		return
	}

	reloadGuid := entry.GUID

	postReload := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%s/%s", host, postReloadEndpoint),
		Content:     []byte(fmt.Sprintf("{\"AppID\":\"%s\"}", reloadGuid)),
	}

	sessionState.Rest.QueueRequest(actionState, true, &postReload, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if postReload.ResponseStatusCode != http.StatusCreated {
		actionState.AddErrors(errors.New(fmt.Sprintf("Failed to trigger reload: %s", postReload.ResponseBody)))
		return
	}

	events := sessionState.EventWebsocket()
	reloadEndedChan := make(chan *eventws.Event, 10)
	var eventFunc *eventws.EventFunc
	if events == nil {
		actionState.AddErrors(errors.New("Could not get events websocket"))
		return
	}

	reloadID := ""
	if sessionState.LogEntry.ShouldLogDebug() {
		eventFunc = events.RegisterFunc(eventws.OperationReloadStarted, func(event eventws.Event) {
			if reloadID == event.ReloadId {
				sessionState.LogEntry.LogDebugf("reload started %s", event.Time)
			}
		}, false)
	}

	eventFunc = events.RegisterFunc(eventws.OperationReloadEnded, func(event eventws.Event) {
		if reloadID == event.ReloadId {
			sessionState.LogEntry.LogDebugf("reload ended %s", event.Time)
		}
		reloadEndedChan <- &event
	}, false)

	var postReloadResponse elasticstructs.ReloadResponse
	if err := jsonit.Unmarshal(postReload.ResponseBody, &postReloadResponse); err != nil {
		actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload POST reponse: %s", postReload.ResponseBody)))
		return
	}
	reloadID = postReloadResponse.ID // potential very low risk for race against debug logging, but not important, actual reload event check is safe

	var reloadEndedEvent *eventws.Event

forLoop:
	for {
		select {
		case <-sessionState.BaseContext().Done():
			return
		case event, ok := <-reloadEndedChan:
			if !ok {
				actionState.AddErrors(errors.New("reload channel closed unexpectedly"))
				return
			}
			if reloadID == event.ReloadId {
				reloadEndedEvent = event
				break forLoop
			}
		}
	}

	if eventFunc != nil {
		events.DeRegisterFunc(eventFunc)
		close(reloadEndedChan)
	}

	if !reloadEndedEvent.Success {
		actionState.AddErrors(errors.New("reload finished with success false"))
	}

	// TODO log reload duration
	//if duration, err := time.ParseDuration(reloadTime); err != nil || reloadTime == "" {
	//	sessionState.LogEntry.LogInfo("ReloadDuration", reloadTime)
	//} else {
	//	sessionState.LogEntry.LogInfo("ReloadDuration", fmt.Sprintf("%dms", duration.Milliseconds()))
	//}
}

func checkStatus(sessionState *session.State, actionState *action.State, host, id string) (*elasticstructs.ReloadResponse, error) {
	reqOptions := session.DefaultReqOptions()
	statusRequest, err := sessionState.Rest.GetSync(fmt.Sprintf("%s/%s/%s", host, getReloadEndpoint, id), actionState, sessionState.LogEntry, &reqOptions)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var reloadStatus elasticstructs.ReloadResponse
	if err := jsonit.Unmarshal(statusRequest.ResponseBody, &reloadStatus); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload status reponse: %s", statusRequest.ResponseBody))
	}

	return &reloadStatus, nil
}
