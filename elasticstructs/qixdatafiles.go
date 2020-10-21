package elasticstructs

import "time"

type QixDataFile struct {
	CreatedDate  time.Time `json:"createdDate"`
	ID           string    `json:"id"`
	ModifiedDate time.Time `json:"modifiedDate"`
	Name         string    `json:"name"`
	OwnerID      string    `json:"ownerId"`
	Size         int       `json:"size"`
}
