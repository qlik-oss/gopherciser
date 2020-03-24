package elasticstructs

type (
	AppImportResponse struct {
		Attributes struct {
			ID               string      `json:"id"`
			Name             string      `json:"name"`
			Description      string      `json:"description"`
			Thumbnail        string      `json:"thumbnail"`
			LastReloadTime   string      `json:"lastReloadTime"`
			CreatedDate      string      `json:"createdDate"`
			ModifiedDate     string      `json:"modifiedDate"`
			Owner            string      `json:"owner"`
			OwnerID          string      `json:"ownerId"`
			DynamicColor     string      `json:"dynamicColor"`
			Published        bool        `json:"published"`
			PublishTime      string      `json:"publishTime"`
			SpaceID          string      `json:"spaceId"`
			Custom           interface{} `json:"custom"`
			HasSectionAccess bool        `json:"hasSectionAccess"`
			Encrypted        bool        `json:"encrypted"`
			OriginAppID      string      `json:"originAppId"`
			ResourceType     string      `json:"_resourcetype"`
		} `json:"attributes"`
		Privileges interface{} `json:"privileges"`
		Create     interface{} `json:"create"`
	}
)
