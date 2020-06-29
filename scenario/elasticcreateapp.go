package scenario

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ElasticCreateAppSettings specify app to create
	ElasticCreateAppSettings struct {
		CanAddToCollection
	}
)

// Validate action (Implements ActionSettings interface)
func (settings ElasticCreateAppSettings) Validate() error {
	if settings.Title.String() == "" {
		return errors.New("No Title specified")
	}
	return nil
}

// Execute action (Implements ActionSettings interface)
func (settings ElasticCreateAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
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
}
