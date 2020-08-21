package elasticstructs

type ReloadResponse struct {
	ID           string `json:"id"`
	AppID        string `json:"appId"`
	TenantID     string `json:"tenantId"`
	UserID       string `json:"userId"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Log          string `json:"log"`
	Duration     string `json:"duration"`
	CreationTime string `json:"creationTime"`
	StartTime    string `json:"startTime"`
	EndTime      string `json:"endTime"`
	EngineTime   string `json:"engineTime"`
}
