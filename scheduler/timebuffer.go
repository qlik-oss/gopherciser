package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// TimeBufMode type of time buffer
	TimeBufMode int

	// TimeBuffer time buffer at end of a user iteration
	TimeBuffer struct {
		Mode     TimeBufMode          `json:"mode,omitempty" displayname:"Time buffer mode" doc-key:"config.scheduler.iterationtimebuffer.mode"`
		Duration helpers.TimeDuration `json:"duration,omitempty" displayname:"Time buffer duration" doc-key:"config.scheduler.iterationtimebuffer.duration"`

		startTime time.Time
	}
)

const (
	// TimeBufNoWait no time buffer
	TimeBufNoWait TimeBufMode = iota
	// TimeBufConstant constant time buffer
	TimeBufConstant
	// TimeBufOnError time buffer on error
	TimeBufOnError
	// TimeBufMinDur buffer time until minimum duration
	TimeBufMinDur
)

func (mode TimeBufMode) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"nowait":      int(TimeBufNoWait),
		"constant":    int(TimeBufConstant),
		"onerror":     int(TimeBufOnError),
		"minduration": int(TimeBufMinDur),
	})
	return enumMap
}

// SetDurationStart mark start of duration time (to be used with TimeBufMinDur)
func (timeBuf *TimeBuffer) SetDurationStart(time time.Time) {
	timeBuf.startTime = time
}

// Wait inserts time buffer or context
func (timeBuf *TimeBuffer) Wait(ctx context.Context, hasErrors bool) error {
	if timeBuf == nil || timeBuf.Mode == TimeBufNoWait {
		return nil
	}

	duration := time.Duration(timeBuf.Duration)

	if duration < time.Nanosecond {
		return errors.Errorf("No duration defined for mode<%v>", timeBuf.Mode)
	}

	switch timeBuf.Mode {
	case TimeBufConstant:
		helpers.WaitFor(ctx, duration)
	case TimeBufMinDur:
		if timeBuf.startTime.IsZero() || timeBuf.startTime.After(time.Now()) {
			return errors.Errorf("illegal start time<%v>", timeBuf.startTime)
		}
		dur := duration - time.Since(timeBuf.startTime)
		if dur > 0 {
			helpers.WaitFor(ctx, dur)
		}
	case TimeBufOnError:
		if hasErrors {
			helpers.WaitFor(ctx, duration)
		}
	default:
		return errors.Errorf("TimeBuf mode<%v> not supported", timeBuf.Mode)
	}

	return nil
}

// Validate settings of time buffer
func (timeBuf *TimeBuffer) Validate() error {
	if timeBuf == nil || timeBuf.Mode == TimeBufNoWait {
		return nil // not enabled
	}

	if timeBuf.Duration > 0 {
		return nil
	}

	return errors.New("no duration defined for time buffer")
}

// UnmarshalJSON unmarshal time buffer mode from JSON
func (mode *TimeBufMode) UnmarshalJSON(arg []byte) error {
	i, err := mode.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal timeBufModeEnum")
	}

	*mode = TimeBufMode(i)
	return nil
}

// MarshalJSON marshal time buffer mode to JSON
func (mode TimeBufMode) MarshalJSON() ([]byte, error) {
	str, err := mode.GetEnumMap().String(int(mode))
	if err != nil {
		return nil, errors.Errorf("Unknown timeBufModeEnum<%d>", mode)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}
