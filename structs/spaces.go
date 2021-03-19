package structs

type (
	Href struct {
		Href string `json:"href"`
	}

	// Space QSEoK space definition
	Space struct {
		CreatedAt   string `json:"createdAt"`
		CreatedBy   string `json:"createdBy"`
		Description string `json:"description"`
		ID          string `json:"id"`
		Links       struct {
			Assignments Href `json:"assignments"`
			Self        Href `json:"self"`
		} `json:"links"`
		Meta struct {
			Actions         []string      `json:"actions"`
			AssignableRoles []string      `json:"assignableRoles"`
			Roles           []interface{} `json:"roles"`
		} `json:"meta"`
		Name      string `json:"name"`
		OwnerID   string `json:"ownerId"`
		TenantID  string `json:"tenantId"`
		Type      string `json:"type"`
		UpdatedAt string `json:"updatedAt"`
	}

	// Spaces response of spaces request
	Spaces struct {
		Data  []Space `json:"data"`
		Links struct {
			Self Href `json:"self"`
			Next Href `json:"next"`
			Prev Href `json:"prev"`
		} `json:"links"`
		Meta struct {
			Count int `json:"count"`
		} `json:"meta"`
	}

	Filter struct {
		Ids   []string `json:"ids,omitempty"`
		Names []string `json:"names,omitempty"`
	}

	SpaceReference struct {
		SpaceID string `json:"spaceId"`
	}
)
