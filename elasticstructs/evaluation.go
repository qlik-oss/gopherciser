package elasticstructs

type (
	EvaluationProperties struct {
		SpaceID string `json:"spaceId,omitempty"`
		Owner   string `json:"owner,omitempty"`
		Type    string `json:"type,omitempty"`
	}

	EvaluationResource struct {
		ID         string               `json:"id"`
		Type       string               `json:"type"`
		Properties EvaluationProperties `json:"properties"`
	}

	Evaluation struct {
		Resources []EvaluationResource `json:"resources"`
	}
)
