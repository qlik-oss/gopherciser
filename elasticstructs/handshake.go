package elasticstructs

type Locale struct {
	Locale string `json:"locale"`
}

type Favorites struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	CreatedAt interface{} `json:"createdAt"`
	UpdatedAt interface{} `json:"updatedAt"`
	CreatorID string      `json:"creatorId"`
	UpdaterID string      `json:"updaterId"`
	TenantID  string      `json:"tenantId"`
	Links     struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Items struct {
			Href string `json:"href"`
		} `json:"items"`
	} `json:"links"`
	ItemCount int `json:"itemCount"`
}
