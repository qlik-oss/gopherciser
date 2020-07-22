package scenario

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//DeletionModeEnum defines what apps to remove
	DeletionModeEnum int

	// ElasticDeleteAppCoreSettings
	ElasticDeleteAppCoreSettings struct {
		DeletionMode   DeletionModeEnum `json:"mode" displayname:"Deletion mode" doc-key:"elasticdeleteapp.mode"`
		CollectionName string           `json:"collectionname" displayname:"Collection name" doc-key:"elasticdeleteapp.collectionname"`
	}

	//ElasticDeleteAppSettings specify app to delete
	ElasticDeleteAppSettings struct {
		session.AppSelection
		ElasticDeleteAppCoreSettings
	}

	deprecatedElasticDeleteAppSettings struct {
		AppGUID string `json:"appguid"`
		AppName string `json:"appname"`
	}
)

const (
	// Single delete context app
	Single DeletionModeEnum = iota
	// Everything delete every visible app
	Everything
	// ClearCollection delete all contents of a given collection
	ClearCollection
)

var deletionModeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"single":          int(Single),
	"everything":      int(Everything),
	"clearcollection": int(ClearCollection),
})

// UnmarshalJSON unmarshals delete app settings from JSON
func (settings *ElasticDeleteAppSettings) UnmarshalJSON(arg []byte) error {
	var deprecated deprecatedElasticDeleteAppSettings
	if err := jsonit.Unmarshal(arg, &deprecated); err == nil { // skip check if error
		hasSettings := make([]string, 0, 2)
		if deprecated.AppGUID != "" {
			hasSettings = append(hasSettings, "appguid")
		}
		if deprecated.AppName != "" {
			hasSettings = append(hasSettings, "appname")
		}
		if len(hasSettings) > 0 {
			return errors.Errorf("%s settings<%s> are no longer used", ActionElasticDeleteApp, strings.Join(hasSettings, ","))
		}
	}
	var actionCore ElasticDeleteAppCoreSettings
	if err := jsonit.Unmarshal(arg, &actionCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticDeleteApp)
	}
	var appSelectCore session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelectCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticDeleteApp)
	}
	(*settings).ElasticDeleteAppCoreSettings = actionCore
	(*settings).AppSelection = appSelectCore
	return nil
}

func (value DeletionModeEnum) GetEnumMap() *enummap.EnumMap {
	return deletionModeEnumMap
}

// UnmarshalJSON unmarshal DeletionModeEnum
func (value *DeletionModeEnum) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal DeletionModeEnum")
	}

	*value = DeletionModeEnum(i)
	return nil
}

// MarshalJSON marshal DeletionModeEnum type
func (value DeletionModeEnum) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("unknown DeletionModeEnum<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

const getAppsEndpoint = "api/v1/items"
const deleteAppEngineEndpoint = "api/v1/apps"
const deleteAppEndpoint = "api/v1/items"
const getCollectionEndpoint = "api/v1/collections"

// Validate EfeDeleteApp action (Implements ActionSettings interface)
func (settings ElasticDeleteAppSettings) Validate() error {
	if err := settings.AppSelection.Validate(); err != nil {
		return err
	}
	if settings.DeletionMode == Everything || settings.DeletionMode == ClearCollection {
		if settings.AppSelection.AppMode != session.AppModeCurrent {
			return errors.Errorf("cannot specify any app selection options together with deletion mode 'everything' or 'clearcollection' remove rows or use 'current'")
		}
	}
	if settings.DeletionMode == ClearCollection {
		if settings.CollectionName == "" {
			return errors.Errorf("must specify CollectionName with deletion mode 'clearcollection'")
		}
	}
	return nil
}

// Execute EfeDeleteApp action (Implements ActionSettings interface)
func (settings ElasticDeleteAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	switch settings.DeletionMode {
	case Single:
		entry, err := settings.AppSelection.Select(sessionState)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to perform app selection"))
			return
		}

		err = settings.deleteAppByGuid(host, entry.GUID, sessionState, actionState)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
	case Everything:
		if len(sessionState.ArtifactMap.AppList) == 0 {
			sessionState.LogEntry.Logf(logger.WarningLevel, "deletion mode 'everything' - no apps to delete")
		}
		for _, deleteApp := range sessionState.ArtifactMap.AppList {
			err = settings.deleteAppByGuid(host, deleteApp.GUID, sessionState, actionState)
			if err != nil {
				actionState.AddErrors(err)
				return
			}
		}
	case ClearCollection:
		guidsInCollection, err := settings.appsInCollection(host, sessionState, actionState)
		deletedApps := make([]string, 0, len(guidsInCollection))
		if err != nil {
			actionState.AddErrors(errors.Wrapf(err, "failed to query for apps in collection <%s>", settings.CollectionName))
			return
		}
		for _, deleteApp := range guidsInCollection {
			err = settings.deleteAppByGuid(host, deleteApp, sessionState, actionState)
			if err != nil {
				actionState.AddErrors(err)
			} else {
				deletedApps = append(deletedApps, deleteApp)
			}
		}
		sessionState.LogEntry.LogInfo("NumDeletedApps", fmt.Sprintf("%d", len(deletedApps)))
		sessionState.LogEntry.LogInfo("DeletedApps", strings.Join(deletedApps, ","))
		if len(deletedApps) == 0 {
			sessionState.LogEntry.Logf(logger.WarningLevel, "no apps deleted from collection")
		}
	default:
		actionState.AddErrors(errors.Errorf("deletion mode <%v> not supported", settings.DeletionMode))
		return
	}
}

