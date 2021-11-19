package structs

type Privileges struct {
	Data []struct {
		Attributes struct {
			Definition struct {
				App          []string `json:"app"`
				Features     []string `json:"features"`
				Installation string   `json:"installation"`
				Override     struct {
				} `json:"override"`
				QvVersion string        `json:"qvVersion"`
				Stream    []interface{} `json:"stream"`
				Schemas   []interface{} `json:"schemas"`
				Routes    []interface{} `json:"routes"`
			} `json:"definition"`
		} `json:"attributes"`
		Type string `json:"type"`
	} `json:"data"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}
