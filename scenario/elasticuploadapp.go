package scenario

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eventials/go-tus"
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
	UploadMode int

	// ElasticUploadAppSettings specify app to upload
	ElasticUploadAppSettings struct {
		ChunkSize  int64      `json:"chunksize" displayname:"Chunk size (bytes)" doc-key:"elasticuploadapp.chunksize"`
		MaxRetries int        `json:"retries" displayname:"Number of retries on failed chunk upload" doc-key:"elasticuploadapp.retries"`
		Mode       UploadMode `json:"mode" displayname:"Upload mode" doc-key:"elasticuploadapp.mode"`
		Filename   string     `json:"filename" displayname:"Filename" displayelement:"file" doc-key:"elasticuploadapp.filename"`
		SpaceID    string     `json:"spaceid" displayname:"Space ID" doc-key:"elasticuploadapp.spaceid"`
		CanAddToCollection
	}

	CanAddToCollection struct {
		Title      session.SyncedTemplate `json:"title" displayname:"Title" doc-key:"elasticuploadapp.title"`
		Stream     session.SyncedTemplate `json:"stream" displayname:"Stream name" doc-key:"elasticuploadapp.stream"`
		StreamGUID string                 `json:"streamguid" displayname:"Stream ID" doc-key:"elasticuploadapp.streamguid"`
		// Deprecated: This property will be removed.ElasticShareApp action should be used instead, keeping entry here for some time to make sure scripts get validation error.
		Groups []string `json:"groups" displayname:"Groups - deprecated" doc-key:"canaddtocollection.groups"`
	}
)

const (
	// Tus chunked upload using the tus protocol
	Tus UploadMode = iota
	// Legacy upload using a POST payload
	Legacy
)

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

// Current client default: 300 mb
const defaultChunkSize int64 = 300 * 1024 * 1024

// Validate action (Implements ActionSettings interface)
func (settings ElasticUploadAppSettings) Validate() error {
	if _, err := os.Stat(settings.Filename); os.IsNotExist(err) {
		return errors.Errorf("File <%v> not found", settings.Filename)
	}
	if settings.Title.String() == "" {
		return errors.New("No Title specified")
	}
	if len(settings.Groups) > 0 {
		return errors.New("elasticuploadapp action no longer utilizes Group, add a elasticshareapp action to share the app")
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
		host, err := connection.GetHost()
		if err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}

		httpClient, err := session.DefaultClient(connection, sessionState)
		if err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}

		// upload file using tus chunked uploads protocol
		chunkSize := defaultChunkSize
		if settings.ChunkSize > 0 {
			chunkSize = settings.ChunkSize
		}
		tusConfig := tus.DefaultConfig()
		tusConfig.ChunkSize = chunkSize
		tusConfig.Header = sessionState.HeaderJar.GetHeader(host)
		tusConfig.HttpClient = httpClient

		// upload to temporary storage
		client, err := tus.NewClient(fmt.Sprintf("%v/api/v1/temp-contents/files", restUrl), tusConfig)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to create tus client"))
			return
		}
		upload, err := tus.NewUploadFromFile(file)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to create tus upload from file"))
			return
		}
		uploader, err := client.CreateUpload(upload)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to create tus uploader"))
			return
		}

		err = nil
		retries := 0
		retryWithBackoff := func() bool {
			if retries < settings.MaxRetries {
				helpers.WaitFor(sessionState.BaseContext(), time.Second*time.Duration(retries))
				retries++
				return true
			} else {
				return false
			}
		}

		for err == nil || (err.Error() != "EOF" && retryWithBackoff()) {
			err = uploader.UploadChunck()
			if sessionState.IsAbortTriggered() {
				return
			}
		}

		if err.Error() != "EOF" {
			actionState.AddErrors(errors.Wrap(err, "failed to upload using tus"))
			return
		}

		// get url to the uploaded temporary file
		fileUrl := uploader.Url()
		fileUrlSplit := strings.Split(fileUrl, "/")
		if len(fileUrlSplit) < 1 {
			actionState.AddErrors(errors.Errorf("empty file url"))
			return
		}
		fileId := fileUrlSplit[len(fileUrlSplit)-1]

		parameters := ""
		if settings.SpaceID != "" {
			parameters = fmt.Sprintf("&spaceId=%v", settings.SpaceID)
		}
		postApp = session.RestRequest{
			Method:      session.POST,
			ContentType: "application/json",
			Destination: fmt.Sprintf("%v/api/v1/apps/import?fileId=%v%v", restUrl, fileId, parameters),
		}
	case Legacy:
		parameters := ""
		if settings.SpaceID != "" {
			parameters = fmt.Sprintf("?spaceId=%v", settings.SpaceID)
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
		streamID, err = searchForTag(sessionState, actionState, host, stream)
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
		return errors.Wrap(err, "failed to prepare payload to collection service")
	}

	createItemInCollectionService := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%v/api/v1/items", host),
		Content:     createItemInCollectionServiceContent,
	}

	sessionState.Rest.QueueRequest(actionState, true, &createItemInCollectionService, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return errors.New("failed during create item iń collection service")
	}
	if createItemInCollectionService.ResponseStatusCode != http.StatusCreated {
		return errors.Errorf("failed to create item in collection service: %s", createItemInCollectionService.ResponseBody)
	}

	collectionServiceItemResponseRaw := createItemInCollectionService.ResponseBody
	var collectionServiceItemResponse map[string]interface{}
	if err := jsonit.Unmarshal(collectionServiceItemResponseRaw, &collectionServiceItemResponse); err != nil {
		return errors.Wrapf(err, "failed unmarshaling collection service item creation response data: %s", collectionServiceItemResponseRaw)
	}

	itemId := collectionServiceItemResponse["id"].(string)
	appGuid := collectionServiceItemResponse["resourceId"].(string)

	// No collection to add it to; we're done
	if streamID != "" {
		itemCollectionAdd := elasticstructs.ItemCollectionAdd{}
		itemCollectionAdd.ID = itemId
		itemCollectionAddContent, err := jsonit.Marshal(itemCollectionAdd)
		if err != nil {
			return errors.Wrap(err, "failed to prepare payload to publish")
		}

		collectAdd := session.RestRequest{
			Method:      session.POST,
			ContentType: "application/json",
			Destination: fmt.Sprintf("%v/api/v1/collections/%v/items", host, streamID),
			Content:     itemCollectionAddContent,
		}

		sessionState.Rest.QueueRequest(actionState, true, &collectAdd, sessionState.LogEntry)
		if sessionState.Wait(actionState) {
			return errors.New("failed during add collection")
		}

		if collectAdd.ResponseStatusCode != http.StatusCreated {
			return errors.Errorf("failed to publish app to stream: %s", collectAdd.ResponseBody)
		}
	}
	err = sessionState.ArtifactMap.FillAppsUsingName(&session.AppData{
		Data: []session.AppsResp{
			{
				Name:   title,
				ID:     appGuid,
				ItemID: itemId,
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed adding new app to internal artifact map")
	}

	// Set "current" app
	sessionState.CurrentApp = &session.ArtifactEntry{Title: title, GUID: appGuid, ItemID: itemId}

	// Debug log of artifact map in it's entirety after uploading app
	if err := sessionState.ArtifactMap.LogMap(sessionState.LogEntry); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, err)
	}

	return nil
}
