package elasticstructs

type AppImportResponse struct {
	Attributes struct {
		ID             string      `json:"id"`
		Name           string      `json:"name"`
		Description    string      `json:"description"`
		Thumbnail      string      `json:"thumbnail"`
		LastReloadTime interface{} `json:"lastReloadTime"`
		CreatedDate    string      `json:"createdDate"`
		ModifiedDate   string      `json:"modifiedDate"`
		Owner          string      `json:"owner"`
		DynamicColor   string      `json:"dynamicColor"`
		Published      interface{} `json:"published"`
		PublishTime    string      `json:"publishTime"`
		SpaceID        string      `json:"spaceId"`
		Custom         struct {
		} `json:"custom"`
		Resourcetype string `json:"_resourcetype"`
	} `json:"attributes"`
	Privileges interface{} `json:"privileges"`
	Create     interface{} `json:"create"`
}
