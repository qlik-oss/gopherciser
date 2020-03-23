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
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// UploadDataSettings specify data file to upload
	UploadDataSettings struct {
		Filename string `json:"filename" displayname:"Filename" displayelement:"file" doc-key:"uploaddata.filename"`
		Path     string `json:"destinationpath" displayname:"Destination path" doc-key:"uploaddata.destinationpath"`
	}
)

const datafileEndpoint = "api/v1/qix-datafiles"
const refererPath = "%s/sense/app/%s/datamanager/datamanager"
const defaultDataPath = "MyDataFiles"

// Validate action (Implements ActionSettings interface)
func (settings UploadDataSettings) Validate() error {
	if _, err := os.Stat(settings.Filename); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File <%v> not found", settings.Filename))
	}
	return nil
}

// Execute action (Implements ActionSettings interface)
func (settings UploadDataSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	if settings.Path == "" {
		settings.Path = defaultDataPath
	}

	restHandler := sessionState.Rest

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

	// First create the multipart field with the path name
	params := map[string]string{
		"path": settings.Path,
		"name": filepath.Base(settings.Filename),
	}
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

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

	postData := session.RestRequest{
		Method:        session.POST,
		ContentType:   writer.FormDataContentType(),
		Destination:   fmt.Sprintf("%s/%s", host, datafileEndpoint),
		ContentReader: body,
	}

	// If an app is open, use this as the referer field
	if sessionState.Connection != nil {
		senseConnection := sessionState.Connection.Sense()
		if senseConnection.CurrentApp != nil {
			currentAppGUID := senseConnection.CurrentApp.GUID
			appInfo := make(map[string]string)
			appInfo["Referer"] = fmt.Sprintf(refererPath, host, currentAppGUID)
			postData.ExtraHeaders = appInfo
		}
	}

	restHandler.QueueRequest(actionState, true, &postData, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if postData.ResponseStatusCode == http.StatusConflict {
		sessionState.LogEntry.Logf(logger.WarningLevel, "cannot upload data file: filename conflict <%s>", settings.Filename)
		return
	}
	if postData.ResponseStatusCode != http.StatusCreated {
		actionState.AddErrors(errors.New(fmt.Sprintf("failed to upload data file payload: %d <%s>", postData.ResponseStatusCode, postData.ResponseBody)))
		return
	}
}
