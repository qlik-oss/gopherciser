package elasticstructs

type CreateCollection struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
