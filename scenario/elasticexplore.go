package scenario

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	SortingMode int
	OwnerMode   int

	// ElasticExploreSettings fills artifact map with apps to be used in simulation
	ElasticExploreSettings struct {
		// KeepCurrent keep current artifact map and fill with results from explore action
		KeepCurrent bool `json:"keepcurrent" displayname:"Keep current" doc-key:"elasticexplore.keepcurrent"`
		// DoPaging page through all items
		DoPaging bool `json:"paging" displayname:"Do paging" doc-key:"elasticexplore.paging"`
		// Sorting SortingMode of items, defaults to SortingModeDefault
		Sorting SortingMode `json:"sorting" displayname:"Sorting mode" doc-key:"elasticexplore.sorting"`
		// Owner OwnerMode, defaults to OwnerModeAll
		Owner OwnerMode `json:"owner" displayname:"Owner mode" doc-key:"elasticexplore.owner"`
		// SpaceName only get items which members of this space, cannot be used together with SpaceID
		SpaceName session.SyncedTemplate `json:"space" displayname:"Space name" doc-key:"elasticexplore.space"`
		// SpaceId only get items which members of this space, cannot be used together with SpaceName
		SpaceId string `json:"spaceid" displayname:"Space ID" doc-key:"elasticexplore.spaceid"`
		// CollectionIds filter on these tag (collection) ids
		CollectionIds []string `json:"tagids" displayname:"Tag IDs" doc-key:"elasticexplore.tagids"`
		// CollectionNames filter on these tag (collection) names
		CollectionNames []string `json:"tags" displayname:"Tag names" doc-key:"elasticexplore.tags"`
	}
)

// SortingMode
const (
	// SortingModeDefault default sorting mode, should be changed if client changes its default
	SortingModeDefault   SortingMode = iota //-createdAt
	SortingModeUpdatedAt                    //-updatedAt
	SortingModeCreatedAt                    //-createdAt
	SortingModeName                         //-name
)

// OwnerMode
const (
	OwnerModeAll OwnerMode = iota
	OwnerModeMe
	OwnerModeOthers
)

var (
	sortingEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"default": int(SortingModeDefault),
		"created": int(SortingModeCreatedAt),
		"updated": int(SortingModeUpdatedAt),
		"name":    int(SortingModeName),
	})

	ownerEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"all":    int(OwnerModeAll),
		"me":     int(OwnerModeMe),
		"others": int(OwnerModeOthers),
	})
)

// GetEnumMap for sorting mode
func (sorting SortingMode) GetEnumMap() *enummap.EnumMap {
	return sortingEnum
}

// UnmarshalJSON unmarshal SortingMode
func (sorting *SortingMode) UnmarshalJSON(arg []byte) error {
	i, err := sortingEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal SortingMode")
	}

	*sorting = SortingMode(i)
	return nil
}

