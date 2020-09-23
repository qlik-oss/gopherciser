package scenario

import (
	"context"
	"fmt"
	"net/http"
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
	//ElasticReloadSettings specify app to reload
	ElasticReloadSettings struct {
		session.AppSelection
	}
)

const (
	postReloadEndpoint = "api/v1/reloads"
	getReloadEndpoint  = "api/v1/reloads"

	// Delay time between re-connect of event websocket and checking status page if reload is still not done
	StatusCheckDelay = 30 * time.Second
)

const (
	statusCreated   = "CREATED"
	statusQueued    = "QUEUED"
	statusReloading = "RELOADING"
	statusSuccess   = "SUCCEEDED"
	statusFailed    = "FAILED"
)

// UnmarshalJSON unmarshals reload settings from JSON
func (settings *ElasticReloadSettings) UnmarshalJSON(arg []byte) error {
	if err := HasDeprecatedFields(arg, []string{
		"/appguid",
		"/appname",
		"/pollinterval",
		"/log",
		"/pollingoff",
	}); err != nil {
		return errors.Errorf("%s %s, please remove from script", ActionElasticReload, err.Error())
	}

	var appSelection session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelection); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionOpenApp)
	}
	*settings = ElasticReloadSettings{appSelection}
	return nil
}

// Validate EfeReload action (Implements ActionSettings interface)
func (settings ElasticReloadSettings) Validate() error {
	return nil
}

// Execute EfeReload action (Implements ActionSettings interface)
func (settings ElasticReloadSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	settings.execute(sessionState, actionState, connection)
	sessionState.Wait(actionState)
}

func (settings ElasticReloadSettings) execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings) {
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

	// Create channels and register events
	checkStatusChan := make(chan *eventws.Event)
	statusContext, cancelStatusCheck := context.WithCancel(sessionState.BaseContext())
	reloadEventChan := make(chan *eventws.Event)
	eventEndedFunc := events.RegisterFunc(eventws.OperationResult, func(event eventws.Event) {
		defer helpers.RecoverWithError(nil)
		if !helpers.IsContextTriggered(statusContext) && event.ResourceType == eventws.ResourceTypeReload {
			reloadEventChan <- &event
		}
	}, true)
	// Re-use event structure to "listen" on websocket re-connecting
	wsReconnectFunc := events.RegisterFunc(session.EventWsReconnectEnded, func(event eventws.Event) {
		// If event websocket was re-connected during reload, wait "StatusCheckDelay" then check status page if reload event still hasn't triggered to make sure reload is still ongoing
		helpers.WaitFor(statusContext, StatusCheckDelay)
		if !helpers.IsContextTriggered(statusContext) {
			checkStatusChan <- &event
		}
	}, false)

	defer func() {
		var panicErr error
		func() {
			defer helpers.RecoverWithError(&panicErr)
			events.DeRegisterFunc(wsReconnectFunc)
			cancelStatusCheck()
			events.DeRegisterFunc(eventEndedFunc)
			emptyAndCloseEventChan(checkStatusChan)
			emptyAndCloseEventChan(reloadEventChan)
		}()
		if panicErr != nil {
			actionState.AddErrors(panicErr)
		}
	}()

	var postReloadResponse elasticstructs.ReloadResponse
	if err := jsonit.Unmarshal(postReload.ResponseBody, &postReloadResponse); err != nil {
		actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload POST reponse: %s", postReload.ResponseBody)))
		return
	}

	reloadID := postReloadResponse.ID
forLoop:
	for {
		select {
		case <-sessionState.BaseContext().Done():
			return
		case event, ok := <-reloadEventChan:
			if !ok {
				actionState.AddErrors(errors.New("reload channel closed unexpectedly"))
				break forLoop
			}

			eventReloadID := ""
			switch event.Operation {
			case eventws.OperationResult:
				// reloadiID is in data map instead of event.ReloadID if event is of type reload.result
				if len(event.Data) < 1 {
					actionState.AddErrors(errors.New("reload result event contains no data"))
					break forLoop
				}
				eventReloadIDEntry, ok := event.Data["reloadId"]
				if !ok {
					actionState.AddErrors(errors.New("reload result event contains no reload ID"))
					break forLoop
				}
				eventReloadID, ok = eventReloadIDEntry.(string)
				if !ok {
					actionState.AddErrors(errors.Errorf("reload result event contains reload id of unexpected type<%T> value<%v>", eventReloadIDEntry, eventReloadIDEntry))
					break forLoop
				}
			default:
				eventReloadID = event.ReloadId
			}

			if reloadID == eventReloadID {
				switch event.Operation {
				case eventws.OperationResult:
					sessionState.LogEntry.LogDebugf("reload ended time<%s> success<%v>", event.Time, event.Success)
					if !event.Success {
						actionState.AddErrors(errors.New("reload finished with success false"))
					}
					break forLoop
				}
			}
		case <-checkStatusChan:
			// We had a re-connect of event websocket and need to check if reload is still ongoing
			ongoing, err := checkStatusOngoing(sessionState, actionState, host, reloadID)
			if err != nil {
				actionState.AddErrors(err)
				break forLoop
			}
			if !ongoing {
				sessionState.LogEntry.Log(logger.WarningLevel, "reload finished while event websocket was down")
				break forLoop
			}
		}
	}
	cancelStatusCheck() // make sure not to try to write on channel after close
}

func emptyAndCloseEventChan(c chan *eventws.Event) {
	for {
		select {
		case _, open := <-c:
			if !open {
				return
			}
		default:
			close(c)
			return
		}
	}
}

func checkStatusOngoing(sessionState *session.State, actionState *action.State, host, id string) (bool, error) {
	reqOptions := session.DefaultReqOptions()
	statusRequest, err := sessionState.Rest.GetSync(fmt.Sprintf("%s/%s/%s", host, getReloadEndpoint, id), actionState, sessionState.LogEntry, &reqOptions)
	if err != nil {
		return false, errors.WithStack(err)
	}

	var reloadResponse elasticstructs.ReloadResponse
	if err := jsonit.Unmarshal(statusRequest.ResponseBody, &reloadResponse); err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("failed unmarshaling reload status reponse: %s", statusRequest.ResponseBody))
	}

	switch reloadResponse.Status {
	case statusCreated, statusQueued, statusReloading:
		return true, nil
	case statusSuccess:
		return false, nil
	case statusFailed:
		return false, errors.Errorf("reload failed")
	default:
		return false, errors.Errorf("unknown status<%s>", reloadResponse.Status)
	}
}
