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
	postReloadEndopoint = "api/v1/reloads"
	getReloadEndopoint  = "api/v1/reloads"
)

const (
	statusCreated     = "CREATED"
	statusQueued      = "QUEUED"
	statusReloading   = "RELOADING"
	statusSuccess     = "SUCCEEDED"
	statusInterrupted = "INTERRUPTED"
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
		Destination: fmt.Sprintf("%s/%s", host, postReloadEndopoint),
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

	var postReloadResponse elasticstructs.PostReloadResponse
	if err := jsonit.Unmarshal(postReload.ResponseBody, &postReloadResponse); err != nil {
		actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload POST reponse: %s", postReload.ResponseBody)))
		return
	}

	status := postReloadResponse.Status
	var prevStatus string
	log := ""
	reloadTime := ""

	for status == statusCreated || status == statusQueued || status == statusReloading || status == statusInterrupted {
		helpers.WaitFor(sessionState.BaseContext(), time.Duration(settings.PollInterval))
		if sessionState.IsAbortTriggered() {
			return
		}

		statusRequest := session.RestRequest{
			Method:      session.GET,
			Destination: fmt.Sprintf("%s/%s/%s", host, getReloadEndopoint, postReloadResponse.ID),
		}

		sessionState.Rest.QueueRequest(actionState, true, &statusRequest, sessionState.LogEntry)
		if sessionState.Wait(actionState) {
			return // we had an error
		}

		var reloadStatus elasticstructs.PostReloadResponse
		if err := jsonit.Unmarshal(statusRequest.ResponseBody, &reloadStatus); err != nil {
			actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload status reponse: %s", statusRequest.ResponseBody)))
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

	if status != statusSuccess {
		actionState.AddErrors(errors.Errorf("reload finished with unexpected status <%s>", status))
		return
	}

	if settings.ElasticReloadCore.SaveLog {
		sessionState.LogEntry.LogInfo("ReloadLog", log)
	}
	sessionState.LogEntry.LogInfo("ReloadDuration", reloadTime)
}
