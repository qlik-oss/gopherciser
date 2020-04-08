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
		DestinationSpace
	}

	DestinationSpace struct {
		DestinationSpaceId   string `json:"destinationspaceid" displayname:"New space ID" doc-key:"destinationspace.destinationspaceid"`
		DestinationSpaceName string `json:"destinationspacename" displayname:"New space name" doc-key:"destinationspace.destinationspacename"`
	}
)

// UnmarshalJSON unmarshals ElasticMoveSpacesSettings from JSON
func (settings *ElasticMoveSpacesSettings) UnmarshalJSON(arg []byte) error {
	var actionCore DestinationSpace
	if err := jsonit.Unmarshal(arg, &actionCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticMoveSpaces)
	}
	var appSelectCore session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelectCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticMoveSpaces)
	}
	(*settings).DestinationSpace = actionCore
	(*settings).AppSelection = appSelectCore
	return nil
}

// Validate action (Implements ActionSettings interface)
func (settings ElasticMoveSpacesSettings) Validate() error {
	if (settings.DestinationSpaceId == "") == (settings.DestinationSpaceName == "") {
		return errors.New("either specify DestinationSpaceId or DestinationSpaceName")
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
		actionState.AddErrors(errors.Wrap(err, "failed to perform app selection"))
		return
	}

	destSpace, err := settings.ResolveDestinationSpace(sessionState)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	spaceReference := elasticstructs.SpaceReference{SpaceID: destSpace.ID}
	spaceReferenceJson, err := json.Marshal(spaceReference)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	putApp := session.RestRequest{
		Method:      session.PUT,
		ContentType: "application/octet-stream",
		Destination: fmt.Sprintf("%s/api/v1/apps/%s/space", host, entry.GUID),
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

func (settings DestinationSpace) ResolveDestinationSpace(sessionState *session.State) (*elasticstructs.Space, error) {
	var moveToSpace *elasticstructs.Space
	var err error
	if settings.DestinationSpaceId != "" {
		moveToSpace, err = sessionState.ArtifactMap.GetSpaceByID(settings.DestinationSpaceId)
	} else {
		moveToSpace, err = sessionState.ArtifactMap.GetSpaceByName(settings.DestinationSpaceName)
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not find specified destination space")
	}
	return moveToSpace, nil
}
