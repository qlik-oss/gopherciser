package elastic

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/tempcontent"
)

type (
	UploadMode int

	// ElasticUploadAppSettingsCore settings used in unmarshal interface
	ElasticUploadAppSettingsCore struct {
		ChunkSize  int64                `json:"chunksize" displayname:"Chunk size (bytes)" doc-key:"tus.chunksize"`
		MaxRetries int                  `json:"retries" displayname:"Number of retries on failed chunk upload" doc-key:"tus.retries"`
		TimeOut    helpers.TimeDuration `json:"timeout" displayname:"Timeout upload after this duration" doc-key:"tus.timeout"`
		Mode       UploadMode           `json:"mode" displayname:"Upload mode" doc-key:"elasticuploadapp.mode"`
		Filename   string               `json:"filename" displayname:"Filename" displayelement:"file" doc-key:"elasticuploadapp.filename"`
	}

	// ElasticUploadAppSettings specify app to upload
	ElasticUploadAppSettings struct {
		ElasticUploadAppSettingsCore
		DestinationSpace
		CanAddToCollection
	}

	// CanAddToCollection common collection settings
	CanAddToCollection struct {
		Title      session.SyncedTemplate `json:"title" displayname:"Title" doc-key:"elasticuploadapp.title"`
		Stream     session.SyncedTemplate `json:"stream" displayname:"Stream name" doc-key:"elasticuploadapp.stream"`
		StreamGUID string                 `json:"streamguid" displayname:"Stream ID" doc-key:"elasticuploadapp.streamguid"`
	}
)

const (
	// Tus chunked upload using the tus protocol
	Tus UploadMode = iota
	// Legacy upload using a POST payload
	Legacy
)

// GetEnumMap for upload mode
func (value UploadMode) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"tus":    int(Tus),
		"legacy": int(Legacy),
	})
	return enumMap
}

// UnmarshalJSON unmarshal DistributionType
func (value *UploadMode) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal UploadMode")
	}

	*value = UploadMode(i)
	return nil
}

// MarshalJSON marshal ThinkTime type
func (value UploadMode) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown UploadMode<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// UnmarshalJSON unmarshals ElasticUploadAppSettings
func (settings *ElasticUploadAppSettings) UnmarshalJSON(arg []byte) error {
	// Check for deprecated fields
	if err := helpers.HasDeprecatedFields(arg, []string{
		"/spaceid",
	}); err != nil {
		return errors.Errorf("%s %s, please remove from script", ActionElasticUploadApp, err.Error())
	}

	if err := jsonit.Unmarshal(arg, &settings.ElasticUploadAppSettingsCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticUploadApp)
	}

	if err := jsonit.Unmarshal(arg, &settings.DestinationSpace); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticUploadApp)
	}

	if err := jsonit.Unmarshal(arg, &settings.CanAddToCollection); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticUploadApp)
	}

	return nil
}

// Validate action (Implements ActionSettings interface)
func (settings ElasticUploadAppSettings) Validate() error {
	if _, err := os.Stat(settings.Filename); os.IsNotExist(err) {
		return errors.Errorf("File <%v> not found", settings.Filename)
	}
	if settings.Title.String() == "" {
		return errors.New("No Title specified")
	}
	if settings.ChunkSize < 0 {
		return errors.New("ChunkSize must be a positive value")
	}
	return nil
}