// Get all apps in a named collection
func (settings ElasticDeleteAppSettings) appsInCollection(host string, sessionState *session.State, actionState *action.State) ([]string, error) {
	restHandler := sessionState.Rest

	query := url.Values{}
	query.Add("name", settings.CollectionName)
	getCollection := session.RestRequest{
		Method:      session.GET,
		Destination: fmt.Sprintf("%v/%v?%v", host, getCollectionEndpoint, query.Encode()),
	}

	restHandler.QueueRequest(actionState, true, &getCollection, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return nil, errors.New("failed during getting list of collections data ")
	}
	if getCollection.ResponseStatusCode != http.StatusOK {
		return nil, errors.Errorf("failed getting list of collections data: %s", getCollection.ResponseStatus)
	}

	collectionDataRaw := getCollection.ResponseBody
	var collectionData elasticstructs.CollectionRequest
	if err := jsonit.Unmarshal(collectionDataRaw, &collectionData); err != nil {
		return nil, errors.Wrap(err, "failed unmarshaling list of collections data")
	}

	if len(collectionData.Data) != 1 {
		return nil, errors.Errorf("expected 1 item looking up collection <%s>, but got %v", settings.CollectionName, len(collectionData.Data))
	}

	guids := make([]string, 0, 1000)
	urlParams := "?limit=10"
	getItemsUrl := fmt.Sprintf("%v/%v/%v/items%v", host, getCollectionEndpoint, collectionData.Data[0].ID, urlParams)

	for getItemsUrl != "" {
		collectionItems, err := settings.getItemsQuery(host, getItemsUrl, sessionState, actionState)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get collection items")
		}
		getItemsUrl = collectionItems.Links.Next.Href
		for _, elem := range collectionItems.Data {
			guid := elem.ResourceID
			guids = append(guids, guid)
		}
	}

	return guids, nil
}

func (settings ElasticDeleteAppSettings) getItemsQuery(host string, url string, sessionState *session.State, actionState *action.State) (*elasticstructs.CollectionItems, error) {
	getItems := session.RestRequest{
		Method:      session.GET,
		Destination: url,
	}
	sessionState.Rest.QueueRequest(actionState, true, &getItems, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return nil, errors.New("failed during get items")
	}
	collectionItemsRaw := getItems.ResponseBody
	var collectionItems *elasticstructs.CollectionItems
	if err := jsonit.Unmarshal(collectionItemsRaw, &collectionItems); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed unmarshaling collection items in <%s>", getItems.ResponseBody))
	}

	return collectionItems, nil
}

func (settings ElasticDeleteAppSettings) deleteAppByGuid(host string, deleteGuid string, sessionState *session.State, actionState *action.State) error {
	restHandler := sessionState.Rest

	// Look up the database ID for the app GUID
	getItems := session.RestRequest{
		Method:      session.GET,
		Destination: fmt.Sprintf("%v/%v?resourceType=app&resourceId=%v", host, getAppsEndpoint, deleteGuid),
	}
	restHandler.QueueRequest(actionState, true, &getItems, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return errors.New("failed during get items")
	}
	var item *elasticstructs.GetItems
	if err := jsonit.Unmarshal(getItems.ResponseBody, &item); err != nil {
		return errors.Wrap(err, "failed to unmarshal getitems")
	}
	if len(item.Data) != 1 {
		return errors.Errorf("expected 1 item looking up GUID <%s>, but got %v", deleteGuid, len(item.Data))
	}
	dbId := item.Data[0].ID

	// First delete the app fromn engine using the app guid
	deleteEngineItem := session.RestRequest{
		Method:      session.DELETE,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%v/%v/%v", host, deleteAppEngineEndpoint, deleteGuid),
	}
	restHandler.QueueRequest(actionState, true, &deleteEngineItem, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return errors.New("failed during delete item from engine")
	}
	if deleteEngineItem.ResponseStatusCode != http.StatusOK {
		if deleteEngineItem.ResponseStatusCode == http.StatusNotFound {
			sessionState.LogEntry.LogError(errors.Errorf("** continue execution for client compliance ** failed to delete app from engine: DELETE %s returns 404", deleteAppEngineEndpoint))
		} else {
			return errors.New(fmt.Sprintf("failed to delete app from engine: %d %s", deleteEngineItem.ResponseStatusCode, deleteEngineItem.ResponseBody))
		}
	}

	// Construct the Delete request with the database ID
	deleteItem := session.RestRequest{
		Method:      session.DELETE,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%v/%v/%v", host, deleteAppEndpoint, dbId),
	}
	restHandler.QueueRequest(actionState, true, &deleteItem, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return errors.New("failed during delete item from collection service")
	}
	if deleteItem.ResponseStatusCode != http.StatusNoContent {
		return errors.New(fmt.Sprintf("failed to delete app from collection service: %d %s", deleteItem.ResponseStatusCode, deleteItem.ResponseBody))
	}

	sessionState.ArtifactMap.DeleteApp(deleteGuid)
	return nil
}
