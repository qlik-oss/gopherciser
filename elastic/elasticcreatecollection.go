package elastic

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
	// ElasticCreateCollectionSettings settings for new collection
	ElasticCreateCollectionSettings struct {
		Name        session.SyncedTemplate `json:"name" displayname:"Collection name" doc-key:"elasticcreatecollection.name"`
		Description string                 `json:"description" displayname:"Collection description" doc-key:"elasticcreatecollection.description"`
		Private     bool                   `json:"private" displayname:"Private collection" doc-key:"elasticcreatecollection.private"`
	}
)

const postCollectionsEndpoint = "api/v1/collections"

// Validate ElasticCreateCollectionSettings action (Implements ActionSettings interface)
func (settings ElasticCreateCollectionSettings) Validate() error {
	if settings.Name.String() == "" {
		return errors.New("No name specified")
	}
	return nil
}

// Execute ElasticCreateCollectionSettings action (Implements ActionSettings interface)
func (settings ElasticCreateCollectionSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	name, err := sessionState.ReplaceSessionVariables(&settings.Name)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	_, err = searchForTag(sessionState, actionState, host, name, 20)
	if err == nil {
		actionState.Failed = true
		sessionState.LogEntry.Log(logger.WarningLevel, "collection already exists")
		return
	}

	restHandler := sessionState.Rest

	createCollection := elasticstructs.CreateCollection{}
	createCollection.Name = name
	createCollection.Description = settings.Description
	if settings.Private {
		createCollection.Type = "private"
	} else {
		createCollection.Type = "public"
	}

	createCollectionJSON, err := jsonit.Marshal(createCollection)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to marshal createCollection for POST"))
	}

	postCreateCollection := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%v/%v", host, postCollectionsEndpoint),
		Content:     createCollectionJSON,
	}

	restHandler.QueueRequest(actionState, true, &postCreateCollection, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return // we had an error
	}
	if postCreateCollection.ResponseStatusCode != http.StatusCreated {
		actionState.AddErrors(errors.New(fmt.Sprintf("Failed to create collection: %s", postCreateCollection.ResponseBody)))
	}

	var createCollectionResponse *elasticstructs.CreateCollectionResponse
	if err := jsonit.Unmarshal(postCreateCollection.ResponseBody, &createCollectionResponse); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to unmarshal CreateCollectionResponse"))
		return
	}

	sessionState.ArtifactMap.FillStreams([]elasticstructs.Collection{{Name: name, ID: createCollectionResponse.ID}})
}
