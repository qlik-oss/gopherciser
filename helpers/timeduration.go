package helpers

import (
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type TimeDuration time.Duration

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

// UnmarshalJSON unmarshal time buffer duration from JSON
func (duration *TimeDuration) UnmarshalJSON(arg []byte) error {
	var v interface{}
	if err := jsonit.Unmarshal(arg, &v); err != nil {
		return errors.Wrap(err, "failed to unmarshal time buffer duration")
	}

	switch value := v.(type) {
	case float64:
		*duration = TimeDuration(value)
	case string:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal time buffer duration ")
		}
		*duration = TimeDuration(dur)
	default:
		return errors.Errorf("failed to unmarshal time buffer duration %T<%v>", value, value)
	}

	return nil
}

// MarshalJSON marshal time buffer duration to JSON
func (duration TimeDuration) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%s"`, time.Duration(duration))
	return []byte(s), nil
}
