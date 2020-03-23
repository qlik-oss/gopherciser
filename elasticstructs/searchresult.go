package elasticstructs

type SearchResult struct {
	Data  []interface{} `json:"data"`
	Links interface{}   `json:"links"`
}
