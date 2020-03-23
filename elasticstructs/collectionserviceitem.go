package elasticstructs

import "time"

type CollectionServiceItem struct {
	Name               string `json:"name"`
	ResourceID         string `json:"resourceId"`
	ResourceType       string `json:"resourceType"`
	Description        string `json:"description"`
	SpaceID            string `json:"spaceId,omitempty"`
	ResourceAttributes struct {
		ID               string    `json:"id"`
		Name             string    `json:"name"`
		Description      string    `json:"description"`
		Thumbnail        string    `json:"thumbnail"`
		LastReloadTime   time.Time `json:"lastReloadTime"`
		CreatedDate      string    `json:"createdDate"`
		ModifiedDate     string    `json:"modifiedDate"`
		Owner            string    `json:"owner"`
		DynamicColor     string    `json:"dynamicColor"`
		Published        bool      `json:"published"`
		PublishTime      string    `json:"publishTime"`
		HasSectionAccess bool      `json:"hasSectionAccess"`
		Resourcetype     string    `json:"_resourcetype"`
		SpaceID          string    `json:"spaceId,omitempty"`
	} `json:"resourceAttributes"`
	ResourceCustomAttributes interface{} `json:"resourceCustomAttributes"`
	ResourceCreatedAt        time.Time   `json:"resourceCreatedAt"`
	ResourceCreatedBySubject string      `json:"resourceCreatedBySubject"`
}

type ShareAppsEndpointPayload struct {
	Attributes struct {
		Custom struct {
			Groupswithaccess  []string `json:"groupswithaccess"`
			UserIdsWithAccess []string `json:"userIdsWithAccess"`
		} `json:"custom"`
	} `json:"attributes"`
}

type ShareItemsEndpointPayload struct {
	Name                     string `json:"name"`
	ResourceType             string `json:"resourceType"`
	ResourceID               string `json:"resourceId"`
	ResourceCustomAttributes struct {
		Groupswithaccess  []string `json:"groupswithaccess"`
		UserIdsWithAccess []string `json:"userIdsWithAccess"`
	} `json:"resourceCustomAttributes"`
}
