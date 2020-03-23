package scenario

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	ElasticDeleteCollectionSettings struct {
		CollectionName string `json:"name" displayname:"Collection name" doc-key:"elasticdeletecollection.name"`
		DeleteContents bool   `json:"deletecontents" displayname:"Delete collection contents" doc-key:"elasticdeletecollection.deletecontents"`
	}
)

const deleteCollectionEndpoint = "api/v1/collections"

// Validate ElasticDeleteCollectionSettings action (Implements ActionSettings interface)
func (settings ElasticDeleteCollectionSettings) Validate() error {
	if settings.CollectionName == "" {
		return errors.New("No name specified")
	}
	return nil
}

// Execute ElasticDeleteCollectionSettings action (Implements ActionSettings interface)
func (settings ElasticDeleteCollectionSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, setEfeDeleteCollectionStart func()) {
	collectionID, err := sessionState.ArtifactMap.GetStreamID(settings.CollectionName)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "no such collection <%s>", settings.CollectionName))
		return
	}

	if settings.DeleteContents {
		var deleteAppLabel string
		if label != "" {
			deleteAppLabel = fmt.Sprintf("%s - DA", label)
		}
		deleteApp := Action{
			ActionCore{
				Type:  ActionElasticDeleteApp,
				Label: deleteAppLabel,
			},
			&ElasticDeleteAppSettings{
				session.AppSelection{},
				ElasticDeleteAppCoreSettings{
					DeletionMode:   ClearCollection,
					CollectionName: settings.CollectionName,
				},
			},
		}
		if isAborted, err := CheckActionError(deleteApp.Execute(sessionState, connection)); isAborted {
			return // action is aborted, we should not continue
		} else if err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to clear collection"))
			return
		}
	}

	setEfeDeleteCollectionStart() // reset action start to this point

	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	restHandler := sessionState.Rest

	deleteCollection := session.RestRequest{
		Method:      session.DELETE,
		Destination: fmt.Sprintf("%v/%v/%v", host, deleteCollectionEndpoint, collectionID),
	}

	restHandler.QueueRequest(actionState, true, &deleteCollection, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if deleteCollection.ResponseStatusCode != http.StatusNoContent {
		actionState.AddErrors(errors.New(fmt.Sprintf("Failed to delete collection (%s): %s", deleteCollection.ResponseStatus, deleteCollection.ResponseBody)))
		return
	}

	sessionState.ArtifactMap.DeleteStream(settings.CollectionName)
}
