package scenario

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/eventws"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ElasticReloadCore Currently used ElasticReloadCore (as opposed to deprecated settings)
	ElasticReloadCore struct {
		// PollInterval time in-between polling for reload status
		PollInterval helpers.TimeDuration `json:"pollinterval" displayname:"Poll interval" doc-key:"elasticreload.pollinterval"`
		SaveLog      bool                 `json:"log" displayname:"Save log" doc-key:"elasticreload.log"`
	}

	//ElasticReloadSettings specify app to reload
	ElasticReloadSettings struct {
		session.AppSelection
		ElasticReloadCore
	}

	// Older settings no longer used, if exist in JSON, an error will be thrown
	deprecatedElasticReloadSettings struct {
		AppGUID string `json:"appguid"`
		AppName string `json:"appname"`
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
		if len(hasSettings) > 0 {
			return errors.Errorf("%s settings<%s> are no longer used", ActionElasticReload, strings.Join(hasSettings, ","))
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
	if time.Duration(settings.PollInterval) < time.Nanosecond {
		settings.PollInterval = constant.ReloadPollInterval
	}

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

	eventFunc = events.RegisterFunc(eventws.OperationReloadEnded, func(event eventws.Event) {
		reloadEndedChan <- &event
	}, false)

	var postReloadResponse elasticstructs.ReloadResponse
	if err := jsonit.Unmarshal(postReload.ResponseBody, &postReloadResponse); err != nil {
		actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload POST reponse: %s", postReload.ResponseBody)))
		return
	}

	status := postReloadResponse.Status
	var prevStatus string
	log := ""
	reloadTime := ""

	// functions for checking and updating status
	updateStatus := func() {
		reloadStatus, err := checkStatus(sessionState, actionState, host, postReloadResponse.ID)
		if err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}
		prevStatus = status
		status = reloadStatus.Status
		if status != prevStatus {
			sessionState.LogEntry.LogDebugf("StatusChange: <%s> -> <%s>", prevStatus, status)
		}
		log = reloadStatus.Log
		reloadTime = reloadStatus.Duration
	}
	statusCheck := func() bool { return status == statusCreated || status == statusQueued || status == statusReloading }

	pollInterval := settings.PollInterval
	for statusCheck() {
		select {
		case event, ok := <-reloadEndedChan:
			if ok && postReloadResponse.AppID == event.ResourceID && postReloadResponse.UserID == event.Origin {
				// event doesn't contain reload ID, we have to check status to be sure it's the correct ID
				updateStatus()
				if statusCheck() {
					// status  updates very slowly. and was not updated for the specific ID was not updated yet, start faster polling.
					// if event was on a different reload ID this will poll every second until the correct one is done which is not optimal
					// but currently nothing we can do about it.
					if pollInterval > helpers.TimeDuration(time.Second) {
						pollInterval = helpers.TimeDuration(time.Second)
					}
				}
			}
		case <-time.After(time.Duration(pollInterval)):
			updateStatus()
		case <-sessionState.BaseContext().Done():
			return
		}
	}

	if eventFunc != nil {
		events.DeRegisterFunc(eventFunc)
		close(reloadEndedChan)
	}

	if status != statusSuccess {
		actionState.AddErrors(errors.Errorf("reload finished with unexpected status <%s>", status))
		return
	}

	if settings.ElasticReloadCore.SaveLog {
		sessionState.LogEntry.LogInfo("ReloadLog", log)
	}

	if duration, err := time.ParseDuration(reloadTime); err != nil || reloadTime == "" {
		sessionState.LogEntry.LogInfo("ReloadDuration", reloadTime)
	} else {
		sessionState.LogEntry.LogInfo("ReloadDuration", fmt.Sprintf("%dms", duration.Milliseconds()))
	}
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
