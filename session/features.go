package session

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	// Features maps feature flags and values
	Features struct {
		m  map[string]bool
		mu sync.Mutex
	}

	// FeatureAllocationError returned when feature map is expected to be allocated but is not
	FeatureAllocationError struct{}

	// FeatureFlagNotFoundError returned when feature flag was not found
	FeatureFlagNotFoundError string
)

func (err FeatureAllocationError) Error() string {
	return "features map not allocated"
}

func (err FeatureFlagNotFoundError) Error() string {
	return fmt.Sprintf("feature %s does not exist in feature map", string(err))
}

// UpdateFeatureMap request features from server and updates feature map
func (features *Features) UpdateFeatureMap(rest *RestHandler, host string, actionState *action.State, logEntry *logger.LogEntry) {
	features.mu.Lock()
	defer features.mu.Unlock()

	req, err := rest.GetSync(fmt.Sprintf("%s/api/v1/features", host), actionState, logEntry, nil)
	if err != nil {
		return
	}

	if features.m == nil {
		features.m = make(map[string]bool)
	}

	if err := jsonit.Unmarshal(req.ResponseBody, &features.m); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
}

// IsFeatureEnabled check if feature flag is enabled
func (features *Features) IsFeatureEnabled(feature string) (bool, error) {
	features.mu.Lock()
	defer features.mu.Unlock()

	if features.m == nil {
		return false, FeatureAllocationError{}
	}

	enabled, exists := features.m[feature]
	if !exists {
		return false, FeatureFlagNotFoundError(feature)
	}

	return enabled, nil
}

// IsFeatureEnabledDefault check if feature flag is enabled, if not existing return default value
func (features *Features) IsFeatureEnabledDefault(feature string, d bool) bool {
	enabled, err := features.IsFeatureEnabled(feature)
	if err != nil {
		return d
	}
	return enabled
}
