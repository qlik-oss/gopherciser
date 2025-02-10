package scenario

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/structs"
	"github.com/qlik-oss/gopherciser/synced"
)

type (
	//DeleteOdagSettings settings for DeleteOdag
	DeleteOdagSettings struct {
		Name synced.Template `json:"linkname" displayname:"ODAG link name" doc-key:"deleteodag.linkname"`
	}
)

// Validate DeleteOdagSettings action (Implements ActionSettings interface)
func (settings DeleteOdagSettings) Validate() ([]string, error) {
	if settings.Name.String() == "" {
		return nil, errors.New("no ODAG link name specified")
	}
	return nil, nil
}

// Execute DeleteOdagSettings action (Implements ActionSettings interface)
func (settings DeleteOdagSettings) Execute(sessionState *session.State, actionState *action.State,
	connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	odagEndpoint := WindowsOdagEndpointConfiguration
	err := DeleteOdag(sessionState, settings, actionState, connectionSettings, odagEndpoint, "")
	if err != nil {
		actionState.AddErrors(err)
	}
}

// DeleteOdag delete ODAG app
func DeleteOdag(sessionState *session.State, settings DeleteOdagSettings, actionState *action.State, connectionSettings *connection.ConnectionSettings, odagEndpoint OdagEndpointConfiguration, selectionAppId string) error {
	odagLinkName, err := sessionState.ReplaceSessionVariables(&settings.Name)
	if err != nil {
		return err
	}
	host, err := connectionSettings.GetRestUrl()
	if err != nil {
		return err
	}

	// first, find the ID of the ODAG link we want
	odagLink, err := getOdagLinkByName(odagLinkName, host, sessionState, actionState, odagEndpoint.Main, selectionAppId)
	if err != nil {
		return errors.Wrap(err, "failed to find ODAG link")
	}

	// find the IDs of each request from this odag link ID
	noErrOptions := session.DefaultReqOptions()
	noErrOptions.FailOnError = false
	odagRequests, err := sessionState.Rest.GetSync(fmt.Sprintf("%s/%s/%s/requests", host, odagEndpoint.Main, odagLink.ID), actionState, sessionState.LogEntry, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get ODAG links: %s", odagRequests.ResponseBody))
	}

	var odagRequestsByLink structs.OdagRequestsByLink
	if err := json.Unmarshal(odagRequests.ResponseBody, &odagRequestsByLink); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed unmarshaling ODAG requests GET reponse: %s", odagRequests.ResponseBody))
	}

	deletedApps := make([]string, 0, len(odagRequestsByLink))
	failedDeletedApps := make([]string, 0, len(odagRequestsByLink))

	// delete each request by its ID
	for _, odagRequest := range odagRequestsByLink {
		if odagRequest.GeneratedApp.ID == "" {
			continue // the generated app does not exist
		}
		idToDelete := odagRequest.ID

		deleteOptions := session.DefaultReqOptions()
		deleteOptions.ExpectedStatusCode = []int{http.StatusNoContent}
		deleteOdagRequest, err := sessionState.Rest.DeleteSync(fmt.Sprintf("%s/%s/%s/app", host, odagEndpoint.Requests, idToDelete), actionState, sessionState.LogEntry, deleteOptions)
		if err != nil {
			actionState.AddErrors(errors.Errorf("unexpected response code <%d> from delete request: %s", deleteOdagRequest.ResponseStatusCode, deleteOdagRequest.ResponseBody))
			failedDeletedApps = append(failedDeletedApps, idToDelete)
		} else {
			deletedApps = append(deletedApps, idToDelete)
		}
	}

	sessionState.LogEntry.LogInfo("NumDeletedApps", fmt.Sprintf("%d", len(deletedApps)))
	sessionState.LogEntry.LogInfo("DeletedApps", strings.Join(deletedApps, ","))
	if len(failedDeletedApps) > 0 {
		sessionState.LogEntry.LogInfo("NumFailedDeletedApps", fmt.Sprintf("%d", len(failedDeletedApps)))
		sessionState.LogEntry.LogInfo("FailedDeletedApps", strings.Join(failedDeletedApps, ","))
	}
	if len(deletedApps) == 0 {
		sessionState.LogEntry.Logf(logger.WarningLevel, "no apps deleted from ODAG link")
	}

	return nil
}
