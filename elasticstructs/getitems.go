package elasticstructs

type GetItems struct {
	Data []struct {
		Name                     string      `json:"name"`
		Description              string      `json:"description"`
		ResourceAttributes       interface{} `json:"resourceAttributes"`
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
	Links interface{} `json:"links"`
}
