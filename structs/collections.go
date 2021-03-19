package structs

import "time"

type (
	Collection struct {
		Name             string    `json:"name"`
		Description      string    `json:"description"`
		Type             string    `json:"type"`
		ID               string    `json:"id"`
		CreatedAt        time.Time `json:"createdAt"`
		UpdatedAt        time.Time `json:"updatedAt"`
		CreatorID        string    `json:"creatorId"`
		UpdaterID        string    `json:"updaterId"`
		CreatedBySubject string    `json:"createdBySubject"`
		UpdatedBySubject string    `json:"updatedBySubject"`
		TenantID         string    `json:"tenantId"`
		Links            struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
			Items struct {
				Href string `json:"href"`
			} `json:"items"`
		} `json:"links"`
		ItemCount int `json:"itemCount"`
	}

	CollectionRequest struct {
		Data  []Collection `json:"data"`
		Links struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
			Next struct {
				Href string `json:"href"`
			} `json:"next"`
		} `json:"links"`
	}
)
