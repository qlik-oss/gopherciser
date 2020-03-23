package elasticstructs

type PostCopyApp struct {
	Attributes struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		SpaceID     string `json:"spaceId"`
	} `json:"attributes"`
}
