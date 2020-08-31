package scenario

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/eventws"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
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
	getReloadEndopoint = "api/v1/reloads"

	// Delay time between re-connect of event websocket and checking status page if reload is still not done
	StatusCheckDelay = 30 * time.Second
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
			return errors.Errorf("%s settings<%s> are no longer used, remove listed setting(s) from script", ActionElasticReload, strings.Join(hasSettings, ","))
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
	if events == nil {
		actionState.AddErrors(errors.New("Could not get events websocket"))
		return
	}

	reloadEventChan := make(chan *eventws.Event, 10)
	defer close(reloadEventChan)

	eventStartedFunc := events.RegisterFunc(eventws.OperationReloadStarted, func(event eventws.Event) {
		reloadEventChan <- &event
	}, true)
	defer events.DeRegisterFunc(eventStartedFunc)

	eventEndedFunc := events.RegisterFunc(eventws.OperationReloadEnded, func(event eventws.Event) {
		reloadEventChan <- &event
	}, true)
	defer events.DeRegisterFunc(eventEndedFunc)

	// Re-use event structure to "listen" on websocket re-connecting
	checkStatusChan := make(chan struct{})
	statusContext, cancelStatusCheck := context.WithCancel(sessionState.BaseContext())
	wsReconnectFunc := events.RegisterFunc(session.EventWsReconnectEnded, func(event eventws.Event) {
		// If event websocket was re-connected during reload, wait "StatusCheckDelay" then check status page if reload event still hasn't triggered to make sure reload is still ongoing
		helpers.WaitFor(statusContext, StatusCheckDelay)
		if !helpers.IsContextTriggered(statusContext) {
			checkStatusChan <- struct{}{}
		}
	}, false)
	defer events.DeRegisterFunc(wsReconnectFunc)
	defer close(checkStatusChan)

	var postReloadResponse elasticstructs.ReloadResponse
	if err := jsonit.Unmarshal(postReload.ResponseBody, &postReloadResponse); err != nil {
		actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload POST reponse: %s", postReload.ResponseBody)))
		return
	}

	reloadID := postReloadResponse.ID
	reloadStarted := ""
forLoop:
	for {
		select {
		case <-sessionState.BaseContext().Done():
			return
		case event, ok := <-reloadEventChan:
			if !ok {
				actionState.AddErrors(errors.New("reload channel closed unexpectedly"))
				return
			}

			if reloadID == event.ReloadId {
				switch event.Operation {
				case eventws.OperationReloadStarted:
					sessionState.LogEntry.LogDebugf("reload started time<%s>", event.Time)
					reloadStarted = event.Time
				case eventws.OperationReloadEnded:
					sessionState.LogEntry.LogDebugf("reload ended time<%s> success<%v>", event.Time, event.Success)
					if !event.Success {
						actionState.AddErrors(errors.New("reload finished with success false"))
					}
					logReloadDuration(reloadStarted, event.Time, sessionState.LogEntry)
					break forLoop
					// TODO if event websocket was disconnected during reload action, add check of status check after 30s
				}
			}
		case <-time.After(time.Second): // check to verify we didn't have errors and should abort
			if actionState.Failed {
				break
			}
		case <-checkStatusChan:
			host, err := connection.GetRestUrl()
			if err != nil {
				actionState.AddErrors(err)
				return
			}
			// We had a re-connect of event websocket and need to check if reload is still ongoing
			ongoing, err := checkStatusOngoing(sessionState, actionState, host, reloadID)
			if err != nil {
				actionState.AddErrors(err)
				return
			}
			if !ongoing {
				sessionState.LogEntry.Log(logger.WarningLevel, "reload finished while event websocket was down")
				break forLoop
			}
		}
	}
	cancelStatusCheck() // make sure not to try to write on channel after close

	sessionState.Wait(actionState)
}

func logReloadDuration(started, ended string, logEntry *logger.LogEntry) {
	startedTS, err := time.Parse(time.RFC3339, started)
	if err != nil {
		logEntry.Logf(logger.WarningLevel, "failed to parse reload started timestamp: %s", started)
		return
	}

	endedTS, err := time.Parse(time.RFC3339, ended)
	if err != nil {
		logEntry.Logf(logger.WarningLevel, "failed to parse reload ended timestamp: %s", ended)
		return
	}

	duration := endedTS.Sub(startedTS)
	logEntry.LogInfo("ReloadDuration", fmt.Sprintf("%.fs", duration.Seconds())) // format to no decimals since timestamp is with entire seconds only
}

func checkStatusOngoing(sessionState *session.State, actionState *action.State, host, id string) (bool, error) {
	reqOptions := session.DefaultReqOptions()
	statusRequest, err := sessionState.Rest.GetSync(fmt.Sprintf("%s/%s/%s", host, getReloadEndopoint, id), actionState, sessionState.LogEntry, &reqOptions)
	if err != nil {
		return false, errors.WithStack(err)
	}

	var reloadResponse elasticstructs.ReloadResponse
	if err := jsonit.Unmarshal(statusRequest.ResponseBody, &reloadResponse); err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload status reponse: %s", statusRequest.ResponseBody))
	}

	if reloadResponse.Status == statusCreated || reloadResponse.Status == statusQueued || reloadResponse.Status == statusReloading {
		return true, nil
	}
	if reloadResponse.Status == statusSuccess {
		return false, nil
	}
	return false, errors.Errorf("unknown status<%s>", reloadResponse.Status)
}
