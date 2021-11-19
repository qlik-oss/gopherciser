package structs

import (
	"time"
)

const (
	StreamTypeApp = "App"
)

type Stream struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			FileSize           int           `json:"fileSize"`
			SourceAppID        string        `json:"sourceAppId"`
			TargetAppID        string        `json:"targetAppId"`
			Name               string        `json:"name"`
			Privileges         []string      `json:"privileges"`
			LastReloadTime     time.Time     `json:"lastReloadTime"`
			CreatedDate        time.Time     `json:"createdDate"`
			PublishTime        time.Time     `json:"publishTime"`
			Published          bool          `json:"published"`
			Description        string        `json:"description"`
			DynamicColor       string        `json:"dynamicColor"`
			Thumbnail          string        `json:"thumbnail"`
			AvailabilityStatus int           `json:"availabilityStatus"`
			Features           []string      `json:"features"`
			CustomProperties   []interface{} `json:"customProperties"`
		} `json:"attributes"`
		Relationships struct {
			Owner  StreamRelationships `json:"owner"`
			Stream StreamRelationships `json:"stream"`
		} `json:"relationships"`
	} `json:"data"`
	Included []struct {
		Attributes struct {
			Name          string      `json:"name"`
			UserID        *string     `json:"userId,omitempty"`
			UserDirectory *string     `json:"userDirectory,omitempty"`
			Privileges    interface{} `json:"privileges,omitempty"`
		} `json:"attributes"`
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"included"`
}
type StreamRelationshipsData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
type StreamRelationships struct {
	Data StreamRelationshipsData `json:"data"`
}