// MarshalJSON marshal SortingMode
func (sorting SortingMode) MarshalJSON() ([]byte, error) {
	str, err := sortingEnum.String(int(sorting))
	if err != nil {
		return nil, errors.Errorf("unknown SortingMode<%d>", sorting)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Value of SortingMode for URL parameters
func (sorting SortingMode) Value() string {
	switch sorting {
	case SortingModeCreatedAt:
		return "-createdAt"
	case SortingModeUpdatedAt:
		return "-updatedAt"
	case SortingModeName:
		return "%2Bname"
	default:
		return "-createdAt"
	}
}

// GetEnumMap for owner mode
func (owner OwnerMode) GetEnumMap() *enummap.EnumMap {
	return ownerEnum
}

// UnmarshalJSON unmarshal OwnerMode
func (owner *OwnerMode) UnmarshalJSON(arg []byte) error {
	i, err := ownerEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal OwnerMode")
	}

	*owner = OwnerMode(i)
	return nil
}

// MarshalJSON marshal OwnerMode
func (owner OwnerMode) MarshalJSON() ([]byte, error) {
	str, err := ownerEnum.String(int(owner))
	if err != nil {
		return nil, errors.Errorf("unknown OwnerMode<%d>", owner)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil

}

// Validate ElasticExplore action
func (settings ElasticExploreSettings) Validate() error {
	if settings.SpaceId != "" && settings.SpaceName.String() != "" {
		return errors.Errorf("Both spaceid<%s> and space<%s> defined", settings.SpaceId, &settings.SpaceName)
	}
	return nil
}

// Execute  ElasticExplore action
func (settings ElasticExploreSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if !settings.KeepCurrent {
		sessionState.ArtifactMap.ClearArtifactMap()
	}

	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	urlParams := make(helpers.Params)

	// Apply space filter
	var space *elasticstructs.Space
	spaceName, err := sessionState.ReplaceSessionVariables(&settings.SpaceName)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
	if spaceName != "" {
		if strings.ToLower(spaceName) == "personal" {
			urlParams["spaceId"] = "personal"
		} else {
			var err error
			space, err = SearchForSpaceByName(sessionState, actionState, host, spaceName)
			if err != nil {
				actionState.AddErrors(err)
				return
			}
			urlParams["spaceId"] = space.ID
		}
	}
	if settings.SpaceId != "" {
		urlParams["spaceId"] = settings.SpaceId
		space, err = searchForSpaceByID(sessionState, actionState, host, settings.SpaceId)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
	}
	if space != nil && space.Links.Assignments.Href != "" {
		sessionState.Rest.GetAsync(space.Links.Assignments.Href+"?limit=100", actionState, sessionState.LogEntry, nil)
	}

	collectionCount := len(settings.CollectionNames) + len(settings.CollectionIds)

	// Was tags dropdown "clicked"?
	if collectionCount > 0 {
		collectionIds := make([]string, 0, collectionCount)
		sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/collections?type=public&sort=-name&limit=100", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
			var collectionData elasticstructs.CollectionRequest
			if err := jsonit.Unmarshal(req.ResponseBody, &collectionData); err != nil {
				actionState.AddErrors(errors.Wrap(err, "failed unmarshaling collection data"))
				return
			}
			sessionState.ArtifactMap.FillStreams(collectionData.Data)
		})
		if sessionState.Wait(actionState) {
			return // We had an error
		}

		// Do we have any tag names
		for _, collection := range settings.CollectionNames {
			id, err := searchForTag(sessionState, actionState, host, collection, 100)
			if err != nil {
				actionState.AddErrors(errors.WithStack(err))
				return
			}
			collectionIds = append(collectionIds, id)
		}
		if len(settings.CollectionIds) > 0 { // Todo check for existence?
			collectionIds = append(collectionIds, settings.CollectionIds...)
		}
		urlParams["collectionId"] = strings.Join(collectionIds, ",")
	}

	// Make sure we have session CurrentUser data
	if sessionState.CurrentUser == nil {
		userDataReq := sessionState.Rest.GetAsync(fmt.Sprintf("%s/api/v1/users/me", host), actionState, sessionState.LogEntry, nil)
		if sessionState.Wait(actionState) {
			return
		}
		var userData elasticstructs.User
		if err := jsonit.Unmarshal(userDataReq.ResponseBody, &userData); err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed unmarshaling user data"))
			return
		}
		sessionState.CurrentUser = &userData
	}

	// apply filter on owner
	switch settings.Owner {
	case OwnerModeMe:
		if !isSessionUserOk(sessionState, actionState) {
			return
		}
		urlParams["ownerId"] = sessionState.CurrentUser.ID
	case OwnerModeOthers:
		if !isSessionUserOk(sessionState, actionState) {
			return
		}
		urlParams["notOwnerId"] = sessionState.CurrentUser.ID
	}

	// apply per page limit
	urlParams["limit"] = "24"

	// apply sorting
	urlParams["sort"] = settings.Sorting.Value()

	// apply resourceType
	urlParams["resourceType"] = "app,qvapp,qlikview,genericlink,sharingservicetask"

	sessionState.Rest.GetAsyncWithCallback(fmt.Sprintf("%s/api/v1/items%s", host, urlParams), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		fillAppMapFromItemRequest(sessionState, actionState, req, settings.DoPaging)
	})
	sessionState.Wait(actionState)

	// Debug log of artifact map in it's entirety
	if err := sessionState.ArtifactMap.LogMap(sessionState.LogEntry); err != nil {
		sessionState.LogEntry.Log(logger.WarningLevel, err)
	}
}

// AppStructureAction implements AppStructureAction interface
func (settings ElasticExploreSettings) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     true,
	}, nil
}

func isSessionUserOk(sessionState *session.State, actionState *action.State) bool {
	if sessionState.CurrentUser == nil || sessionState.CurrentUser.ID == "" {
		actionState.AddErrors(errors.New("No current user to be used for owner filter"))
		return false
	}
	return true
}

