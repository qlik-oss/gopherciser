package elasticstructs

type CollectionItems struct {
	Data []struct {
		Name               string `json:"name"`
		Description        string `json:"description"`
		ResourceAttributes struct {
			CreatedDate    interface{} `json:"createdDate"`
			Description    string      `json:"description"`
			DynamicColor   string      `json:"dynamicColor"`
			ID             string      `json:"id"`
			LastReloadTime interface{} `json:"lastReloadTime"`
			ModifiedDate   interface{} `json:"modifiedDate"`
			Name           string      `json:"name"`
			Owner          string      `json:"owner"`
			PublishTime    interface{} `json:"publishTime"`
			Published      bool        `json:"published"`
			Thumbnail      string      `json:"thumbnail"`
		} `json:"resourceAttributes"`
		ResourceCustomAttributes interface{} `json:"resourceCustomAttributes"`
		ResourceUpdatedAt        interface{} `json:"resourceUpdatedAt"`
		ResourceUpdatedBySubject string      `json:"resourceUpdatedBySubject"`
		ResourceType             string      `json:"resourceType"`
		ResourceID               string      `json:"resourceId"`
		ResourceCreatedAt        interface{} `json:"resourceCreatedAt"`
		ResourceCreatedBySubject string      `json:"resourceCreatedBySubject"`
		ID                       string      `json:"id"`
		CreatedAt                interface{} `json:"createdAt"`
		UpdatedAt                interface{} `json:"updatedAt"`
		CreatorID                string      `json:"creatorId"`
		UpdaterID                string      `json:"updaterId"`
		CreatedBySubject         string      `json:"createdBySubject"`
		UpdatedBySubject         string      `json:"updatedBySubject"`
		TenantID                 string      `json:"tenantId"`
		IsFavorited              bool        `json:"isFavorited"`
		Links                    interface{} `json:"links"`
		Actions                  []string    `json:"actions"`
		CollectionIds            []string    `json:"collectionIds"`
	} `json:"data"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Prev struct {
			Href string `json:"href"`
		} `json:"prev"`
		Next struct {
			Href string `json:"href"`
		} `json:"next"`
		Collection struct {
			Href string `json:"href"`
		} `json:"collection"`
	} `json:"links"`
}