// Execute action (Implements ActionSettings interface)
func (settings ElasticUploadAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	restUrl, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	destSpace, err := settings.ResolveDestinationSpace(sessionState, actionState, restUrl)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	file, err := os.Open(settings.Filename)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to open file <%s>", settings.Filename))
		return
	}
	defer func() {
		_ = file.Close()
	}()

	var postApp session.RestRequest
	switch settings.Mode {
	case Tus:
		uploadCtx := sessionState.BaseContext()
		if settings.TimeOut > 0 {
			ctx, cancel := context.WithTimeout(uploadCtx, time.Duration(settings.TimeOut))
			uploadCtx = ctx
			defer cancel()
		}
		tempFile, err := tempcontent.UploadTempContentFromFile(uploadCtx, sessionState,
			connection, file, settings.ChunkSize, settings.MaxRetries)
		if err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}

		query := url.Values{}
		query.Add("fallbackname", filepath.Base(settings.Filename))

		parameters := ""
		if destSpace != nil {
			parameters = fmt.Sprintf("&spaceId=%v", destSpace.ID)
		}
		postApp = session.RestRequest{
			Method:      session.POST,
			ContentType: "application/json",
			Destination: fmt.Sprintf("%v/api/v1/apps/import?fileId=%v&%v%v", restUrl, tempFile.ID, query.Encode(), parameters),
		}
	case Legacy:
		parameters := ""
		if destSpace != nil {
			parameters = fmt.Sprintf("?spaceId=%v", destSpace.ID)
		}
		postApp = session.RestRequest{
			Method:        session.POST,
			ContentType:   "application/octet-stream",
			Destination:   fmt.Sprintf("%v/api/v1/apps/import%v", restUrl, parameters),
			ContentReader: file,
		}
	default:
		actionState.AddErrors(errors.Errorf("unknown upload format <%v>", settings.Mode))
		return
	}

	sessionState.Rest.QueueRequest(actionState, true, &postApp, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if postApp.ResponseStatusCode != http.StatusOK {
		actionState.AddErrors(errors.Errorf("Failed to upload app payload: %d <%s>", postApp.ResponseStatusCode, postApp.ResponseBody))
		return
	}

	appImportResponseRaw := postApp.ResponseBody
	var appImportResponse elasticstructs.AppImportResponse
	if err := jsonit.Unmarshal(appImportResponseRaw, &appImportResponse); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed unmarshaling app import response data: %s", appImportResponseRaw))
		return
	}

	err = AddAppToCollection(settings.CanAddToCollection, sessionState, actionState, appImportResponse, restUrl)
	if err != nil {
		actionState.AddErrors(err)
	}
}

