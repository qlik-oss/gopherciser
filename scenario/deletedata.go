package scenario

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// DeleteDataSettings specify data file to delete
	DeleteDataSettings struct {
		Filename string `json:"filename" displayname:"Filename" doc-key:"deletedata.filename"`
		Path     string `json:"path" displayname:"Path" doc-key:"deletedata.path"`
	}
)

const dataListEndpoint = "api/v1/qix-datafiles"

// Validate action (Implements ActionSettings interface)
func (settings DeleteDataSettings) Validate() error {
	if settings.Filename == "" {
		return errors.New("no filename specified")
	}
	return nil
}

// Execute action (Implements ActionSettings interface)
func (settings DeleteDataSettings) Execute(
	sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string,
	reset func(),
) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	if settings.Path == "" {
		settings.Path = defaultDataPath
	}

	restHandler := sessionState.Rest

	// Look up the database ID for the file GUID
	getItems := session.RestRequest{
		Method:      session.GET,
		Destination: fmt.Sprintf("%s/%s?path=%s", host, dataListEndpoint, settings.Path),
	}
	restHandler.QueueRequest(actionState, false, &getItems, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}

	var folder *elasticstructs.GetDataFolders

	if err := jsonit.Unmarshal(getItems.ResponseBody, &folder); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to unmarshal getdatafolders"))
		return
	}

	n := 0
	for _, file := range *folder {
		if file.Name == settings.Filename {
			deleteId := file.ID

			// Construct the Delete request with the database ID
			deleteItem := session.RestRequest{
				Method:      session.DELETE,
				ContentType: "application/json",
				Destination: fmt.Sprintf("%v/%v/%v", host, datafileEndpoint, deleteId),
			}

			restHandler.QueueRequest(actionState, true, &deleteItem, sessionState.LogEntry)
			if sessionState.Wait(actionState) {
				return // we had an error
			}

			if deleteItem.ResponseStatusCode != http.StatusNoContent {
				actionState.AddErrors(
					errors.Errorf(
						"failed to delete data file: %d %s", deleteItem.ResponseStatusCode, deleteItem.ResponseBody,
					),
				)
			} else {
				n++
			}
		}
	}

	sessionState.LogEntry.LogInfo("NumDeletedFiles", fmt.Sprintf("%d", n))
	if n == 0 {
		sessionState.LogEntry.Logf(logger.WarningLevel, "no files deleted")
	}

	sessionState.Rest.GetAsync(
		fmt.Sprintf("%s/%s/quota", host, datafileEndpoint), actionState, sessionState.LogEntry, nil,
	)
	sessionState.Rest.GetAsync(
		fmt.Sprintf(
			"%s/%s?connectionId=%s&top=1000", host, datafileEndpoint, sessionState.DataConnectionId,
		), actionState, sessionState.LogEntry, nil,
	)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
}
