package scenario

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/session"
	"net/http"
)

type (
	// ElasticMoveSpacesSettings settings for moving an app between spaces
	ElasticMoveSpacesSettings struct {
		session.AppSelection
		ElasticMoveSpacesSettingsCore
	}

	ElasticMoveSpacesSettingsCore struct {
		NewSpaceID    string     `json:"newspaceid" displayname:"New space ID" doc-key:"elasticmovespaces.spaceid"`
	}
)

// UnmarshalJSON unmarshals ElasticMoveSpacesSettings from JSON
func (settings *ElasticMoveSpacesSettings) UnmarshalJSON(arg []byte) error {
	var actionCore ElasticMoveSpacesSettingsCore
	if err := jsonit.Unmarshal(arg, &actionCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticMoveSpaces)
	}
	var appSelectCore session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelectCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticMoveSpaces)
	}
	(*settings).ElasticMoveSpacesSettingsCore = actionCore
	(*settings).AppSelection = appSelectCore
	return nil
}

// Validate action (Implements ActionSettings interface)
func (settings ElasticMoveSpacesSettings) Validate() error {
	if settings.NewSpaceID == "" {
		return errors.New("No SpaceID specified")
	}
	return nil
}

// Execute action (Implements ActionSettings interface)
func (settings ElasticMoveSpacesSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
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

	spaceReference := elasticstructs.SpaceReference{settings.NewSpaceID}
	spaceReferenceJson, err := json.Marshal(spaceReference)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	putApp := session.RestRequest{
		Method:        session.PUT,
		ContentType:   "application/octet-stream",
		Destination:   fmt.Sprintf("%s/api/v1/apps/%s/space", host, entry.GUID),
		Content:     spaceReferenceJson,
	}

	restHandler := sessionState.Rest
	restHandler.QueueRequest(actionState, true, &putApp, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		actionState.AddErrors(errors.New("failed during app move"))
		return
	}
	if putApp.ResponseStatusCode != http.StatusOK {
		actionState.AddErrors(errors.Errorf("unexpected response code <%d> when putting app in new space: %s", putApp.ResponseStatusCode, putApp.ResponseBody))
	}
}
