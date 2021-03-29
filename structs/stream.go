package structs

import (
	"time"
)

const (
	StreamTypeApp = "App"
)

type Stream struct {
	Data []struct {
		Attributes struct {
			AvailabilityStatus int       `json:"availabilityStatus"`
			CreatedDate        time.Time `json:"createdDate"`
			Description        string    `json:"description"`
			DynamicColor       string    `json:"dynamicColor"`
			Features           []string  `json:"features"`
			FileSize           int       `json:"fileSize"`
			LastReloadTime     time.Time `json:"lastReloadTime"`
			Name               string    `json:"name"`
			Privileges         []string  `json:"privileges"`
			PublishTime        time.Time `json:"publishTime"`
			Published          bool      `json:"published"`
			SourceAppID        string    `json:"sourceAppId"`
			TargetAppID        string    `json:"targetAppId"`
			Thumbnail          string    `json:"thumbnail"`
		} `json:"attributes"`
		ID            string `json:"id"`
		Relationships struct {
			Owner  StreamRelationships `json:"owner"`
			Stream StreamRelationships `json:"stream"`
		} `json:"relationships"`
		Type string `json:"type"`
	} `json:"data"`
	Included []struct {
		Attributes struct {
			Name          string  `json:"name"`
			UserDirectory *string `json:"userDirectory,omitempty"`
			UserID        *string `json:"userId,omitempty"`
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
