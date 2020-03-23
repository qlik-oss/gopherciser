package scenario

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ElasticExportAppSettingsCore Currently used ElasticExportAppSettings (as opposed to deprecated settings)
	ElasticExportAppSettingsCore struct {
		NoData     bool                   `json:"nodata" displayname:"Export without data" doc-key:"elasticexportapp.nodata"`
		FileName   session.SyncedTemplate `json:"exportname" displayname:"Export filename" displayelement:"savefile" doc-key:"elasticexportapp.filename"`
		SaveToFile bool                   `json:"savetofile" displayname:"Save to file" doc-key:"elasticexportapp.savetofile"`
	}
	// ElasticExportAppSettings
	ElasticExportAppSettings struct {
		session.AppSelection
		ElasticExportAppSettingsCore
	}

	deprecatedElasticExportAppSettings struct {
		Title   string `json:"title"`
		AppGUID string `json:"appguid"`
	}
)

// UnmarshalJSON unmarshals export app settings from JSON
func (settings *ElasticExportAppSettings) UnmarshalJSON(arg []byte) error {
	var deprecated deprecatedElasticExportAppSettings
	if err := jsonit.Unmarshal(arg, &deprecated); err == nil { // skip check if error
		hasSettings := make([]string, 0, 2)
		if deprecated.AppGUID != "" {
			hasSettings = append(hasSettings, "appguid")
		}
		if deprecated.Title != "" {
			hasSettings = append(hasSettings, "title")
		}
		if len(hasSettings) > 0 {
			return errors.Errorf("%s settings<%s> are no longer used", ActionElasticExportApp, strings.Join(hasSettings, ","))
		}
	}
	var core ElasticExportAppSettingsCore
	if err := jsonit.Unmarshal(arg, &core); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticExportApp)
	}
	var appSelectCore session.AppSelection
	if err := jsonit.Unmarshal(arg, &appSelectCore); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionElasticExportApp)
	}

	(*settings).ElasticExportAppSettingsCore = core
	(*settings).AppSelection = appSelectCore
	return nil
}

// Validate ElasticExportApp
func (settings ElasticExportAppSettings) Validate() error {
	if err := settings.AppSelection.Validate(); err != nil {
		return err
	}
	return nil
}

// Execute ElasticExportApp
func (settings ElasticExportAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {

	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	entry, err := settings.AppSelection.Select(sessionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "Failed to perform app selection"))
		return
	}

	postExport := session.RestRequest{
		Method:      session.POST,
		ContentType: "",
		Destination: fmt.Sprintf("%s/api/v1/apps/%s/export?NoData=%v", host, entry.GUID, settings.NoData),
	}
	sessionState.Rest.QueueRequest(actionState, true, &postExport, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // error occurred
	}

	if postExport.ResponseStatusCode != http.StatusCreated {
		actionState.AddErrors(errors.Errorf("Unexpected response on app export request: %d", postExport.ResponseStatusCode))
		return
	}

	if len(postExport.ResponseHeaders) < 1 {
		actionState.AddErrors(errors.New("response has no headers"))
		return
	}

	location := postExport.ResponseHeaders["Location"]
	if len(location) < 1 {
		actionState.AddErrors(errors.New("no location header found in response"))
		return
	}

	tempContent := location[0]
	if tempContent == "" {
		actionState.AddErrors(errors.New("no temp-content path in header"))
		return
	}

	downloadReq := sessionState.Rest.FireOffGet(fmt.Sprintf("%s%s", host, tempContent), actionState, false)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if err := session.CheckResponseStatus(downloadReq, []int{200}); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to download app"))
		return
	}

	if settings.SaveToFile {
		filename := entry.Title

		if settings.FileName.String() != "" {
			data := struct {
				Title string
			}{Title: strings.TrimSuffix(entry.Title, ".qvf")}

			filename, err = sessionState.ReplaceSessionVariablesWithLocalData(&settings.FileName, data)
			if err != nil {
				actionState.AddErrors(errors.WithStack(err))
				return
			}
		}

		if !strings.HasSuffix(filename, ".qvf") {
			filename += ".qvf"
		}

		fileFullPath := path.Join(sessionState.OutputsDir, filename)

		errWrite := ioutil.WriteFile(fileFullPath, downloadReq.ResponseBody, 0644)
		if errWrite != nil {
			actionState.AddErrors(errors.Wrap(errWrite, "failed writing exported app to file"))
			return
		}
	}

	sessionState.Wait(actionState)
}
