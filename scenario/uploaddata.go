package scenario

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/tempcontent"
)

type (
	// UploadDataSettingsCore core parameters used in UnMarshalJSON interface
	UploadDataSettingsCore struct {
		Filename   string               `json:"filename" displayname:"Filename" displayelement:"file" doc-key:"uploaddata.filename"`
		SpaceID    string               `json:"spaceid" displayname:"Space ID" doc-key:"uploaddata.spaceid"`
		Replace    bool                 `json:"replace" displayname:"Replace file" doc-key:"uploaddata.replace"`
		TimeOut    helpers.TimeDuration `json:"timeout" displayname:"Timeout upload after this duration" doc-key:"tus.timeout"`
		ChunkSize  int64                `json:"chunksize" displayname:"Chunk size (bytes)" doc-key:"tus.chunksize"`
		MaxRetries int                  `json:"retries" displayname:"Number of retries on failed chunk upload" doc-key:"tus.retries"`
	}

	// UploadDataSettings specify data file to upload
	UploadDataSettings struct {
		UploadDataSettingsCore
	}
)

const (
	datafileEndpoint = "api/v1/qix-datafiles"
)

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

	tempFileClient, err := tempcontent.NewTUSClient(sessionState, connection, settings.ChunkSize, settings.MaxRetries)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	uploadCtx := sessionState.BaseContext()
	if settings.TimeOut > 0 {
		ctx, cancel := context.WithTimeout(uploadCtx, time.Duration(settings.TimeOut))
		uploadCtx = ctx
		defer cancel()
	}
	file, err := os.Open(settings.Filename)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to open file <%s>", settings.Filename))
		return
	}
	defer file.Close()
	tempFile, err := tempFileClient.UploadFromFile(uploadCtx, file)
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
