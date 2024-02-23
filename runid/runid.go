package runid

import "github.com/google/uuid"

type RunID string

var runID = uuid.NewString()

// Get returns an id unique to an execution of gopherciser
func Get() string {
	return runID
}
