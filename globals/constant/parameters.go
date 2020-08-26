package constant

import (
	"time"

	"github.com/qlik-oss/gopherciser/helpers"
)

const (
	// MaxStallTime is the threshold time before triggering stall warnings.
	MaxStallTime = 31 * time.Millisecond //windows time api used by golang may have up to 30ms error margin
	// MaxBodySize Maximum byte size for logging request/response body into traffic logs.
	MaxBodySize = 64000
	// ReloadPollInterval Default interval between polls for reload status
	ReloadPollInterval = helpers.TimeDuration(1 * time.Second)

	// ResourceTypeQVapp Resource type for QlikView application, used for app upload and deletion
	ResourceTypeQVapp = "qvapp"
	// ResourceTypeQVapp Resource type for Qlik Sense application, used for app upload and deletion
	ResourceTypeApp = "app"
)
