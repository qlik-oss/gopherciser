package scenario

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	ElasticOpenHubSettings struct{}
)

// UnmarshalJSON unmarshal ElasticOpenHubSettings
func (openHub *ElasticOpenHubSettings) UnmarshalJSON(arg []byte) error {
	deprecatedSettings := struct {
		StreamMode interface{} `json:"streams"`
		StreamList []string    `json:"streamlist"`
	}{}

	if err := json.Unmarshal(arg, &deprecatedSettings); err == nil {
		if deprecatedSettings.StreamMode != nil {
			if mode, ok := deprecatedSettings.StreamMode.(string); !ok || mode != "default" {
				return errors.Errorf("action<%s> no longer supports streams, use %s", ActionElasticOpenHub, ActionElasticExplore)
			}
		}
		if len(deprecatedSettings.StreamList) > 0 {
			return errors.Errorf("action<%s> no longer supports streamlist, use %s", ActionElasticOpenHub, ActionElasticExplore)
		}
	} else {
		return err
	}

	*openHub = ElasticOpenHubSettings{}
	return nil
}

// Execute open Efe hub
func (openHub ElasticOpenHubSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if !sessionState.LoggedIn {
		headers, err := connection.GetHeaders(sessionState)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to connect using ElasticOpenHub"))
			return
		}
		host, err := connection.GetHost()
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "Failed to extract hostname"))
			return
		}
		sessionState.HeaderJar.SetHeader(host, headers)
		sessionState.LoggedIn = true
	}

	// New hub connection, clear any existing apps.
	sessionState.ArtifactMap = session.NewAppMap()

	client, err := session.DefaultClient(connection, sessionState)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	sessionState.Rest.SetClient(client)

	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	getLocale(sessionState, actionState, host)

	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/collections/favorites", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return // request had errors, don't parse response
		}

		var favorites elasticstructs.Favorites
		if err := jsonit.Unmarshal(req.ResponseBody, &favorites); err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to unmarshal favorites"))
			return
		}
		favCollection := elasticstructs.Collection{Name: "Favorites", ID: favorites.ID}
		if favCollection.ID == "" {
			actionState.AddErrors(errors.New("No favorite collection id"))
			return
		}
		sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/collections/%s/items?limit=30&sort=-createdAt", host, favCollection.ID), actionState, sessionState.LogEntry, nil, func(err error, collectionRequest *session.RestRequest) {
			fillAppMapFromItemRequest(sessionState, actionState, collectionRequest, false)
		})
	})

	var spaces []elasticstructs.Space
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/spaces?limit=100", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		spaces = fillArtifactsFromSpaces(sessionState, actionState, req, false) // Will have execute after next sessionState.Wait
	})

	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/features", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/licenses/status", host), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/quotas?reportUsage=true", host), actionState, sessionState.LogEntry, nil)
	userDataReq := sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/users/me", host), actionState, sessionState.LogEntry, nil)

	// some systems has v0, just warn on identity-providers error
	optionsNoError := session.DefaultReqOptions()
	optionsNoError.FailOnError = false
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/identity-providers/me/meta", host), actionState, sessionState.LogEntry, &optionsNoError)

	if sessionState.Wait(actionState) {
		return // we had an error
	}

	var userData elasticstructs.User
	if err := jsonit.Unmarshal(userDataReq.ResponseBody, &userData); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed unmarshaling user data"))
		return
	}
	sessionState.CurrentUser = &userData
	sessionState.LogEntry.LogInfo("TenantUser", strings.Join([]string{userData.Name, userData.Subject, userData.ID, userData.TenantID}, ";"))

	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/tenants/me", host), actionState, sessionState.LogEntry, nil)

	// This get items request is done by client, but resulting apps are never shown to users
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/items?sort=-updatedAt&limit=30", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		fillAppMapFromItemRequest(sessionState, actionState, req, false) // todo should we fill map as these are never shown to user?
	})

	if len(spaces) > 0 {
		// we own or are a part of a at least one space, send evaluation request for all these spaces
		var evaluation elasticstructs.Evaluation
		evaluation.Resources = make([]elasticstructs.EvaluationResource, 0, len(spaces))
		for _, space := range spaces {
			evaluation.Resources = append(evaluation.Resources, elasticstructs.EvaluationResource{
				ID:   space.ID,
				Type: "app",
				Properties: elasticstructs.EvaluationProperties{
					SpaceID: space.ID,
				},
			})
		}

		postData, err := jsonit.Marshal(evaluation)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to marshal evaluation request"))
		}

		sessionState.Rest.PostAsync(fmt.Sprintf("%s/api/v1/policies/evaluation", host), actionState, sessionState.LogEntry, postData, nil)
	}

	if sessionState.Wait(actionState) {
		return // we had an error
	}

	// send evaluation request
	evaluation := elasticstructs.Evaluation{
		Resources: []elasticstructs.EvaluationResource{
			{
				ID:   userData.ID,
				Type: "app",
				Properties: elasticstructs.EvaluationProperties{
					Owner: userData.Subject,
				},
			},
			// todo these 3 should also be evaluated, unfortunately the first existence of the ID's are in a javascript for the client and thus not reachable for us and can't be tested.
			//{
			//	ID:   "",
			//	Type: "collection",
			//	Properties: elasticstructs.EvaluationProperties{
			//		Type: "public",
			//	},
			//},
			//{
			//	ID:   "",
			//	Type: "space",
			//	Properties: elasticstructs.EvaluationProperties{
			//		Type: "shared",
			//	},
			//},
			//{
			//	ID:   "",
			//	Type: "space",
			//	Properties: elasticstructs.EvaluationProperties{
			//		Type: "managed",
			//	},
			//},
		},
	}

	postData, err := jsonit.Marshal(evaluation)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to marshal evaluation request"))
		return
	}

	sessionState.Rest.PostAsync(fmt.Sprintf("%s/api/v1/policies/evaluation", host), actionState, sessionState.LogEntry, postData, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/qlik-groups?tenantId=%s&limit=0&fields=displayName", host, userData.TenantID), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/users?tenantId=%s&limit=0&fields=name,picture", host, userData.TenantID), actionState, sessionState.LogEntry, nil)
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/items?sort=-createdAt&limit=30&ownerId=%s", host, userData.ID), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		fillAppMapFromItemRequest(sessionState, actionState, req, false)
	})

	if sessionState.Wait(actionState) {
		return // we had an error
	}

	// Debug log of artifact map in it's entirety
	if err := sessionState.ArtifactMap.LogMap(sessionState.LogEntry); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, err)
	}
}

