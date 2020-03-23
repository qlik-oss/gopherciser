package elasticstructs

import "time"

type GetDataFolders []struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	AppID        string    `json:"appId"`
	Size         int       `json:"size"`
	CreatedDate  time.Time `json:"createdDate"`
	ModifiedDate time.Time `json:"modifiedDate"`
}