// SearchForSpaceByName looks for space in artifact map or tries to request from server, returns space ID
func SearchForSpaceByName(sessionState *session.State, actionState *action.State, host, spaceName string) (*elasticstructs.Space, error) {
	space, err := sessionState.ArtifactMap.GetSpaceByName(spaceName)
	if err == nil {
		return space, nil
	}
	switch err.(type) {
	case session.SpaceNameNotFoundError:
		// spaces/filter seems to no be implemented yet, so we have to iterate everything instead, replace with this once api is usable
		//filter := elasticstructs.Filter{
		//	Names: []string{spaceName},
		//}
		//
		//content, err := jsonit.Marshal(filter)
		//if err != nil {
		//	return nil, errors.Wrap(err, "failed to marshal spaces filter request")
		//}
		//var syncReq sync.WaitGroup
		//syncReq.Add(1)
		//sessionState.Rest.PostAsyncWithCallback(fmt.Sprintf("%s/spaces/filter", host), actionState, sessionState.LogEntry, content, nil, func(err error, req *session.RestRequest) {
		//	defer syncReq.Done()
		//	fillArtifactsFromSpaces(sessionState, actionState, req, false)
		//})
		//syncReq.Wait()
		//space, err := sessionState.ArtifactMap.GetSpaceID(spaceName)
		//if err != nil {
		//	return nil, errors.WithStack(err)
		//}
		//return space, nil

		// ugly code to iterate search for space, remove this once spaces/filter exists
		spaceReq, err := sessionState.Rest.GetSyncWithCallback(fmt.Sprintf("%s/api/v1/spaces?limit=100", host), actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return searchIterateSpace(sessionState, actionState, spaceReq, spaceName)
	default:
		return nil, errors.WithStack(err)
	}
}

func searchIterateSpace(sessionState *session.State, actionState *action.State, req *session.RestRequest, spaceName string) (*elasticstructs.Space, error) {
	spaceList := fillArtifactsFromSpaces(sessionState, actionState, req, false)
	for _, space := range spaceList {
		if space.Name == spaceName {
			return &space, nil
		}
	}
	var spaces elasticstructs.Spaces
	if err := jsonit.Unmarshal(req.ResponseBody, &spaces); err != nil {
		return nil, errors.Wrap(err, "failed unmarshaling spaces")
	}
	if spaces.Links.Next.Href != "" {
		spaceReq, err := sessionState.Rest.GetSync(spaces.Links.Next.Href, actionState, sessionState.LogEntry, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return searchIterateSpace(sessionState, actionState, spaceReq, spaceName)
	}
	return nil, errors.Errorf("space<%s> not found", spaceName)
}

func searchForSpaceByID(sessionState *session.State, actionState *action.State, host, spaceID string) (*elasticstructs.Space, error) {
	space, err := sessionState.ArtifactMap.GetSpaceByID(spaceID)
	if err == nil {
		return space, nil
	}
	switch err.(type) {
	case session.SpaceIDNotFoundError:
		spaceReq, err := sessionState.Rest.GetSync(fmt.Sprintf("%s/api/v1/spaces/%s", host, spaceID), actionState, sessionState.LogEntry, nil)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		space := fillArtifactsFromSpace(sessionState, actionState, spaceReq)
		if space == nil {
			return nil, errors.Errorf("failed to find space with ID %s", spaceID)
		}
		return space, nil
	default:
		return nil, errors.WithStack(err)
	}
}

func searchForTag(sessionState *session.State, actionState *action.State, host, collection string, limit int) (string, error) {
	id, err := sessionState.ArtifactMap.GetStreamID(collection)
	if err == nil {
		return id, nil
	}

	query := fmt.Sprintf("%s/api/v1/collections?query=%s&limit=%d&sort=-name&type=public", host, collection, limit)
	for {
		var wg sync.WaitGroup
		wg.Add(1)
		// tag not found in first page of tags or previous searches, search for it
		sessionState.Rest.GetAsyncWithCallback(query, actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
			defer wg.Done()
			if err != nil {
				return // We had an error
			}
			var collectionData elasticstructs.CollectionRequest
			if err := jsonit.Unmarshal(req.ResponseBody, &collectionData); err != nil {
				actionState.AddErrors(errors.Wrap(err, "failed unmarshaling collection data"))
				return
			}
			sessionState.ArtifactMap.FillStreams(collectionData.Data)
			query = collectionData.Links.Next.Href
		})
		wg.Wait()

		id, err = sessionState.ArtifactMap.GetStreamID(collection)
		if err == nil {
			// tag was found
			return id, nil
		}
		if query == "" {
			// no more pages
			return "", errors.Errorf("tag<%s> not found", collection)
		}
	}
}
