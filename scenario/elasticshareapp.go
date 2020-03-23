package scenario

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// ElasticShareAppSettings
	ElasticShareAppSettings struct {
		Title   session.SyncedTemplate `json:"title" displayname:"App title" doc-key:"elasticshareapp.title"`
		AppGUID string                 `json:"appguid" displayname:"App GUID" doc-key:"elasticshareapp.appguid"`
		Groups  []string               `json:"groups" displayname:"Groups" doc-key:"elasticshareapp.groups"`
	}
)

// Validate ElasticShareAppSettings
func (settings ElasticShareAppSettings) Validate() error {
	if settings.Title.String() == "" {
		return errors.Errorf("%s: no app title defined", ActionElasticUploadApp)
	}
	if len(settings.Groups) < 1 {
		return errors.Errorf("%s: no groups for sharing defined", ActionElasticUploadApp)
	}
	return nil
}

// Execute ElasticShareAppSettings
func (settings ElasticShareAppSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	if len(settings.Groups) < 1 {
		actionState.AddErrors(errors.Errorf("No groups defined for %s", ActionElasticUploadApp))
		return
	}

	title, err := sessionState.ReplaceSessionVariables(&settings.Title)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	appId := settings.AppGUID
	if appId == "" {
		appId, err = sessionState.ArtifactMap.GetAppID(title)
		if err != nil {
			actionState.AddErrors(errors.Wrapf(err, "failed to find app<%s>", title))
			return
		}
	}

	itemId, err := sessionState.ArtifactMap.GetItemId(title)
	if err != nil || itemId == "" {
		actionState.AddErrors(errors.Wrapf(err, "failed to get item id for app<%s>", title))
		return
	}

	shareAppsEndpointPayload := elasticstructs.ShareAppsEndpointPayload{}
	shareAppsEndpointPayload.Attributes.Custom.Groupswithaccess = settings.Groups
	shareAppsEndpointPayload.Attributes.Custom.UserIdsWithAccess = []string{}

	shareAppsEndpointPayloadJson, err := jsonit.Marshal(shareAppsEndpointPayload)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to prepare payload to share (api/v1/items)"))
		return
	}

	shareApps := session.RestRequest{
		Method:      session.PUT,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%s/api/v1/apps/%s", host, appId),
		Content:     shareAppsEndpointPayloadJson,
	}
	sessionState.Rest.QueueRequest(actionState, true, &shareApps, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if shareApps.ResponseStatusCode != http.StatusOK {
		actionState.AddErrors(errors.New(fmt.Sprintf("failed to share app (/api/v1/apps): %d %s", shareApps.ResponseStatusCode, shareApps.ResponseBody)))
		return
	}

	shareItemsEndpointPayload := elasticstructs.ShareItemsEndpointPayload{}
	shareItemsEndpointPayload.ResourceCustomAttributes.Groupswithaccess = settings.Groups
	shareItemsEndpointPayload.ResourceCustomAttributes.UserIdsWithAccess = []string{}
	shareItemsEndpointPayload.ResourceID = appId
	shareItemsEndpointPayload.ResourceType = "app"
	shareItemsEndpointPayload.Name = title

	shareItemsEndpointPayloadJson, err := jsonit.Marshal(shareItemsEndpointPayload)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to prepare payload to share (api/v1/items)"))
		return
	}

	shareItems := session.RestRequest{
		Method:      session.PUT,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%s/api/v1/items/%s", host, itemId),
		Content:     shareItemsEndpointPayloadJson,
	}
	sessionState.Rest.QueueRequest(actionState, true, &shareItems, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if shareItems.ResponseStatusCode != http.StatusOK {
		actionState.AddErrors(errors.New(fmt.Sprintf("failed to share app (/api/v1/items): %d %s", shareItems.ResponseStatusCode, shareItems.ResponseBody)))
		return
	}

	sessionState.Wait(actionState)
}
