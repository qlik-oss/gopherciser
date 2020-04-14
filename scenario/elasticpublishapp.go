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
	// ElasticMoveAppSettings settings for moving an app between spaces
	ElasticPublishAppSettings struct {
		session.AppSelection
		DestinationSpace
		ElasticPublishAppSettingsCore
	}

	ElasticPublishAppSettingsCore struct {
		ClearTags bool `json:"cleartags" displayname:"Clear existing tags" doc-key:"elasticpublishapp.cleartags"`
	}
)

// UnmarshalJSON unmarshals ElasticMoveAppSettings from JSON
func (settings *ElasticPublishAppSettings) UnmarshalJSON(arg []byte) error {
	var destSpace DestinationSpace
	if err := jsonit.Unmarshal(arg, &destSpace); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticMoveApp)
	}
	var appSelectCore session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelectCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticMoveApp)
	}
	var actionCore ElasticPublishAppSettingsCore
	if err := jsonit.Unmarshal(arg, &actionCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticMoveApp)
	}
	(*settings).DestinationSpace = destSpace
	(*settings).AppSelection = appSelectCore
	(*settings).ElasticPublishAppSettingsCore = actionCore
	return nil
}

// Validate action (Implements ActionSettings interface)
func (settings ElasticPublishAppSettings) Validate() error {
	if (settings.DestinationSpaceId == "") == (settings.DestinationSpaceName == "") {
		return errors.New("either specify DestinationSpaceId or DestinationSpaceName")
	}
	return nil
}

// Execute action (Implements ActionSettings interface)
func (settings ElasticPublishAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	restHandler := sessionState.Rest

	entry, err := settings.AppSelection.Select(sessionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to perform app selection"))
		return
	}

	destSpace, err := settings.ResolveDestinationSpace(sessionState, actionState, host)
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

	getAppItem := session.RestRequest{
		Method:      session.GET,
		Destination: fmt.Sprintf("%s/api/v1/items/%s", host, entry.ItemID),
	}
	restHandler.QueueRequest(actionState, true, &getAppItem, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		actionState.AddErrors(errors.New("failed to get app item"))
		return
	}
	if getAppItem.ResponseStatusCode != http.StatusOK {
		actionState.AddErrors(errors.Errorf("unexpected response code <%d> when getting app item: %s", getAppItem.ResponseStatusCode, getAppItem.ResponseBody))
	}

	var getItemResponse elasticstructs.CollectionItem
	if err := jsonit.Unmarshal(getAppItem.ResponseBody, &getItemResponse); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed unmarshaling app item response data: %s", getAppItem.ResponseBody))
		return
	}

	publishApp := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%s/api/v1/apps/%s/publish", host, entry.GUID),
		Content:     spaceReferenceJson,
	}

	restHandler.QueueRequest(actionState, true, &publishApp, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		actionState.AddErrors(errors.New("failed during app publish"))
		return
	}
	if publishApp.ResponseStatusCode != http.StatusOK {
		actionState.AddErrors(errors.Errorf("unexpected response code <%d> when posting app to new space: %s", publishApp.ResponseStatusCode, publishApp.ResponseBody))
	}

	appPublishResponseRaw := publishApp.ResponseBody
	var appPublishResponse elasticstructs.AppImportResponse
	if err := jsonit.Unmarshal(appPublishResponseRaw, &appPublishResponse); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed unmarshaling app publish response data: %s", appPublishResponseRaw))
		return
	}

	collectionServiceItemResponse, err := AddItemToCollectionService(sessionState, actionState, appPublishResponse, getItemResponse.Name, host)
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	itemId := collectionServiceItemResponse["id"].(string)

	if settings.ClearTags {
		return
	}

	collectionIds := getItemResponse.CollectionIds
	for _, collectionId := range collectionIds {
		err := AddTag(sessionState, actionState, itemId, host, collectionId)
		if err != nil {
			actionState.AddErrors(errors.Wrapf(err, "failed to add tag: %s", collectionId))
			return
		}
	}
}
