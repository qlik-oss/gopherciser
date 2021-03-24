package structs

import (
	"time"
)

const (
	StreamsTypeStream = "Stream"

	StreamNameEveryone = "Everyone"
	StreamNamePrivate  = "privatestream"
)

type Streams struct {
	Data []struct {
		Attributes struct {
			ModifiedDate time.Time `json:"modifiedDate"`
			Name         string    `json:"name"`
			Privileges   []string  `json:"privileges"`
		} `json:"attributes"`
		ID            string `json:"id"`
		Relationships struct {
			Owner struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
			} `json:"owner"`
		} `json:"relationships"`
		Type string `json:"type"`
	} `json:"data"`
	Included []struct {
		Attributes struct {
			Name          string `json:"name"`
			UserDirectory string `json:"userDirectory"`
			UserID        string `json:"userId"`
		} `json:"attributes"`
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"included"`
}