func AddAppToCollection(settings CanAddToCollection, sessionState *session.State, actionState *action.State, appImportResponse elasticstructs.AppImportResponse, host string) error {
	var streamID string
	if settings.StreamGUID == "" && settings.Stream.String() != "" {
		stream, err := sessionState.ReplaceSessionVariables(&settings.Stream)
		if err != nil {
			return errors.WithStack(err)
		}
		streamID, err = searchForTag(sessionState, actionState, host, stream, 20)
		if err != nil {
			return errors.Wrapf(err, "stream not found")
		}
	} else {
		streamID = settings.StreamGUID
	}

	title, err := sessionState.ReplaceSessionVariables(&settings.Title)
	if err != nil {
		return errors.WithStack(err)
	}
	collectionServiceItemResponse, err := AddItemToCollectionService(sessionState, actionState, appImportResponse, title, host)
	if err != nil {
		return err
	}

	itemID, ok := collectionServiceItemResponse["id"].(string)
	if !ok {
		return errors.New("failed to get id from collection service response")
	}
	appGUID, ok := collectionServiceItemResponse["resourceId"].(string)
	if !ok {
		return errors.New("failed to get resource id from collection service response")
	}

	// No collection to add it to; we're done
	if streamID != "" {
		err := AddTag(sessionState, actionState, itemID, host, streamID)
		if err != nil {
			return err
		}
	}
	err = sessionState.ArtifactMap.FillArtifacts(&session.ItemData{
		Data: []session.ArtifactEntry{
			{
				Name:         title,
				ID:           appGUID,
				ItemID:       itemID,
				ResourceType: session.ResourceTypeApp,
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed adding new app to internal artifact map")
	}

	// Set "current" app
	sessionState.CurrentApp = &session.ArtifactEntry{Name: title, ID: appGUID, ItemID: itemID, ResourceType: session.ResourceTypeApp}

	// Debug log of artifact map in it's entirety after uploading app
	if err := sessionState.ArtifactMap.LogMap(sessionState.LogEntry); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, err)
	}

	return nil
}

func AddItemToCollectionService(sessionState *session.State, actionState *action.State, appImportResponse elasticstructs.AppImportResponse, title string, host string) (map[string]interface{}, error) {
	collectionServiceItem := elasticstructs.CollectionServiceItem{
		Name:         title,
		ResourceID:   appImportResponse.Attributes.ID,
		ResourceType: appImportResponse.Attributes.ResourceType,
		Description:  appImportResponse.Attributes.Description,
		ResourceAttributes: elasticstructs.CollectionServiceResourceAttributes{
			ID:               appImportResponse.Attributes.ID,
			Name:             title,
			Description:      appImportResponse.Attributes.Description,
			Thumbnail:        appImportResponse.Attributes.Thumbnail,
			LastReloadTime:   appImportResponse.Attributes.LastReloadTime,
			CreatedDate:      appImportResponse.Attributes.CreatedDate,
			ModifiedDate:     appImportResponse.Attributes.ModifiedDate,
			OwnerID:          appImportResponse.Attributes.OwnerID,
			DynamicColor:     appImportResponse.Attributes.DynamicColor,
			Published:        appImportResponse.Attributes.Published,
			PublishTime:      appImportResponse.Attributes.PublishTime,
			HasSectionAccess: appImportResponse.Attributes.HasSectionAccess,
			Encrypted:        appImportResponse.Attributes.Encrypted,
			OriginAppID:      appImportResponse.Attributes.OriginAppID,
			SpaceID:          appImportResponse.Attributes.SpaceID,
			ResourceType:     appImportResponse.Attributes.ResourceType,
		},
		ResourceCustomAttributes: appImportResponse.Attributes.Custom,
		ResourceCreatedAt:        time.Now(),
		ResourceCreatedBySubject: appImportResponse.Attributes.Owner,
		SpaceID:                  appImportResponse.Attributes.SpaceID,
	}

	createItemInCollectionServiceContent, err := jsonit.Marshal(collectionServiceItem)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare payload to collection service")
	}

	createItemInCollectionService := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%v/api/v1/items", host),
		Content:     createItemInCollectionServiceContent,
	}

	sessionState.Rest.QueueRequest(actionState, true, &createItemInCollectionService, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return nil, errors.New("failed during create item in collection service")
	}
	if createItemInCollectionService.ResponseStatusCode != http.StatusCreated {
		return nil, errors.Errorf("failed to create item in collection service: %s", createItemInCollectionService.ResponseBody)
	}

	collectionServiceItemResponseRaw := createItemInCollectionService.ResponseBody
	var collectionServiceItemResponse map[string]interface{}
	if err := jsonit.Unmarshal(collectionServiceItemResponseRaw, &collectionServiceItemResponse); err != nil {
		return nil, errors.Wrapf(err, "failed unmarshaling collection service item creation response data: %s", collectionServiceItemResponseRaw)
	}
	return collectionServiceItemResponse, nil
}

func AddTag(sessionState *session.State, actionState *action.State, itemId string, host string, collectionId string) error {
	itemCollectionAdd := elasticstructs.ItemCollectionAdd{}
	itemCollectionAdd.ID = itemId
	itemCollectionAddContent, err := jsonit.Marshal(itemCollectionAdd)
	if err != nil {
		return errors.Wrap(err, "failed to prepare payload to publish")
	}

	collectAdd := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%v/api/v1/collections/%v/items", host, collectionId),
		Content:     itemCollectionAddContent,
	}

	sessionState.Rest.QueueRequest(actionState, true, &collectAdd, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return errors.New("failed during add collection")
	}

	if collectAdd.ResponseStatusCode != http.StatusCreated {
		return errors.Errorf("failed to publish app to stream: %s", collectAdd.ResponseBody)
	}
	return nil
}