// get locale is semi-async, the first request is synchronous and sub-sequent request/-s async.
func getLocale(sessionState *session.State, actionState *action.State, host string) {
	var firstReq sync.WaitGroup
	firstReq.Add(1)
	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/locale", host), actionState, sessionState.LogEntry, &session.ReqOptions{FailOnError: false}, func(err error, req *session.RestRequest) {
		firstReq.Done()

		if err == nil {
			err = session.CheckResponseStatus(req, []int{200})
		}

		// Retry first request once in case of error
		if err != nil {
			sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/locale", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
				if err != nil {
					return
				}
				sendTranslationRequest(sessionState, actionState, req.ResponseBody, host)
			})
			return
		}

		sendTranslationRequest(sessionState, actionState, req.ResponseBody, host)
	})

	// make sure first /locale request has received a response before sending subsequent requests
	firstReq.Wait()
}

func sendTranslationRequest(sessionState *session.State, actionState *action.State, localeRaw json.RawMessage, host string) {
	var locale elasticstructs.Locale
	if err := jsonit.Unmarshal(localeRaw, &locale); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to unmarshal locale"))
		return
	}
	// Get the translations for current locale, i.e en.json for English
	sessionState.Rest.GetAsync(fmt.Sprintf("%s/%s.json", host, locale.Locale), actionState, sessionState.LogEntry, nil)
}

func fillAppMapFromItemRequest(sessionState *session.State, actionState *action.State, itemRequest *session.RestRequest, doPage bool) {
	//Make to wait for requests done here in pending handler
	sessionState.Pending.IncPending()
	defer sessionState.Pending.DecPending()

	if err := session.CheckResponseStatus(itemRequest, []int{200}); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to get collection <%s>", string(itemRequest.ResponseBody)))
		return
	}
	collectionItemsRaw := itemRequest.ResponseBody
	var collectionItems elasticstructs.CollectionItems
	if err := jsonit.Unmarshal(collectionItemsRaw, &collectionItems); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed unmarshaling collection items in <%s>", itemRequest.ResponseBody))
		return
	}
	err := fillAppMapFromCollection(sessionState.ArtifactMap, &collectionItems)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to add apps to ArtifactMap"))
		return
	}

	if doPage && collectionItems.Links.Next.Href != "" {
		sessionState.Rest.GetAsyncWithCallback(collectionItems.Links.Next.Href, actionState, sessionState.LogEntry, nil, func(e error, req *session.RestRequest) {
			fillAppMapFromItemRequest(sessionState, actionState, req, true)
		})
	}
}

func fillAppMapFromCollection(appMap *session.ArtifactMap, items *elasticstructs.CollectionItems) error {
	appsResp := make([]session.AppsResp, 0, len(items.Data))
	for _, item := range items.Data {
		appsResp = append(appsResp, session.AppsResp{Title: item.Name, Name: item.Name, ID: item.ResourceID, ItemID: item.ID})
	}
	err := appMap.FillAppsUsingName(&session.AppData{Data: appsResp})
	if err != nil {
		return err
	}
	return nil
}

// fillArtifactsFromSpaces pages synchronously if doPage = true and returns a list of all spaces
func fillArtifactsFromSpaces(sessionState *session.State, actionState *action.State, spacesReq *session.RestRequest, doPage bool) []elasticstructs.Space {
	if err := session.CheckResponseStatus(spacesReq, []int{200}); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to get spaces <%s>", string(spacesReq.ResponseBody)))
		return nil
	}

	var spaces elasticstructs.Spaces
	if err := jsonit.Unmarshal(spacesReq.ResponseBody, &spaces); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed unmarshaling spaces"))
		return nil
	}

	sessionState.ArtifactMap.FillSpaces(spaces.Data)
	allSpaces := spaces.Data
	if doPage && spaces.Links.Next.Href != "" {
		if _, err := sessionState.Rest.GetSyncWithCallback(spaces.Links.Next.Href, actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
			newSpaces := fillArtifactsFromSpaces(sessionState, actionState, req, true)
			if len(newSpaces) > 0 {
				allSpaces = append(allSpaces, newSpaces...)
			}
		}); err != nil {
			return nil // error already reported on actionState
		}
	}

	return allSpaces
}

func fillArtifactsFromSpace(sessionState *session.State, actionState *action.State, spaceReq *session.RestRequest) *elasticstructs.Space {
	if err := session.CheckResponseStatus(spaceReq, []int{200}); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to get space <%s>", string(spaceReq.ResponseBody)))
		return nil
	}
	var space elasticstructs.Space
	if err := jsonit.Unmarshal(spaceReq.ResponseBody, &space); err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed unmarshaling space"))
		return nil
	}
	sessionState.ArtifactMap.AddSpace(space)
	return &space
}

// Validate open Efe hub settings
func (openHub ElasticOpenHubSettings) Validate() (err error) {
	return nil
}
