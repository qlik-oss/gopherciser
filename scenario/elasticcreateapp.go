package scenario

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/eventws"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ElasticCreateAppSettings specify app to create
	ElasticCreateAppSettings struct {
		CanAddToCollection
		IgnoreEvents bool `json:"ignoreevents" displayname:"Do not send http requests triggered by web socket events." doc-key:"elasticcreateapp.ignoreevents"`
	}
)

// Validate action (Implements ActionSettings interface)
func (settings ElasticCreateAppSettings) Validate() error {
	if settings.Title.String() == "" {
		return errors.New("No Title specified")
	}
	return nil
}

func allTrue(bools ...bool) bool {
	for _, b := range bools {
		if !b {
			return false
		}
	}
	return true
}

// waitForEvents sends nil on the returned channel when all `conditions` are
// fullfileed by at least one event. An error is sent on the channel if context
// is done before all condition are fullfilled.
func waitForEvents(ctx context.Context, events <-chan *eventws.Event, conditions ...func(*eventws.Event) bool) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		fullfilled := make([]bool, len(conditions))
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case e := <-events:
				for i, cond := range conditions {
					if !fullfilled[i] && cond(e) {
						fullfilled[i] = true
					}
				}
				if allTrue(fullfilled...) {
					errChan <- nil
					return
				}
			}
		}

	}()
	return errChan
}

// Execute action (Implements ActionSettings interface)
func (settings ElasticCreateAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	eventChan := make(chan *eventws.Event, 2)
	reqCtx, _, _ := sessionState.ReqContext()
	ctx, cancel := context.WithCancel(reqCtx)
	defer cancel()

	if !settings.IgnoreEvents {
		events := sessionState.EventWebsocket()
		if events == nil {
			actionState.AddErrors(errors.New("Could not get events websocket"))
			return
		}

		events.RegisterFuncUntilCtxDone(ctx, []string{eventws.OperationUpdated, eventws.OperationCreated}, true,
			func(event eventws.Event) {
				if event.ResourceType == eventws.ResourceTypeItems {
					_ = sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/items/%s", host, event.ResourceID), actionState, sessionState.LogEntry, nil)
					eventChan <- &event
				}
			},
		)
	}

	postAppPayload := make(map[string]interface{})
	attributes := make(map[string]interface{})

	title, err := sessionState.ReplaceSessionVariables(&settings.Title)
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	attributes["name"] = title
	postAppPayload["attributes"] = attributes

	postAppPayloadJson, err := jsonit.Marshal(postAppPayload)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to marshal postAppPayload"))
	}

	postApp := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%v/api/v1/apps", host),
		Content:     postAppPayloadJson,
	}

	sessionState.Rest.QueueRequest(actionState, true, &postApp, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if postApp.ResponseStatusCode != http.StatusOK {
		actionState.AddErrors(errors.Errorf("failed to create app<%d>: %s", postApp.ResponseStatusCode, postApp.ResponseBody))
		return
	}

	// Update created apps global counter
	sessionState.Counters.StatisticsCollector.IncCreatedApps()

	appImportResponseRaw := postApp.ResponseBody
	var appImportResponse elasticstructs.AppImportResponse
	if err := jsonit.Unmarshal(appImportResponseRaw, &appImportResponse); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed unmarshaling app create response data: %s", appImportResponseRaw))
		return
	}

	err = AddAppToCollection(settings.CanAddToCollection, sessionState, actionState, appImportResponse, host)
	if err != nil {
		actionState.AddErrors(err)
	}

	appName := appImportResponse.Attributes.Name
	appItemID, err := sessionState.ArtifactMap.GetAppItemID(appName)
	if err != nil {
		actionState.AddErrors(err)
	}

	if !settings.IgnoreEvents {
		err := <-waitForEvents(ctx, eventChan,
			func(e *eventws.Event) bool {
				return e.Operation == eventws.OperationUpdated && e.ResourceID == appItemID
			},
			func(e *eventws.Event) bool {
				return e.Operation == eventws.OperationCreated && e.ResourceID == appItemID
			},
		)
		if err != nil && !helpers.IsContextTriggered(sessionState.BaseContext()) {
			actionState.AddErrors(errors.Wrapf(err,
				"did not recieve created and updated items events for app<%s> with item id<%s>",
				appName, appItemID,
			))
		}
	}

	sessionState.Wait(actionState)
}
