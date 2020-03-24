package elasticstructs

import "time"

type (
	CollectionServiceResourceAttributes struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Description    string `json:"description"`
		Thumbnail      string `json:"thumbnail"`
		LastReloadTime string `json:"lastReloadTime"`
		CreatedDate    string `json:"createdDate"`
		ModifiedDate   string `json:"modifiedDate"`
		// Deprecated: use ownerID
		Owner            string `json:"owner"`
		OwnerID          string `json:"ownerId"`
		DynamicColor     string `json:"dynamicColor"`
		Published        bool   `json:"published"`
		PublishTime      string `json:"publishTime"`
		HasSectionAccess bool   `json:"hasSectionAccess"`
		Encrypted        bool   `json:"encrypted"`
		OriginAppID      string `json:"originAppId"`
		SpaceID          string `json:"spaceId,omitempty"`
		ResourceType     string `json:"_resourcetype"`
	}

	CollectionServiceItem struct {
		Name                     string                              `json:"name"`
		ResourceID               string                              `json:"resourceId"`
		ResourceType             string                              `json:"resourceType"`
		Description              string                              `json:"description"`
		ResourceAttributes       CollectionServiceResourceAttributes `json:"resourceAttributes"`
		ResourceCustomAttributes interface{}                         `json:"resourceCustomAttributes"`
		ResourceCreatedAt        time.Time                           `json:"resourceCreatedAt"`
		ResourceCreatedBySubject string                              `json:"resourceCreatedBySubject"`
		SpaceID                  string                              `json:"spaceId,omitempty"`
	}

	ShareAppsEndpointPayload struct {
		Attributes struct {
			Custom struct {
				Groupswithaccess  []string `json:"groupswithaccess"`
				UserIdsWithAccess []string `json:"userIdsWithAccess"`
			} `json:"custom"`
		} `json:"attributes"`
	}

	ShareItemsEndpointPayload struct {
		Name                     string `json:"name"`
		ResourceType             string `json:"resourceType"`
		ResourceID               string `json:"resourceId"`
		ResourceCustomAttributes struct {
			Groupswithaccess  []string `json:"groupswithaccess"`
			UserIdsWithAccess []string `json:"userIdsWithAccess"`
		} `json:"resourceCustomAttributes"`
	}
)
