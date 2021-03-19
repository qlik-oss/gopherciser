package elastic

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// DeleteDataSettingsCore settings core used by UnmarshalJSON
	DeleteDataSettingsCore struct {
		Filename string `json:"filename" displayname:"Filename" doc-key:"deletedata.filename"`
		SpaceID  string `json:"spaceid" displayname:"Space ID" doc-key:"deletedata.spaceid"`
	}

	// DeleteDataSettings specify data file to delete
	DeleteDataSettings struct {
		DeleteDataSettingsCore
	}
)

// UnmarshalJSON unmarshals upload data settings from JSON
func (settings *DeleteDataSettings) UnmarshalJSON(arg []byte) error {
	// Check for deprecated fields
	if err := helpers.HasDeprecatedFields(arg, []string{
		"/path",
	}); err != nil {
		return errors.Errorf("%s %s, please remove from script", ActionDeleteData, err.Error())
	}
	var deleteDataSettings DeleteDataSettingsCore
	if err := jsonit.Unmarshal(arg, &deleteDataSettings); err != nil {
		return errors.Wrapf(err, "failed to unmarshal action<%s>", ActionDeleteData)
	}
	*settings = DeleteDataSettings{deleteDataSettings}

	return nil
}

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

	n := 0
	for _, file := range dataFiles {
		if file.Name == settings.Filename {
			deleteID := file.ID

			// Construct the Delete request with the database ID
			deleteItem := session.RestRequest{
				Method:      session.DELETE,
				ContentType: "application/json",
				Destination: fmt.Sprintf("%v/%v/%v", host, datafileEndpoint, deleteID),
			}

			sessionState.Rest.QueueRequest(actionState, true, &deleteItem, sessionState.LogEntry)
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

	if _, err := sessionState.FetchQixDataFiles(actionState, host, dataConnectionID); err != nil {
		actionState.AddErrors(errors.WithStack(err))
	}

	sessionState.Wait(actionState)
}
