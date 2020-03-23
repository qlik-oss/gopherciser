package elasticstructs

import "time"

type User struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	TenantID    string    `json:"tenantId"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
	JwtClaims   struct {
		UserID   string   `json:"userId"`
		TenantID string   `json:"tenantId"`
		SubType  string   `json:"subType"`
		Sub      string   `json:"sub"`
		Groups   []string `json:"groups"`
	} `json:"jwtClaims"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}
