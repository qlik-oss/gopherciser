package scenario

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"

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

	if existingFile != nil && !settings.Replace {
		sessionState.LogEntry.Logf(
			logger.WarningLevel, "data file not uploaded, filename<%s> already exists and replace set to false", fileName,
		)
		sessionState.Wait(actionState)
		return
	}

	fileContent, err := ioutil.ReadFile(settings.Filename)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to open file <%s>", settings.Filename))
		return
	}
	fileSize := len(fileContent)

	tempContentPostRequest := sessionState.Rest.PostWithHeadersAsync(
		fmt.Sprintf("%s/%s", host, tempContentFilesEndpoint),
		actionState, sessionState.LogEntry, nil, map[string]string{
			"upload-length": strconv.Itoa(fileSize),
			"tus-resumable": "1.0.0",
		}, &session.ReqOptions{
			ExpectedStatusCode: []int{200, 201},
			FailOnError:        true,
		},
	)

	if sessionState.Wait(actionState) {
		return
	}

	tempLocation := tempContentPostRequest.ResponseHeaders.Get("location")
	if tempLocation == "" {
		actionState.AddErrors(errors.New("temp-content did not return a storage location"))
		return
	}

	tempLocationURL, err := url.Parse(tempLocation)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to parse temp content location"))
		return
	}

	tempContentFileID := path.Base(tempLocationURL.Path)

	_ = sessionState.Rest.PatchWithHeadersAsync(
		fmt.Sprintf("%s/%s/%s", host, tempContentFilesEndpoint, tempContentFileID),
		actionState, sessionState.LogEntry, fileContent,
		map[string]string{
			"tus-resumable": "1.0.0",
			"upload-offset": "0",
		},
		&session.ReqOptions{
			ExpectedStatusCode: []int{200, 204},
			FailOnError:        true,
			ContentType:        "application/offset+octet-stream",
		},
	)

	reqParams := fmt.Sprintf("connectionId=%s&name=%s&tempContentFileId=%s", dataConnectionID, fileName, tempContentFileID)

	if sessionState.Wait(actionState) {
		return
	}

	dataFilesPostRequest := sessionState.Rest.PostAsync(
		fmt.Sprintf("%s/%s?%s", host, datafileEndpoint, reqParams), actionState, sessionState.LogEntry, nil, &session.ReqOptions{
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
