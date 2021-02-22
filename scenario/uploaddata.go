package scenario

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/eventials/go-tus"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// UploadDataSettingsCore core parameters used in UnMarshalJSON interface
	UploadDataSettingsCore struct {
		Filename string `json:"filename" displayname:"Filename" displayelement:"file" doc-key:"uploaddata.filename"`
		SpaceID  string `json:"spaceid" displayname:"Space ID" doc-key:"uploaddata.spaceid"`
		Replace  bool   `json:"replace" displayname:"Replace file" doc-key:"uploaddata.replace"`
	}

	// UploadDataSettings specify data file to upload
	UploadDataSettings struct {
		UploadDataSettingsCore
	}
)

const (
	datafileEndpoint = "api/v1/qix-datafiles"
	// refererPath              = "%s/explore/spaces/%s/data"
	tempContentFilesEndpoint = "api/v1/temp-contents/files"
)

type tempFile struct {
	ID  string
	URL string
}

func uploadTempContentFromFile(sessionState *session.State, connection *connection.ConnectionSettings, file *os.File) (*tempFile, error) {
	const tempContentFilesEndpoint = "api/v1/temp-contents/files"
	restURL, err := connection.GetRestUrl()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	host, err := connection.GetHost()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// upload file using tus chunked uploads protocol
	chunkSize := defaultChunkSize
	tusConfig := tus.DefaultConfig()
	tusConfig.ChunkSize = chunkSize
	tusConfig.Header = sessionState.HeaderJar.GetHeader(host)
	tusConfig.HttpClient = sessionState.Rest.Client

	// upload to temporary storage
	client, err := tus.NewClient(fmt.Sprintf("%s/%s", restURL, tempContentFilesEndpoint), tusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tus client")
	}
	upload, err := tus.NewUploadFromFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tus upload from file")
	}
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tus uploader")
	}

	for err == nil {
		err = uploader.UploadChunck()
		// TODO pass context
		if sessionState.IsAbortTriggered() {
			return nil, errors.Wrap(err, "tus upload aborted")
		}
	}

	if err.Error() != "EOF" {
		return nil, errors.Wrap(err, "failed to upload using tus")
	}

	tempFile := &tempFile{}
	tempFile.URL = uploader.Url()
	tempLocationURL, err := url.Parse(tempFile.URL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse temp content location")
	}
	tempFile.ID = path.Base(tempLocationURL.Path)
	return tempFile, nil
}

// UnmarshalJSON unmarshals upload data settings from JSON
func (settings *UploadDataSettings) UnmarshalJSON(arg []byte) error {
	// Check for deprecated fields
	if err := helpers.HasDeprecatedFields(arg, []string{
		"/destinationpath",
	}); err != nil {
		return errors.Errorf("%s %s, please remove from script", ActionUploadData, err.Error())
	}
	var uploadDataSettings UploadDataSettingsCore
	if err := jsonit.Unmarshal(arg, &uploadDataSettings); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionUploadData)
	}
	*settings = UploadDataSettings{uploadDataSettings}

	return nil
}

// Validate action (Implements ActionSettings interface)
func (settings UploadDataSettings) Validate() error {
	if _, err := os.Stat(settings.Filename); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File <%v> not found", settings.Filename))
	}
	return nil
}

// Execute action (Implements ActionSettings interface)
func (settings UploadDataSettings) Execute(
	sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string,
	reset func(),
) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	sessionState.Rest.GetAsync(
		fmt.Sprintf("%s/%s/quota", host, datafileEndpoint), actionState, sessionState.LogEntry, nil,
	)

	dataConnectionID, err := sessionState.FetchDataConnectionID(actionState, host, settings.SpaceID)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	fileName := filepath.Base(settings.Filename)

	existingFile, err := sessionState.FetchQixDataFile(actionState, host, dataConnectionID, fileName)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
	sessionState.LogEntry.LogDebugf("existingFile %+v\n", existingFile)

	if existingFile != nil && !settings.Replace {
		sessionState.LogEntry.Logf(
			logger.WarningLevel, "data file not uploaded, filename<%s> already exists and replace set to false", fileName,
		)
		sessionState.Wait(actionState)
		return
	}

	file, err := os.Open(settings.Filename)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to open file <%s>", settings.Filename))
		return
	}

	tempFile, err := uploadTempContentFromFile(sessionState, connection, file)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to upload temp content from file"))
		return
	}

	reqURL := fmt.Sprintf("%s/%s", host, datafileEndpoint)
	httpMethodFunc := sessionState.Rest.PostAsync
	reqParams := fmt.Sprintf("connectionId=%s&name=%s&tempContentFileId=%s", dataConnectionID, fileName, tempFile.ID)
	if existingFile != nil {
		reqURL += "/" + existingFile.ID
		httpMethodFunc = sessionState.Rest.PutAsync
	}

	if sessionState.Wait(actionState) {
		return
	}

	dataFilesPostRequest := httpMethodFunc(
		fmt.Sprintf("%s?%s", reqURL, reqParams),
		actionState, sessionState.LogEntry, nil,
		&session.ReqOptions{
			ExpectedStatusCode: []int{200, 201},
			FailOnError:        true,
		},
	)
	if sessionState.Wait(actionState) {
		return
	}

	qixDataFile := elasticstructs.QixDataFile{}

	if err := json.Unmarshal(dataFilesPostRequest.ResponseBody, &qixDataFile); err != nil {
		actionState.AddErrors(err)
		return
	}

	sessionState.Wait(actionState)
}
