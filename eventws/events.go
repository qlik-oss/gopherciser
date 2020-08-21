package eventws

type (
	Event struct {
		Operation    string `json:"operation,omitempty"`
		Origin       string `json:"origin,omitempty"`
		ResourceID   string `json:"resourceId,omitempty"`
		ResourceType string `json:"resourceType,omitempty"`
		Success      bool   `json:"success,omitempty"`
		Time         string `json:"time,omitempty"`
	}
)

// Constants for known operations
const (
	OperationReloadEnded = "reload.ended"
)
