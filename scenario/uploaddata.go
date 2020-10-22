package scenario

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

// TODO update to new form-data
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

const datafileEndpoint = "api/v1/qix-datafiles"
const refererPath = "%s/explore/spaces/%s/data"

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

	restHandler := sessionState.Rest

	dataConnectionID, err := sessionState.FetchDataConnectionID(actionState, host, settings.SpaceID)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	dataFiles, err := sessionState.FetchQixDataFiles(actionState, host, dataConnectionID)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	fileName := filepath.Base(settings.Filename)

	var existingFile *elasticstructs.QixDataFile
	// check to see if file exists already
	for _, file := range dataFiles {
		if file.Name == fileName {
			existingFile = &file
			break
		}
	}

	if existingFile != nil && !settings.Replace {
		sessionState.LogEntry.Logf(
			logger.WarningLevel, "data file not uploaded, filename<%s> already exists and replace set to false", settings.Filename,
		)
		sessionState.Wait(actionState)
		return
	}

	file, err := os.Open(settings.Filename)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to open file <%s>", settings.Filename))
		return
	}
	defer func() {
		err = file.Close()
		if err != nil {
			actionState.AddErrors(errors.Wrapf(err, "failed to close file <%s>", settings.Filename))
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Then create the binary multipart field
	part, err := writer.CreateFormFile("data", filepath.Base(settings.Filename))
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to create multipart form"))
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to copy file contents to part"))
		return
	}

	err = writer.Close()
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to close multipart writer"))
		return
	}

	sessionState.Rest.GetAsync(
		fmt.Sprintf("%s/%s/quota", host, datafileEndpoint), actionState, sessionState.LogEntry, nil,
	)

	postData := session.RestRequest{
		Method:        session.POST,
		ContentType:   writer.FormDataContentType(),
		Destination:   fmt.Sprintf("%s/%s?connectionId=%s&name=%s", host, datafileEndpoint, dataConnectionID, fileName),
		ContentReader: body,
	}

	if existingFile != nil {
		postData.Method = session.PUT
		postData.Destination = fmt.Sprintf("%s/%s/%s?connectionId=%s&name=%s", host, datafileEndpoint, existingFile.ID, dataConnectionID, fileName)
	}

	// Set referer to personal or space ID for space we are uploading to
	referer := "personal"
	if settings.SpaceID != "" {
		referer = settings.SpaceID
	}
	postData.ExtraHeaders = map[string]string{"Referer": fmt.Sprintf(refererPath, host, referer)}

	restHandler.QueueRequest(actionState, true, &postData, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if postData.ResponseStatusCode == http.StatusConflict {
		sessionState.LogEntry.Logf(
			logger.WarningLevel, "cannot upload data file: filename conflict <%s>", settings.Filename,
		)
		return
	}
	if postData.ResponseStatusCode != http.StatusCreated {
		actionState.AddErrors(
			errors.New(
				fmt.Sprintf(
					"failed to upload data file payload: %d <%s>", postData.ResponseStatusCode, postData.ResponseBody,
				),
			),
		)
		return
	}

	sessionState.Rest.GetAsync(
		fmt.Sprintf("%s/%s/quota", host, datafileEndpoint), actionState, sessionState.LogEntry, nil,
	)

	if _, err := sessionState.FetchQixDataFiles(actionState, host, dataConnectionID); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		// no return here, wait for async quota too
	}

	sessionState.Wait(actionState)
}
