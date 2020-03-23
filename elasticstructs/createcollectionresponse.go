package elasticstructs

type CreateCollectionResponse struct {
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Type             string      `json:"type"`
	ID               string      `json:"id"`
	CreatedAt        interface{} `json:"createdAt"`
	UpdatedAt        interface{} `json:"updatedAt"`
	CreatorID        string      `json:"creatorId"`
	UpdaterID        string      `json:"updaterId"`
	CreatedBySubject string      `json:"createdBySubject"`
	UpdatedBySubject string      `json:"updatedBySubject"`
	TenantID         string      `json:"tenantId"`
	Links            interface{} `json:"links"`
	ItemCount        int         `json:"itemCount"`
}
