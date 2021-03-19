package elastic

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
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//DeletionModeEnum defines what apps to remove
	DeletionModeEnum int

	// ElasticDeleteAppCoreSettings settings core used by Unmarshal interface
	ElasticDeleteAppCoreSettings struct {
		DeletionMode   DeletionModeEnum `json:"mode" displayname:"Deletion mode" doc-key:"elasticdeleteapp.mode"`
		CollectionName string           `json:"collectionname" displayname:"Collection name" doc-key:"elasticdeleteapp.collectionname"`
	}

	//ElasticDeleteAppSettings specify app to delete
	ElasticDeleteAppSettings struct {
		session.AppSelection
		ElasticDeleteAppCoreSettings
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
	// Check for deprecated fields
	if err := helpers.HasDeprecatedFields(arg, []string{
		"/appguid",
		"/appname",
	}); err != nil {
		return errors.Errorf("%s %s, please remove from script", ActionElasticDeleteApp, err.Error())
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

// GetEnumMap return enum map to gui
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

const deleteAppEngineEndpoint = "api/v1/apps"
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

		err = settings.deleteAppByGuid(host, entry.ID, sessionState, actionState)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
	case Everything:
		if sessionState.ArtifactMap.Count(session.ResourceTypeApp) == 0 {
			sessionState.LogEntry.Logf(logger.WarningLevel, "deletion mode 'everything' - no apps to delete")
		}

		for entry := sessionState.ArtifactMap.First(session.ResourceTypeApp); entry != nil; entry = sessionState.ArtifactMap.First(session.ResourceTypeApp) {
			if err := settings.deleteAppByGuid(host, entry.ID, sessionState, actionState); err != nil {
				actionState.AddErrors(err)
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
	options := session.DefaultReqOptions()
	options.ExpectedStatusCode = []int{http.StatusOK, http.StatusNotFound}
	deleteAppRequest := sessionState.Rest.DeleteAsync(fmt.Sprintf("%v/%v/%v", host, deleteAppEngineEndpoint, deleteGuid), actionState, sessionState.LogEntry, nil)

	if sessionState.Wait(actionState) {
		return errors.New("failed during delete item from collection service")
	}

	if deleteAppRequest.ResponseStatusCode == http.StatusNotFound {
		sessionState.LogEntry.LogError(errors.Errorf("** continue execution for client compliance ** failed to delete app from engine: DELETE %s returns 404", deleteAppEngineEndpoint))
	}

	sessionState.ArtifactMap.DeleteApp(deleteGuid)
	return nil
}
