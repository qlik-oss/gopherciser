package eventws

type (
	Event struct {
		Operation          string                 `json:"operation,omitempty"`
		Origin             string                 `json:"origin,omitempty"`
		ResourceID         string                 `json:"resourceId,omitempty"`
		ResourceType       string                 `json:"resourceType,omitempty"`
		ResourceSubType    string                 `json:"resourceSubType,omitempty"`
		ResourceSubSubType string                 `json:"resourceSubSubType,omitempty"`
		Success            bool                   `json:"success,omitempty"`
		Time               string                 `json:"time,omitempty"`
		SpaceId            string                 `json:"spaceId,omitempty"`
		ReloadId           string                 `json:"reloadId,omitempty"`
		Data               map[string]interface{} `json:"data,omitempty"`
	}
)

// Constants for known operations
const (
	OperationReloadStarted     = "reload.started"
	OperationReloadEnded       = "reload.ended"
	OperationDataUpdated       = "data.updated"
	OperationAttributesUpdated = "attributes.updated"
	OperationResult            = "result"
	OperationUpdated           = "updated"
	OperationCreated           = "created"
	OperationExecuted          = "executed"
)

// Constants for known ResourceType
const (
	ResourceTypeApp           = "app"
	ResourceTypeReload        = "reload"
	ResourceTypeEvaluation    = "evaluation"
	ResourceTypeItems         = "items"
	ResourceTypeSharingTask   = "sharing-task"
	ResourceTypeReportingTask = "reporting-task"
)
